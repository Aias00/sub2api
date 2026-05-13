# Hermes 快速开始指南

Hermes 接入 cloudbase 的方式与其他自定义 Provider 客户端类似：配置 Base URL、API Key 和模型名称。

## 接入步骤

1. 登录 [cloudbase](https://cloudbase.eu.org)。
2. 创建 API Key。
3. 在 **调用说明** 页面复制当前 Key 的调用参数。
4. 在 Hermes 的模型服务配置里新增 cloudbase Provider。
5. 保存后发送测试请求。

## 测试建议

先使用短文本请求验证连通性，再进行长上下文或文件分析任务。这样可以避免因配置错误造成不必要的重复消耗。

## 常见问题

- 如果返回认证错误，重新复制 API Key。
- 如果返回模型错误，检查模型名称是否和分组一致。
- 如果没有任何响应，检查本地网络代理和客户端日志。
