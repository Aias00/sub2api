# cloudbase 使用指南

cloudbase 是一个面向 AI 开发工具的统一接入网关。你可以把 Claude Code、Codex、Cherry Studio、Claude Desktop 等工具接到同一个平台账号下，用一组 API Key 管理模型调用、余额、订阅和用量统计。

控制台地址：[https://cloudbase.eu.org](https://cloudbase.eu.org)

## 什么是 cloudbase？

cloudbase 把多个上游 AI 服务统一到一个可控入口，核心目标是降低接入和维护成本。

- 一个账号管理余额、订阅、订单和 API Key。
- 一个网关地址适配多种开发工具和调用场景。
- 按分组管理不同模型、渠道和倍率策略。
- 在仪表盘里查看请求量、Token、消费和响应时间。

## 核心功能

### 统一 API Key

你只需要在控制台创建 API Key，再把它填入 Claude Code、Codex、Cherry Studio 或其他支持自定义 Provider 的工具中。

### 网关调用说明

调用说明页会根据当前 API Key 绑定的分组，生成对应的网关地址、认证头和调用示例。实际接入时优先以控制台页面显示的地址为准。

### 余额与订阅

平台支持余额充值和订阅套餐。管理员可以配置商品，用户侧会按已配置商品展示购买入口。

### 用量可视化

仪表盘和使用记录页可以帮助你定位成本来源、模型使用频率和异常请求。

## 快速开始

### 1. 注册账号

打开 [https://cloudbase.eu.org/register](https://cloudbase.eu.org/register)，使用邮箱或已开放的第三方登录方式创建账号。

### 2. 获取额度

进入 [充值/订阅](https://cloudbase.eu.org/purchase) 页面选择可用商品。页面只展示管理员已经配置并开放的商品。

### 3. 创建 API Key

进入控制台的 **API 密钥** 页面，创建一个新的密钥。建议为不同工具分别创建 Key，方便后续统计和停用。

### 4. 配置开发工具

根据你使用的工具选择对应指南：

- [Claude Code 快速开始指南](quickstart/claude-code)
- [Codex 快速开始指南](quickstart/codex)
- [Cherry Studio 快速开始指南](quickstart/cherry-studio)
- [Claude Desktop 第三方 Provider 配置指南](quickstart/claude-desktop)

## 获取帮助

如果接入过程中遇到问题，优先准备这几项信息：

- 当前使用的工具和版本。
- 控制台调用说明页展示的 Base URL。
- 报错截图或终端日志。
- 使用的模型名称和请求时间。

不要在聊天、工单或截图中泄露完整 API Key。
