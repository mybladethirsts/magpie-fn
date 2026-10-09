# magpie for fnOS（飞牛安装包 / fnOS Package）

## 本仓库为 fork / About this fork

本仓库是 [yetone/magpie](https://github.com/yetone/magpie) 的社区 fork，用于发布飞牛 fnOS 封装（fpk/ 目录）。上游 magpie 官方**不接受 Pull Request**（仅维护者可提交，见官方文档"Community"一节），因此封装相关改动以本 fork 形式维护。

This repository is a community fork of [yetone/magpie](https://github.com/yetone/magpie) hosting the fnOS packaging (in `fpk/`). The upstream magpie project **does not accept Pull Requests** (only maintainers can commit, see the "Community" section of the official docs), so all packaging changes are maintained in this fork.

由衷感谢原作者 **yetone** 与 magpie 社区（MIT 许可）。Credit goes to the original author **yetone** and the magpie community (MIT License).

## 定位 / Positioning

**集中 AI 网关：多台电脑共用一套模型配置。**
**A centralized AI gateway: one model configuration shared by all your computers.**

将开源 AI 模型路由网关 yetone/magpie 封装为飞牛 fnOS 应用（.fpk），镜像由本 fork 构建（`ghcr.nju.edu.cn/mybladethirsts/magpie-fn`，国内拉取快；网络好可改回 `ghcr.io/mybladethirsts/magpie-fn`）。

## 为什么封装成飞牛版 / Why this fnOS build

**最直接的原因：作者自己有一台飞牛 NAS。**
The simplest reason: the maintainer owns a fnOS NAS.

magpie 官方的桌面版需要每台电脑单独安装、单独配置模型与订阅；而 NAS 天然是"常开、集中、家庭/团队共享"的设备。把 magpie 装进 NAS，等于把"每家每户的 AI 网关"变成一件常驻设备，这也是本项目存在的全部理由，以及后续加入公网隧道（Cloud tunnel）的原因——让内网之外、以及其他人也能用上这台网关。

The official desktop app must be installed and configured per computer. A NAS, however, is always-on, centralized, and shared by a family or team — so putting magpie on the NAS turns an "AI gateway per machine" into a "gateway that lives on the always-on home server". That is the entire reason for this project, and the reason the Cloud tunnel was added later: to reach that gateway from outside the LAN, or to share it with others.

## 挂在 NAS 上的用途 / What it is for on your NAS

1. **多台电脑共用一套模型 / 订阅配置**：NAS 常开，7×24 提供网关服务，各电脑无需各自安装配置；
   **One model/subscription config for all computers**: the NAS is always on, and every computer's agent just points to it.
2. **局域网 Agent 模型支持**：Claude Code / Codex / Gemini CLI / OpenCode 等直接连 NAS 网关（3425）即获得模型能力，无需每台电脑安装 magpie；
   **LAN agent support**: Claude Code / Codex / Gemini CLI / OpenCode etc. connect straight to the NAS gateway (3425).
3. **砍掉本机 Agent 配置覆盖**：官方桌面版安装时会自动改写本机 agent 配置；NAS 版只提供服务端网关，模型 / 订阅 / 路由统一在 Web UI 配置，各客户端直连；
   **No local agent config rewriting**: the NAS edition only serves the gateway; models / subscriptions / routing are all managed in the Web UI.
4. **自托管、数据不出内网**：订阅凭据与路由数据留在自己的 NAS 上；需要外网访问时，可用内置 Cloud tunnel（可选）暴露一个临时或固定的公网入口。
   **Self-hosted, data stays on your LAN**; when outside access is needed, the built-in Cloud tunnel (optional) exposes a temporary or named public URL.

## 改动清单（vs 上游官方镜像）/ Changes vs upstream

1. **权限**：compose 使用 `user: "0"`（容器内以 root 运行保证 /config 可写）；
2. **局域网访问**：Web UI 3430 与网关 3425 均开放 `0.0.0.0`，供局域网 Agent 直连；Web UI 有密钥保护；
3. **访问密钥随机化（v0.1.1121-10 起）**：不再内置固定默认密钥。安装时自动生成每台设备唯一的随机密钥（主机信息 + 时间 + 随机熵 → 16 位 hex），持久化保存于 NAS 数据目录（`/var/apps/magpie/shares/magpie/data/web_key`）；应用中心点图标直达（浏览器地址栏可见 `?k=` 密钥），容器日志也会显著打印。升级 / 重装不改变已保存的密钥；
4. **云隧道（Cloud tunnel）**：镜像内置 cloudflared，Web UI「Settings → Network and sharing → Cloud tunnel」一键开启：快速隧道（临时 `*.trycloudflare.com` 公网 URL）或固定隧道（自有域名 + CF Token），可暴露网关（3425）或 Web UI（3430）；
5. **镜像源**：默认 `ghcr.nju.edu.cn/mybladethirsts/magpie-fn`（国内拉取快），可改回 `ghcr.io`；
6. **修复反复重启 / 覆盖安装**：compose 不依赖安装回调占位符，容器启动自带密钥兜底逻辑；升级回调清理旧版残留，覆盖安装必须使用更高版本号；
7. **版本规则**：`0.1.1121-N`，N 为封装版本号（递增；覆盖安装必须使用更高的版本号）。

## 文件 / Files

| 文件 / File | 架构 / Arch | 说明 / Notes |
|---|---|---|
| magpie-0.1.1121-11-x86.fpk.b64 | x86_64 | 安装包 base64 文本，解码后为 .fpk |
| magpie-0.1.1121-11-arm.fpk.b64 | ARM64 | 同上（按飞牛规范打包，未在真机验证，ARM 用户请先在测试环境安装） |

## 安装 / Install

1. 还原 .fpk：`certutil -decode magpie-0.1.1121-11-x86.fpk.b64 magpie-0.1.1121-11-x86.fpk`（Windows）或 `base64 -d <file>.b64 > <file>.fpk`（Linux/macOS）；
2. 飞牛应用中心 → 手动安装 → 选择 .fpk；
3. 安装向导可手动设置访问密钥（≥16 位字母数字）；**留空则自动生成随机密钥**，安装完成后应用中心点 magpie 图标即可直达，密钥见浏览器地址栏 `?k=` 与容器日志。

## 使用 / Usage

1. 浏览器打开 Web UI（应用中心点图标直达），添加供应商 / 登录订阅、统一选模型；
2. 每台电脑配置 Agent 直连 NAS：
   - Claude Code：`claude config set --global baseURL http://<NAS-IP>:3425`
   - Codex：`OPENAI_BASE_URL=http://<NAS-IP>:3425/v1`，`OPENAI_API_KEY=<gateway key>`
   - Gemini CLI：`GOOGLE_GEMINI_BASE_URL=http://<NAS-IP>:3425`，`GEMINI_API_KEY=<gateway key>`
3. 多设备建议在 Web UI「Settings → Share on local network」为每台电脑创建独立 gateway key。

## 访问 / Access

- Web UI：应用中心点图标直达（带随机密钥 `?k=`）；局域网 `http://<NAS-IP>:3430`（按提示输入密钥）
- 网关：`http://<NAS-IP>:3425`
- 访问密钥：**安装时自动随机生成，每台 NAS 唯一**；持久保存于数据目录 `web_key` 文件，应用中心直达 URL 与容器日志均可见。忘记可查看应用日志，或删除数据目录 web_key 后重启（会重新生成，旧密钥失效）。

## 公网隧道（可选）/ Public tunnels (optional)

- **Cloud tunnel（已内置）**：Web UI「Settings → Network and sharing → Cloud tunnel」
  - Quick tunnel（快速）：一键开启，生成临时 `*.trycloudflare.com` 公网 URL，进程运行期间有效，重启后 URL 变化；
  - Named tunnel（固定）：填 Cloudflare 账号 Token + 自有域名，URL 持久；
  - 可暴露网关（3425）或 Web UI（3430）。
- 提醒：网关含订阅凭据，暴露公网前请先开启 Web UI 的 Share 并设置 gateway key 认证。

## 许可 / License

- 上游 magpie：MIT（Copyright (c) 2026 yetone），LICENSE 全文见包内；
- 本封装为社区作品，与官方无关，不提供任何担保。