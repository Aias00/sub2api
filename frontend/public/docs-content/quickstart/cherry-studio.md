# Cherry Studio 快速开始指南

Cherry Studio 支持配置多个模型服务。你可以把 cloudbase 添加为一个自定义 Provider，用同一套余额和 API Key 管理桌面端调用。

## 新增 Provider

在 Cherry Studio 的模型服务设置中新增一个 Provider：

```text
名称: cloudbase
API Key: 你的 cloudbase API Key
API 地址: 以调用说明页展示为准
模型: 当前分组开放的模型
```

字段名称可能随客户端版本变化，含义以 “Base URL / Endpoint / API Base” 和 “API Key” 为准。

## 验证

保存后新建一个会话，选择 cloudbase Provider 下的模型，发送一句测试消息。

## 使用建议

- 每个桌面客户端单独创建一个 Key。
- 不要把完整 Key 截图发送给别人。
- 如果更换分组，重新检查模型列表和 Base URL。
