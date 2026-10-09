# magpie for fnOS v0.1.1121-11

Release Notes / 发布说明

## 概述 / Overview

将开源 AI 模型路由网关 [yetone/magpie](https://github.com/yetone/magpie) 封装为飞牛 fnOS 安装包（.fpk）。定位：**集中 AI 网关**——多台电脑共用一套模型 / 订阅配置，NAS 7×24 提供服务，各电脑 Agent 直连即可获得模型能力。

This release packages yetone/magpie as a fnOS application (.fpk). Positioning: **a centralized AI gateway** — one model / subscription configuration shared by all your computers, served 24×7 by the NAS.

## 为什么封装成飞牛版 / Why this fnOS build

**最直接的原因：作者自己有一台飞牛 NAS。** 官方桌面版需要每台电脑单独安装、单独配置；NAS 常开且全家 / 全团队共享，把 magpie 装进 NAS，等于把"AI 网关"变成一件常驻的家庭设备。挂在 NAS 上的用途：

1. 多台电脑共用一套模型 / 订阅配置（NAS 常开，无需每台安装）；
2. 局域网 Agent 模型支持（Claude Code / Codex / Gemini CLI / OpenCode 直连网关 3425）；
3. 砍掉本机 Agent 配置覆盖（模型 / 订阅 / 路由统一在 Web UI 配置）；
4. 自托管、数据不出内网；需要外网访问时用内置云隧道（这也是加入 Cloud tunnel 的原因）。

The simplest reason: the maintainer owns a fnOS NAS. The NAS is always-on and shared by the household / team, which turns the AI gateway into an always-on home appliance — and the built-in Cloud tunnel exists precisely to reach it from outside the LAN.

## 本版重点 / Highlights (v0.1.1121-11)

1. **访问密钥随机化 + 16 位长度**：移除内置固定默认密钥，避免"所有安装者共用一个密钥"。v10 起安装时自动生成**每台设备唯一**的随机密钥；本版将随机密钥精简为 **16 位 hex**（满足官方 ≥16 字符要求），文档与向导文案同步更新。
   - 安装时自动生成**每台设备唯一**的随机密钥（主机信息 + 时间 + 随机熵 → 16 位 hex，满足官方 ≥16 字符要求）；
   - 密钥持久化保存于 NAS 数据目录（`/var/apps/magpie/shares/magpie/data/web_key`），升级 / 重装不改变已保存的密钥；
   - **显著展示**：安装向导提示 + 应用中心点图标直达（浏览器地址栏可见 `?k=` 密钥）+ 容器日志显著打印；
   - 向导也可手动填写你自己的密钥（≥16 位字母数字）；
   - 容器首次启动自带兜底：web_key 不存在时自动生成，不依赖任何安装回调。
2. **云隧道（Cloud tunnel）**：镜像内置 cloudflared，Web UI「Settings → Network and sharing → Cloud tunnel」：
   - 快速隧道：一键开启临时 `*.trycloudflare.com` 公网 URL（进程运行期间有效）；
   - 固定隧道：填 Cloudflare 账号 Token + 自有域名，URL 持久；
   - 可暴露网关（3425）或 Web UI（3430）。
3. **覆盖安装 / 反复重启修复延续**：compose 不依赖回调占位符；升级回调清理旧版残留（含旧固定默认密钥）；覆盖安装必须使用更高版本号。

## 本分支 / About this fork

本仓库是 yetone/magpie 的社区 fork。上游官方**不接受 Pull Request**（仅维护者可提交，见官方文档 "Community" 一节），因此封装相关改动以本 fork 形式维护。由衷感谢原作者 **yetone** 与 magpie 社区（MIT 许可）。

This repository is a community fork of yetone/magpie. The upstream project does not accept Pull Requests (maintainers only), so packaging changes are maintained in this fork. Credit to the original author **yetone** and the magpie community (MIT License).

## 安装 / Install

1. 解码 `.fpk.b64` 还原 `.fpk`：Windows `certutil -decode <file>.b64 <file>.fpk`；Linux/macOS `base64 -d <file>.b64 > <file>.fpk`；
2. 飞牛应用中心 → 手动安装 → 选择 `.fpk`；
3. 安装向导可设置 Web UI 访问密钥（≥16 位字母数字）；**留空则自动生成随机密钥**，安装完成后应用中心点图标直达，密钥见浏览器地址栏与容器日志。

## 资产 / Assets

| 文件 / File | 架构 / Arch |
|---|---|
| magpie-0.1.1121-11-x86.fpk | x86_64 |
| magpie-0.1.1121-11-arm.fpk | ARM64 |

（仓库 fpk/ 目录提供对应 `.fpk.b64` base64 文本，供在线预览 / 审计。）

## 注意 / Notes

- 访问密钥为安装时自动生成的随机密钥，**每台 NAS 唯一**；忘记可查看容器日志，或删除数据目录 web_key 后重启（会重新生成，旧密钥失效）；
- 公网隧道：网关含订阅凭据，暴露公网前请先开启 Web UI 的 Share 并设置 gateway key 认证；
- 本封装为社区作品，与官方无关，不提供任何担保。

## 许可 / License

- 上游 magpie：MIT（Copyright (c) 2026 yetone），LICENSE 全文见包内 `app/LICENSE`；
- 本封装遵循 MIT，代码与文档见本仓库。