# 🚀 Release 发布指南

## 如何创建新版本

### 1. 创建并推送标签
```bash
# 创建标签（版本号格式：v主版本.次版本.修订版本）
git tag v1.0.0

# 推送标签到GitHub
git push origin v1.0.0
```

### 2. 自动构建和发布
推送标签后，GitHub Action会自动：
- 🔨 构建5个平台的二进制文件：
  - `tReader-linux-amd64` (Linux x64)
  - `tReader-linux-arm64` (Linux ARM64)
  - `tReader-windows-amd64.exe` (Windows x64)
  - `tReader-darwin-amd64` (macOS Intel)
  - `tReader-darwin-arm64` (macOS Apple Silicon)
- 📦 创建GitHub Release
- 📋 自动生成Release Notes
- ⬆️ 上传所有二进制文件

### 3. 版本信息
每个构建的二进制文件都包含版本信息：
```bash
./tReader --version
```

## 支持的平台

| 平台 | 架构 | 文件名 |
|------|------|--------|
| Linux | x86_64 | `tReader-linux-amd64` |
| Linux | ARM64 | `tReader-linux-arm64` |
| Windows | x86_64 | `tReader-windows-amd64.exe` |
| macOS | Intel | `tReader-darwin-amd64` |
| macOS | Apple Silicon | `tReader-darwin-arm64` |

## 下载和使用

1. 前往 [Releases页面](../../releases)
2. 下载适合你系统的二进制文件
3. 添加执行权限（Linux/macOS）：
   ```bash
   chmod +x tReader-*
   ```
4. 运行程序：
   ```bash
   ./tReader-* [小说文件路径]
   ```

## 版本规范

我们使用 [语义化版本](https://semver.org/lang/zh-CN/)：
- `v1.0.0` - 主版本（不兼容的API修改）
- `v1.1.0` - 次版本（向下兼容的功能性新增）
- `v1.0.1` - 修订版本（向下兼容的问题修正） 