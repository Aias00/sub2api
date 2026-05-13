# Claude Code 快速开始指南

本页说明如何把 Claude Code 接入 cloudbase。不同版本的 Claude Code 配置方式可能调整，安装命令请以工具官方文档为准；cloudbase 侧以控制台调用说明页展示的参数为准。

## 准备工作

- 已注册 cloudbase 账号。
- 账号有可用余额或订阅。
- 已创建 API Key。
- 本机已安装 Node.js。

## 配置步骤

1. 打开 cloudbase 控制台。
2. 进入 **调用说明** 页面。
3. 选择用于 Claude Code 的 API Key。
4. 复制 Base URL、API Key 和模型名称。
5. 按 Claude Code 当前版本要求写入环境变量或配置文件。

常见配置形态：

```bash
export ANTHROPIC_BASE_URL="以调用说明页为准"
export ANTHROPIC_API_KEY="你的 cloudbase API Key"
```

如果你的工具使用 OpenAI 兼容接口，则使用对应的 OpenAI Base URL 和 Key 字段。

## 验证

启动 Claude Code 后先提一个小问题，例如：

```text
请用一句话说明当前目录是什么项目。
```

如果请求成功，再执行更复杂的代码任务。

## 常见问题

- 认证失败：检查 API Key 是否完整复制。
- 模型不存在：检查当前分组是否开放该模型。
- 网络超时：检查本地代理、DNS 或公司网络限制。
- 扣费异常：到使用记录里按时间筛选请求明细。
