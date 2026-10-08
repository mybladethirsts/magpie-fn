# magpie for fnOS（飞牛安装包）

将开源 AI 模型路由网关 yetone/magpie 封装为飞牛 fnOS 应用安装包（.fpk）。
镜像使用官方 ghcr.io/yetone/magpie，本包仅为部署壳（MIT 许可，LICENSE 见包内 app/LICENSE）。

## 文件

- magpie-0.1.1121-6.fpk.b64 — 安装包（base64 文本，解码还原为 .fpk 后安装）

## 安装

1. 还原 .fpk：
   - Windows：certutil -decode magpie-0.1.1121-6.fpk.b64 magpie-0.1.1121-6.fpk
   - Linux/macOS：base64 -d magpie-0.1.1121-6.fpk.b64 > magpie-0.1.1121-6.fpk
2. 飞牛应用中心 → 手动安装 → 选择 .fpk；
3. 安装向导可设置访问密钥（至少 16 位，留空用默认密钥）。

## 访问

- Web UI：应用中心点图标直达；局域网 http://<NAS-IP>:3430/?k=<密钥>
- 网关（给局域网 agent 用）：http://<NAS-IP>:3425
- 默认密钥：a3f9c27e5b8d41f6a9c3e7b2d4f81a6c

## 公网隧道（可选）

- Cloudflare 快速隧道：compose 内置 cloudflared sidecar 模板，取消注释填 CF token 即可；
- Tailscale / ngrok：见包内 app/README.md。

## 许可

- 上游 magpie：MIT（Copyright (c) 2026 yetone）；
- 社区封装，与官方无关，不提供任何担保。