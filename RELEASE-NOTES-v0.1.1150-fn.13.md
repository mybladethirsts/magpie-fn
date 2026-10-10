# magpie for fnOS v0.1.1150-fn.13

Release Notes / 发布说明

## 概述 / Overview

将开源 AI 模型路由网关 [yetone/magpie](https://github.com/yetone/magpie) 封装为飞牛 fnOS 安装包（.fpk）。定位：**集中 AI 网关**——多台电脑共用一套模型 / 订阅配置，NAS 7×24 提供服务，各电脑 Agent 直连即可获得模型能力；内置云隧道，让内网之外的设备也能接入。

This release packages yetone/magpie as a fnOS application (.fpk). Positioning: **a centralized AI gateway** — one model / subscription configuration shared by all your computers, served 24×7 by the NAS, with a built-in Cloud tunnel so devices outside the LAN can also reach it.

## 版本号说明 / Versioning

本分支版本号采用 `v0.1.NNNN-fn.M` 规则：前段 **v0.1.NNNN 对齐主仓库（yetone/magpie）版本号**（本版基于上游 v0.1.1150），`-fn` 标识飞牛封装分支，`M` 为本分支封装迭代号（本版 13；覆盖安装必须使用更高的版本号）。

This fork uses `v0.1.NNNN-fn.M`: the leading **v0.1.NNNN tracks the upstream (yetone/magpie) version** (this release is based on upstream v0.1.1150), `-fn` marks the fnOS packaging branch, and `M` is the packaging iteration number (13 here; upgrading requires a higher version).

## 为什么封装成飞牛版 / Why this fnOS build

**最直接的原因：作者自己有一台飞牛 NAS。** 官方桌面版需要每台电脑单独安装、单独配置模型与订阅；而 NAS 天然是"常开、集中、家庭 / 团队共享"的设备。把 magpie 装进 NAS，等于把"AI 网关"变成一件常驻的家庭设备，这也是本项目存在的全部理由，以及加入云隧道的原因——让内网之外、以及其他人的设备也能用上这台网关。

The simplest reason: the maintainer owns a fnOS NAS. The official desktop app must be installed and configured per computer; a NAS is always-on, centralized and shared by a household / team — so putting magpie on the NAS turns the AI gateway into an always-on home appliance. That is the entire reason for this project, and the reason the Cloud tunnel was added: to reach the gateway from outside the LAN, or to share it with others.

挂在 NAS 上的用途：

1. **多台电脑共用一套模型 / 订阅配置**：NAS 常开，各电脑 Agent 直连即可，无需每台安装配置；
2. **局域网 Agent 模型支持**：Claude Code / Codex / Gemini CLI / OpenCode 等直连 NAS 网关（3425）；
3. **砍掉本机 Agent 配置覆盖**：NAS 版只提供服务端网关，模型 / 订阅 / 路由统一在 Web UI 配置；
4. **自托管、数据不出内网**；需要外网访问时，用内置云隧道（可选）暴露临时或固定的公网入口。

## 本版重点 / Highlights (v0.1.1150-fn.13)

1. **云隧道真正落地（本版核心）**：
   - 镜像内置 **cloudflared** 与独立的**隧道管理服务 tunnel-admin（端口 3431）**，与 magpie 一起随容器启动；
   - 隧道管理页：`http://<NAS-IP>:3431`，提供快速隧道与命名隧道两种形态：
     - **快速隧道（Quick）**：免 Cloudflare 账号，一键开启即获得临时 `*.trycloudflare.com` 公网 URL，进程运行期间有效，适合临时分享；
     - **命名隧道（Named）**：填 Cloudflare Token + 自有域名，URL 持久稳定，适合长期暴露；
     - 两种隧道均可选择暴露**网关 3425**（供外部 Agent 接入）或 **Web UI 3430**；
   - **实现参考并致谢 omniroute**（[diegosouzapw/OmniRoute](https://github.com/diegosouzapw/OmniRoute)，MIT）：隧道形态采用其 Cloudflare Tunnel 标准做法（快速隧道 `cloudflared tunnel --url` / 命名隧道 `cloudflared tunnel run --token`）。本项目只借鉴隧道方案，不引入 omniroute 主体。
   - 说明：前序版本（fn.12 及更早）文档中描述的"Web UI Settings 内嵌 Cloud tunnel"并未真正实现（上游 magpie 原版 Web UI 无此设置项），**本版为真实实现**，隧道管理以独立页面（3431）承载。
2. **修复镜像名不一致**：Actions 构建推送的镜像名（`ghcr.io/<owner>/magpie`，无 `-fn`）与 compose 引用（`ghcr.nju.edu.cn/mybladethirsts/magpie-fn`）不一致，导致飞牛拉取失败 / 反复重启。本版统一为 **`ghcr.io/mybladethirsts/magpie-fn`**（国内走南大镜像站 `ghcr.nju.edu.cn`）。
3. **默认免密钥直进延续（fn.12 起）**：不注入 `MAGPIE_WEB_KEY`，Web 界面打开直接进入，局域网直接可用；可选保护：compose 中手动设置 `MAGPIE_WEB_KEY`（≥16 位字母数字）后重启。
4. **覆盖安装修复延续**：compose 不依赖回调占位符，容器直接启动；升级回调清理旧版残留；覆盖安装必须使用更高版本号（本版 v0.1.1150-fn.13 > 0.1.1150-fn.12）。

## 使用 / Usage

1. 安装后打开 Web UI：应用中心点图标，或局域网 `http://<NAS-IP>:3430`（**直接进入，无需密钥**）；
2. 在 Web UI 添加供应商 / 登录订阅、统一选模型；
3. 局域网各电脑 Agent 直连网关：`http://<NAS-IP>:3425`（Claude Code / Codex / Gemini CLI 等，配置示例见仓库 README）；
4. 需要公网访问时，打开**隧道管理页 `http://<NAS-IP>:3431`**：
   - 快速隧道：选暴露目标（网关 3425 / Web 3430）→ 开启 → 复制生成的 `*.trycloudflare.com` URL 分享给外部设备；
   - 命名隧道：粘贴 Cloudflare Token（可选填自有域名）→ 开启；域名映射请在 Cloudflare 控制台 Zero Trust → Tunnels 配置。

## 本分支 / About this fork

本仓库是 yetone/magpie 的社区 fork。上游官方**不接受 Pull Request**（仅维护者可提交，见官方文档 "Community" 一节），因此封装相关改动以本 fork 形式维护。

This repository is a community fork of yetone/magpie. The upstream project does not accept Pull Requests (maintainers only), so packaging changes are maintained in this fork.

**致谢 / Credits**

- 上游 [yetone/magpie](https://github.com/yetone/magpie) 及作者 **yetone**（MIT 许可）；
- 隧道方案参考 [omniroute](https://github.com/diegosouzapw/OmniRoute)（MIT）：其 Cloudflare Tunnel 集成方式（快速隧道 / 命名隧道）是本项目隧道功能的实现范本。

## 安装 / Install

1. 解码 `.fpk.b64` 还原 `.fpk`：Windows `certutil -decode <file>.b64 <file>.fpk`；Linux/macOS `base64 -d <file>.b64 > <file>.fpk`；
2. 飞牛应用中心 → 手动安装 → 选择 `.fpk`（本版可直接覆盖安装 0.1.1150-fn.12）；
3. 安装完成后应用中心点 magpie 图标，或局域网打开 `http://<NAS-IP>:3430`，**直接进入 Web 界面**。

## 资产 / Assets

| 文件 / File | 架构 / Arch |
|---|---|
| magpie-0.1.1150-fn.13-x86.fpk | x86_64 |
| magpie-0.1.1150-fn.13-arm.fpk | ARM64 |

（仓库 fpk/ 目录提供对应 `.fpk.b64` base64 文本，供在线预览 / 审计。）

## 注意 / Notes

- Web 界面默认免密钥：个人 / 家庭 NAS 场景直接可用；若部署在办公室等共享网络，或需要暴露公网，请自行在 compose 中设置 `MAGPIE_WEB_KEY`；
- 网关（3425）官方默认接受任意 key，仅建议可信局域网使用；
- 公网隧道：网关含订阅凭据，**暴露公网前**请设置好保护（见上）；快速隧道 URL 在进程停止后失效，请勿长期依赖；
- 命名隧道需要 Cloudflare 账号与 Token，域名映射在 Cloudflare 控制台完成；
- 本封装为社区作品，与官方无关，不提供任何担保。

## 许可 / License

- 上游 magpie：MIT（Copyright (c) 2026 yetone），LICENSE 全文见包内 `app/LICENSE`；
- 参考项目 omniroute：MIT；
- 本封装遵循 MIT，代码与文档见本仓库。
