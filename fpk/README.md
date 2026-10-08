# magpie for fnOS（飞牛安装包）

## 定位：多台电脑共用一套模型配置的集中 AI 网关

将开源 AI 模型路由网关 yetone/magpie 封装为飞牛 fnOS 应用（.fpk）。
镜像默认使用南大 ghcr 镜像站（ghcr.nju.edu.cn/yetone/magpie），国内拉取快；如网络好或自建中转，可在 Docker 设置改回 ghcr.io 官方源。

与官方桌面版不同，本包**不涉及本机 Agent 配置覆盖**：模型 / 订阅 / 路由统一在 Web UI 配置，
各电脑的 Agent 直连 NAS 网关，共用同一套模型配置（适合多台电脑共用订阅与模型列表）。

## 文件

- magpie-0.1.1121-8.fpk.b64 — 安装包（base64 文本，解码还原为 .fpk 后安装）

## 安装

1. 还原 .fpk：
   - Windows：certutil -decode magpie-0.1.1121-8.fpk.b64 magpie-0.1.1121-8.fpk
   - Linux/macOS：base64 -d magpie-0.1.1121-8.fpk.b64 > magpie-0.1.1121-8.fpk
2. 飞牛应用中心 → 手动安装 → 选择 .fpk；
3. 安装向导可设置 Web UI 访问密钥（至少 16 位，留空用默认密钥）。

## 使用流程

1. 浏览器打开 Web UI（应用中心点图标直达），添加供应商 / 登录订阅、统一选模型；
2. 每台电脑配置 Agent 直连 NAS：
   - Claude Code：claude config set --global baseURL http://<NAS-IP>:3425
   - Codex：OPENAI_BASE_URL=http://<NAS-IP>:3425/v1，OPENAI_API_KEY=<gateway key>
   - Gemini CLI：GOOGLE_GEMINI_BASE_URL=http://<NAS-IP>:3425，GEMINI_API_KEY=<gateway key>
3. 多设备建议在 Web UI「Settings → Share on local network」为每台电脑创建独立 gateway key。

## 访问

- Web UI：应用中心点图标直达；局域网 http://<NAS-IP>:3430/?k=<密钥>
- 网关：http://<NAS-IP>:3425
- 默认密钥：a3f9c27e5b8d41f6a9c3e7b2d4f81a6c

## 公网隧道（可选）

- Cloudflare 快速隧道：compose 内置 cloudflared sidecar 模板，取消注释填 CF token 即可；
- Tailscale / ngrok：见包内 app/README.md。

## 许可

- 上游 magpie：MIT（Copyright (c) 2026 yetone）；
- 社区封装，与官方无关，不提供任何担保。
