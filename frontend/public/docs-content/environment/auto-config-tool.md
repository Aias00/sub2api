# 自动配置工具

如果你使用的客户端支持自定义 Provider，通常只需要填入三类信息：

- Base URL：来自 cloudbase 调用说明页。
- API Key：来自 API 密钥页面。
- Model：工具或调用说明中展示的模型名称。

## 推荐流程

1. 登录 [cloudbase 控制台](https://cloudbase.eu.org/dashboard)。
2. 创建或选择一个 API Key。
3. 打开 **调用说明** 页面。
4. 复制页面生成的 Base URL 和认证头。
5. 粘贴到目标工具的 Provider、Endpoint 或 API Base 配置中。
6. 发送一次低成本测试请求。

## 配置模板

```text
Provider Name: cloudbase
Base URL: 以控制台调用说明页展示为准
API Key: sk-********
Model: 选择当前分组已开放的模型
```

## 什么时候需要重新配置

- 你更换了 API Key。
- 管理员调整了分组或模型。
- 工具升级后重置了 Provider 配置。
- 从本地环境切换到服务器或新电脑。

## 排查建议

如果配置后无法调用，先用控制台调用说明页里的示例验证，再回到目标工具排查。这样可以区分是平台配置问题，还是客户端配置问题。
