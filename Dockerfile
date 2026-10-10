FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 隧道管理服务（独立 module，仅标准库，无第三方依赖）。
# 供飞牛封装版在镜像内管理 cloudflared 快速/命名隧道，
# 隧道方案参考 omniroute (github.com/diegosouzapw/OmniRoute) 的 Cloudflare Tunnel 做法。
COPY tunnel/go.mod /src/tunnel/go.mod
COPY tunnel/tunnel-admin.go /src/tunnel/tunnel-admin.go

ARG VERSION=dev
RUN CGO_ENABLED=0 go build -tags nogui -trimpath \
      -ldflags "-s -w -X main.version=${VERSION}" -o /out/magpie . \
    && cd /src/tunnel && CGO_ENABLED=0 go build -trimpath -o /out/tunnel-admin . \
    && mkdir -p /config/home /config/cache /config/data /config/state

# cloudflared：Cloudflare Tunnel 客户端（快速隧道 trycloudflare.com + 命名隧道 --token）
FROM debian:12-slim AS cloudflared
ARG TARGETARCH
RUN apt-get update && apt-get install -y --no-install-recommends curl ca-certificates \
    && case "${TARGETARCH}" in arm64) CFARCH=arm64;; *) CFARCH=amd64;; esac \
    && curl -fsSL "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-${CFARCH}" -o /out/cloudflared \
    && chmod +x /out/cloudflared \
    && mkdir -p /out/certs && cp -a /etc/ssl/certs/. /out/certs/

# A shell for the container's terminal: NAS panels (Synology, 1Panel,
# Portainer) open /bin/bash or /bin/sh, and distroless has neither, so a
# terminal there failed with "stat /bin/bash: no such file or directory".
# Debian 12's bash (with the libtinfo it links), which runs on the glibc of
# distroless's own Debian 12, and busybox for ls, cat and the like.
FROM debian:12-slim AS shell
RUN mkdir -p /out/bin /out/lib \
    && cp /bin/bash /out/bin/bash \
    && cp -L "$(ldd /bin/bash | awk '/libtinfo/ {print $3}')" /out/lib/

FROM busybox:1.37-uclibc AS busybox

# cc, not static: plugins run on Bun, which magpie downloads on first use
# and which needs glibc (static has no libc at all, and Bun's musl build
# would need musl). cc is static plus glibc, libgcc and libstdc++.
FROM gcr.io/distroless/cc-debian12:nonroot

COPY --from=build /out/magpie /magpie
COPY --from=build /out/tunnel-admin /usr/local/bin/tunnel-admin
COPY --from=cloudflared /out/cloudflared /usr/local/bin/cloudflared
COPY --from=cloudflared /out/certs /etc/ssl/certs
COPY --from=busybox /bin/busybox /bin/busybox
COPY --from=shell /out/bin/bash /bin/bash
COPY --from=shell /out/lib/ /usr/lib/
# busybox's commands, magpie and cloudflared on PATH, plus the entry wrapper
# that starts tunnel-admin alongside magpie (tunnel-admin listens on :3431)
USER root
RUN ["/bin/busybox", "sh", "-c", "/bin/busybox --install -s /bin && mkdir -p /usr/local/bin && ln -s /magpie /usr/local/bin/magpie && printf '#!/bin/bash\\n/usr/local/bin/tunnel-admin &\\nexec /magpie \"$@\"\\n' > /usr/local/bin/magpie-entry.sh && chmod +x /usr/local/bin/magpie-entry.sh"]
USER nonroot
COPY --from=build --chown=65532:65532 /config /config

# Everything magpie and the sign-ins it manages write goes into the volume:
# its own files (/config/magpie), sign-ins kept where an agent keeps them
# (~/.codex, ~/.claude…, under HOME), and the cache holding Bun and the
# model catalog. magpie makes the folders a volume from an older image
# lacks.
ENV HOME=/config/home \
    XDG_CONFIG_HOME=/config \
    XDG_CACHE_HOME=/config/cache \
    XDG_DATA_HOME=/config/data \
    XDG_STATE_HOME=/config/state \
    MAGPIE_ADDR=0.0.0.0:3425

VOLUME /config

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/magpie", "healthcheck"]

EXPOSE 3425 3430 3431

# 默认入口同时启动隧道管理服务（:3431）与 magpie（默认 serve）
ENTRYPOINT ["/bin/bash", "/usr/local/bin/magpie-entry.sh"]
CMD ["serve"]
