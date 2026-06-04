# 快速接入

这一组文档面向“已经拿到 cloudbase 账号、额度和 API Key，下一步就是接到具体客户端”的用户。

## 这部分适合谁

- 已经注册并登录 cloudbase
- 已经创建 API Key 或准备创建 API Key
- 想把某个现有工具接到统一网关，而不是继续维护各自独立配置

## 开始前先确认

在阅读具体客户端指南前，建议先完成这三步：

1. 在控制台确认当前账号已有可用额度或有效订阅
2. 在 `API 密钥` 页面创建一个专用于当前客户端的 Key
3. 在 `调用说明` 页面确认当前站点实际使用的 Base URL 和推荐模型

## 推荐阅读顺序

如果你不确定从哪篇开始，按你当前使用的工具进入：

- [Claude Code 快速开始指南](./claude-code)
- [Codex 快速开始指南](./codex)
- [Gemini CLI 快速开始指南](./gemini-cli)
- [OpenClaw 快速开始指南](./openclaw)
- [Hermes 快速开始指南](./hermes)
- [Cherry Studio 快速开始指南](./cherry-studio)
- [Claude Desktop 第三方 Provider 配置指南](./claude-desktop)
- [GPT-Image-2 使用指南](./gpt-image-2)

## 选择指南时的原则

- `Claude Code / Codex / Gemini CLI`
  适合命令行和本地项目工作流
- `Cherry Studio / Claude Desktop / OpenClaw / Hermes`
  适合桌面客户端、图形化交互和多模型切换
- `GPT-Image-2`
  适合图像生成或编辑场景

## 接入完成后检查什么

完成任意一篇快速接入文档后，至少做一次最小验证：

1. 发起一条最简单的测试请求
2. 确认工具没有继续走旧配置
3. 返回 cloudbase 控制台查看：
   - 是否出现对应请求
   - 是否命中了预期模型或分组
   - 用量和成本是否开始累计

## 如果遇到问题

优先准备下面这些信息，再排查：

- 你使用的客户端名称和版本
- 当前使用的 API Key
- 控制台 `调用说明` 页里的 Base URL
- 报错截图或终端输出
- 请求发生的大致时间

不要直接贴出完整 API Key。
