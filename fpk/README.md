# magpie for fnOS（飞牛安装包 / fnOS Package）

## 本仓库为 fork / About this fork

本仓库是 [yetone/magpie](https://github.com/yetone/magpie) 的社区 fork，用于发布飞牛 fnOS 封装（fpk/ 目录）。上游 magpie 官方**不接受 Pull Request**（仅维护者可提交），因此封装相关改动以本 fork 形式维护。

This repository is a community fork of [yetone/magpie](https://github.com/yetone/magpie) hosting the fnOS packaging (in `fpk/`). The upstream magpie project **does not accept Pull Requests** (only maintainers can commit), so all packaging changes are maintained in this fork.

由衷感谢原作者 **yetone** 与 magpie 社区（MIT 许可）。
Credit goes to the original author **yetone** and the magpie community (MIT License).

## 定位 / Positioning

**集中 AI 网关：多台电脑共用一套模型配置。**
**A centralized AI gateway: one model configuration shared by all your computers.**

将开源 AI 模型路由网关 yetone/magpie 封装为飞牛 fnOS 应用（.fpk）。
This repo packages the open-source AI model routing gateway [yetone/magpie](https://github.com/yetone/magpie) as a fnOS application (.fpk).

镜像默认使用南大 ghcr 镜像站（ghcr.nju.edu.cn/mybladethirsts/magpie-fn），国内拉取快；网络好可改回 ghcr.io/mybladethirsts/magpie-fn。
The image defaults to the Nanjing University ghcr mirror (ghcr.nju.edu.cn/mybladethirsts/magpie-fn) for faster pulls in mainland China; switch back to `ghcr.io/mybladethirsts/magpie-fn` if your network allows.

## 为什么封装成飞牛版 / Why this fnOS build

1. **多台电脑共用一套模型 / 订阅配置**：NAS 常开，7×24 提供网关服务，各电脑无需各自安装配置；
2. **局域网 Agent 模型支持**：Claude Code / Codex / Gemini CLI / OpenCode 等直接连 NAS 网关（3425）即获得模型能力；
3. **砍掉本机 Agent 配置覆盖**：官方桌面版安装时会自动改写本机 agent 配置；NAS 版只提供服务端网关，模型 / 订阅 / 路由统一在 Web UI 配置，各客户端直连；
4. **自托管、数据不出内网**（可选公网隧道见下）。

## 改动清单（vs 上游官方）/ Changes vs upstream

1. **免密钥直进（fn.15）**：magpie 主程序新增 `MAGPIE_WEB_NO_AUTH` 开关（compose 默认 `"1"`），Web UI 点图标直接进入，不再弹 Key 输入页；如需密钥认证，设置 `MAGPIE_WEB_KEY`（≥16 字符）会覆盖免密；
2. **内置云隧道（fn.13，参考 omniroute 的 Cloudflare 隧道做法）**：镜像内集成 cloudflared + 隧道管理服务 tunnel-admin（端口 3431），支持快速隧道（trycloudflare 临时公网 URL，免账号）与命名隧道（固定域名，需 Cloudflare Token）；可暴露网关 3425 或 Web 3430；
3. **局域网访问**：Web UI 3430 与网关 3425 均开放 `0.0.0.0`，供局域网 Agent 直连；
4. **权限**：compose 使用 `user: "0"`（飞牛生命周期脚本需 root 执行 chown，容器内以 root 运行保证 /config 可写）；
5. **镜像源**：默认 `ghcr.nju.edu.cn/mybladethirsts/magpie-fn`（国内拉取快），可改回官方 ghcr.io；compose 带 `pull_policy: always`，升级 / 启动自动拉取最新镜像；
6. **版本规则**：`0.1.1150-fn.N`，对齐上游 magpie 版本号 + `-fn.N` 封装迭代号（严格递增，覆盖安装必须使用更高的版本号）。

## 文件 / Files

| 文件 / File | 架构 / Arch | 说明 / Notes |
|---|---|---|
| magpie-0.1.1150-fn.15-x86.fpk.b64 | x86_64 | 当前版安装包（base64 文本，解码后为 .fpk） |
| magpie-0.1.1150-fn.15-arm.fpk.b64 | ARM64 | 当前版安装包 |
| magpie-0.1.1150-fn.14-x86.fpk.b64 | x86_64 | 上一版（回退用） |
| magpie-0.1.1150-fn.14-arm.fpk.b64 | ARM64 | 上一版（回退用） |

## 安装 / Install

1. 还原 .fpk：`certutil -decode magpie-0.1.1150-fn.15-x86.fpk.b64 magpie-0.1.1150-fn.15-x86.fpk`（Windows）或 `base64 -d ... > magpie-0.1.1150-fn.15-x86.fpk`（Linux/macOS）；
2. 飞牛应用中心 → 手动安装 → 选择 .fpk；
3. 安装完成后点应用图标直接进入 Web UI（免密钥，不再弹 Key）。

## 使用 / Usage

1. 浏览器打开 Web UI（应用中心点图标直达），添加供应商 / 登录订阅、统一选模型；
2. 每台电脑配置 Agent 直连 NAS：
   - Claude Code：`claude config set --global baseURL http://<NAS-IP>:3425`
   - Codex：`OPENAI_BASE_URL=http://<NAS-IP>:3425/v1`，`OPENAI_API_KEY=<gateway key>`
   - Gemini CLI：`GOOGLE_GEMINI_BASE_URL=http://<NAS-IP>:3425`，`GEMINI_API_KEY=<gateway key>`
3. 多设备建议在 Web UI「Settings → Share on local network」为每台电脑创建独立 gateway key。

## 访问 / Access

- Web UI：应用中心点图标直达；局域网 `http://<NAS-IP>:3430`（免密钥）
- 隧道管理页：`http://<NAS-IP>:3431`（快速隧道 / 命名隧道）
- 网关：`http://<NAS-IP>:3425`

## 公网隧道（可选）/ Public tunnels (optional)

- 快速隧道（Quick Tunnel）：免账号一键开启，生成临时 `*.trycloudflare.com` 公网 URL，进程运行期间有效；
- 命名隧道（Named Tunnel）：填 Cloudflare Token + 自有域名，URL 持久，重启自动重连；
- 提醒：网关含订阅凭据，暴露公网前请先开启 Web UI 的 Share 并设置 gateway key 认证；快速隧道有 200 并发上限且不支持 SSE，长期使用建议命名隧道。

## 许可 / License

- 上游 magpie：MIT（Copyright (c) 2026 yetone），LICENSE 全文见包内；
- 本封装为社区作品（吴观风岳软件工作室），与官方无关，不提供任何担保。
