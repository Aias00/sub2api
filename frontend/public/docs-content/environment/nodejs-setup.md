# Node.js 环境安装指南

很多 AI 编码工具或配置脚本依赖 Node.js。建议安装当前 LTS 版本，并确保终端可以访问 `node`、`npm` 和 `npx`。

## 检查是否已安装

在终端执行：

```bash
node -v
npm -v
npx -v
```

如果三条命令都能输出版本号，说明基础环境已经可用。

## macOS

推荐使用 Homebrew 或 Node 官方安装包。

```bash
brew install node
```

安装后重新打开终端，再执行版本检查。

## Windows

推荐使用 Node 官方 LTS 安装包。安装时保留默认的 “Add to PATH” 选项。

安装完成后打开新的 PowerShell：

```powershell
node -v
npm -v
npx -v
```

## Linux

可以使用系统包管理器，或使用 Node 官方推荐的版本管理方式安装 LTS 版本。

```bash
node -v
npm -v
```

如果系统仓库版本过旧，建议切换到 Node 官方源或版本管理工具。

## 常见问题

### command not found

说明命令没有进入 PATH。重新打开终端；如果仍然失败，检查 Node 安装路径是否写入系统环境变量。

### 权限错误

全局安装 npm 包时如果出现 `EACCES`，优先修复 npm 全局目录权限，不建议长期使用 `sudo npm install -g`。

### 网络超时

如果安装依赖时超时，检查本地代理、公司网络或镜像源配置。不要把代理地址写进项目代码中。
