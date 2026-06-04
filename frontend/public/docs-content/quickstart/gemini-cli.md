# Gemini CLI 快速开始指南

如果你希望把 Gemini CLI 也接到 cloudbase，而不是单独维护另一套调用入口，可以按下面方式配置。

## 适用场景

这个指南适合：

- 你已经在使用 Gemini CLI
- 想把使用量、额度和分组管理统一收进 cloudbase
- 希望和 Claude Code / Codex 一样走同一个控制台体系

## 准备工作

- 已注册 cloudbase 账号
- 已有余额或订阅
- 已创建 API Key
- 已确认当前账号有可用的 Gemini 分组

如果还没确认分组，先看 [可用分组](#/console/available-groups)。

## 配置步骤

1. 登录 cloudbase 控制台
2. 打开 **API 密钥**
3. 选择一个适合 Gemini CLI 的 Key
4. 打开 **调用说明**
5. 复制当前页提供的 Base URL、认证方式和模型名称

常见配置思路：

```bash
export GEMINI_API_KEY="你的 cloudbase API Key"
export GEMINI_BASE_URL="以控制台调用说明页展示为准"
```

如果当前客户端并不是直接读取这些变量，而是通过配置文件读取，请把控制台里的对应值写入客户端实际使用的配置位置。

## 第一次验证

建议先发一个很小的请求，例如：

```text
请用一句话说明当前目录是做什么的。
```

如果能成功返回，再开始复杂任务。

## 常见问题

### 看不到 Gemini 相关模型

通常不是 CLI 的问题，而是：

- 当前账号没有 Gemini 分组权限
- 当前 Key 没绑定到正确分组

先检查 [可用分组](#/console/available-groups)。

### 调用走错了旧地址

说明客户端可能还在读取旧配置或旧环境变量。  
重开终端、重启客户端，再验证一次。

### 成本不符合预期

去控制台查看：

- 使用记录
- Key 用量
- 当前 Key 所属分组倍率
