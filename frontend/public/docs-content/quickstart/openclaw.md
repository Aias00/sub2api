# OpenClaw 快速开始指南

OpenClaw 这类客户端通常支持自定义模型服务。接入 cloudbase 的关键是把 Provider 指向当前站点的网关地址。

## 配置字段

```text
Provider: cloudbase
Base URL: 调用说明页展示的地址
API Key: cloudbase API Key
Model: 当前分组开放的模型
```

## 操作流程

1. 在 cloudbase 创建一个专用于 OpenClaw 的 API Key。
2. 打开 **调用说明**，确认该 Key 绑定的分组和模型。
3. 在 OpenClaw 中新增 Provider。
4. 保存后发送测试消息。

## 建议

为 OpenClaw 单独创建 Key，便于统计和停用。如果你还同时使用 Codex 或 Claude Code，不要所有工具共用同一个 Key。
