# magpie for fnOS（飞牛安装包 / fnOS Package）

## 本仓库为 fork / About this fork

本仓库是 [yetone/magpie](https://github.com/yetone/magpie) 的社区 fork，用于发布飞牛 fnOS 封装（fpk/ 目录）。上游 magpie 官方**不接受 Pull Request**（仅维护者可提交，见官方文档"Community"一节），因此封装相关改动以本 fork 形式维护。

This repository is a community fork of [yetone/magpie](https://github.com/yetone/magpie) hosting the fnOS packaging (in `fpk/`). The upstream magpie project **does not accept Pull Requests** (only maintainers can commit, see the "Community" section of the official docs), so all packaging changes are maintained in this fork.

由衷感谢原作者 **yetone** 与 magpie 社区（MIT 许可）。
Credit goes to the original author **yetone** and the magpie community (MIT License).

## 定位 / Positioning

**集中 AI 网关：多台电脑共用一套模型配置。**
**A centralized AI gateway: one model configuration shared by all your computers.**

将开源 AI 模型路由网关 yetone/magpie 封装为飞牛 fnOS 应用（.fpk）。
This repo packages the open-source AI model routing gateway [yetone/magpie](https://github.com/yetone/magpie) as a fnOS application (.fpk).

镜像默认使用南大 ghcr 镜像站（ghcr.nju.edu.cn/yetone/magpie），国内拉取快；网络好或自建中转时，可在飞牛 Docker 设置改回 ghcr.io 官方源。
The image defaults to the Nanjing University ghcr mirror (ghcr.nju.edu.cn/yetone/magpie) for faster pulls in mainland China; switch back to the official `ghcr.io/yetone/magpie` in the fnOS Docker settings if your network or mirror setup allows.

## 为什么封装成飞牛版 / Why this fnOS build

1. **多台电脑共用一套模型 / 订阅配置**：NAS 常开，7×24 提供网关服务，各电脑无需各自安装配置；
2. **局域网 Agent 模型支持**：Claude Code / Codex / Gemini CLI / OpenCode 等直接连 NAS 网关（3425）即获得模型能力；
3. **砍掉本机 Agent 配置覆盖**：官方桌面版安装时会自动改写本机 agent 配置；NAS 版只提供服务端网关，模型 / 订阅 / 路由统一在 Web UI 配置，各客户端直连；
4. **自托管、数据不出内网**（可选公网隧道见下）。

## 改动清单（vs 上游官方镜像）/ Changes vs upstream

1. **权限**：compose 使用 `user: "0"`（飞牛生命周期脚本非 root 无法执行 chown，容器内以 root 运行保证 /config 可写）；
2. **局域网访问**：Web UI 3430 与网关 3425 均开放 `0.0.0.0`，供局域网 Agent 直连；Web UI 有密钥保护；
3. **Web UI 密钥**：安装向导可设置访问密钥（≥16 字符）；留空则使用内置默认密钥，容器日志会打印；
4. **镜像源**：默认 `ghcr.nju.edu.cn/yetone/magpie`（国内拉取快），可改回官方源；
5. **修复反复重启 / 覆盖安装**：早期版本占位符依赖安装回调替换，回调未执行时容器报错（`MAGPIE_WEB_KEY` 长度不足）反复重启；现改为 compose 内置合法默认密钥（不依赖回调也能启动），升级回调清理旧版残留占位符；
6. **版本规则**：`0.1.1121-N`，N 为封装版本号（递增；覆盖安装必须使用更高的版本号）。

## 文件 / Files

飞牛封装文件见 `fpk/` 目录：

| 文件 / File | 架构 / Arch | 说明 / Notes |
|---|---|---|
| fpk/magpie-0.1.1121-8.fpk.b64 | x86_64 | 安装包 base64 文本，解码后为 .fpk |
| fpk/magpie-0.1.1121-8-arm.fpk.b64 | ARM64 | 同上（按飞牛规范打包，未在真机验证，ARM 用户请先在测试环境安装） |

## 安装 / Install

1. 还原 .fpk：`certutil -decode magpie-0.1.1121-8.fpk.b64 magpie-0.1.1121-8.fpk`（Windows）或 `base64 -d magpie-0.1.1121-8.fpk.b64 > magpie-0.1.1121-8.fpk`（Linux/macOS）；
2. 飞牛应用中心 → 手动安装 → 选择 .fpk；
3. 安装向导可设置 Web UI 访问密钥（至少 16 位，留空用默认密钥）。

## 使用 / Usage

1. 浏览器打开 Web UI（应用中心点图标直达），添加供应商 / 登录订阅、统一选模型；
2. 每台电脑配置 Agent 直连 NAS：
   - Claude Code：`claude config set --global baseURL http://<NAS-IP>:3425`
   - Codex：`OPENAI_BASE_URL=http://<NAS-IP>:3425/v1`，`OPENAI_API_KEY=<gateway key>`
   - Gemini CLI：`GOOGLE_GEMINI_BASE_URL=http://<NAS-IP>:3425`，`GEMINI_API_KEY=<gateway key>`
3. 多设备建议在 Web UI「Settings → Share on local network」为每台电脑创建独立 gateway key。

## 访问 / Access

- Web UI：应用中心点图标直达；局域网 `http://<NAS-IP>:3430/?k=<密钥>`
- 网关：`http://<NAS-IP>:3425`
- 默认密钥：a3f9c27e5b8d41f6a9c3e7b2d4f81a6c

## 公网隧道（可选）/ Public tunnels (optional)

- Cloudflare 快速隧道：compose 内置 cloudflared sidecar 模板，取消注释填 CF token 即可；
- Tailscale / ngrok：见包内 app/README.md。
- 提醒：网关含订阅凭据，暴露公网前请先开启 Web UI 的 Share 并设置 gateway key 认证。

## 许可 / License

- 上游 magpie：MIT（Copyright (c) 2026 yetone），LICENSE 全文见包内；
- 本封装为社区作品，与官方无关，不提供任何担保。
