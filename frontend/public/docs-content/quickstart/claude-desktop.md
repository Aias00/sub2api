# Claude Desktop 第三方 Provider 配置指南

Claude Desktop 的第三方 Provider 能力取决于客户端版本和插件生态。cloudbase 侧提供的是统一网关、API Key 和用量管理。

## 准备

- Claude Desktop 已安装并可正常启动。
- cloudbase API Key 已创建。
- 调用说明页可正常显示网关信息。

## 配置思路

将 Provider 的请求入口指向 cloudbase：

```text
Base URL: 以 cloudbase 调用说明页为准
API Key: 你的 cloudbase API Key
Model: 当前分组开放的模型
```

如果你通过 MCP、代理工具或第三方扩展接入，请确认该中间层没有覆盖 Base URL 或 Key。

## 验证

配置完成后发送一条短消息。然后在 cloudbase 使用记录中确认是否出现对应请求。

## 安全提醒

第三方插件可能读取本机配置文件。只安装你信任的插件，并定期检查 API Key 是否异常消耗。
