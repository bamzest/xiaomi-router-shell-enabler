# Xiaomi Router Shell Enabler

这是一个用于启用小米路由器 SSH 和 Telnet 访问的工具。

## 功能特性

- 自动通过密码获取 stok，无需手动获取
- 启用/关闭 SSH 和 Telnet 访问
- 检查 SSH 和 Telnet 服务状态
- 执行自定义命令
- 根据序列号计算 SSH 密码
- 支持多种小米路由器型号

## 支持的路由器型号

- Redmi AX5400Pro
- 更多型号将陆续添加...

## 安装

### 从源代码编译

#### 方法 1：直接编译（推荐）

```bash
# 克隆或下载项目（也可以直接下载 ZIP 解压）
git clone https://github.com/bamzest/xiaomi-router-shell-enabler.git
# 或直接下载解压到任意目录

cd xiaomi-router-shell-enabler

# 安装依赖（仅第一次需要，会下载到本地缓存）
go mod download

# 编译（完全离线也可以，只要依赖已下载）
go build -o xiaomi-router-shell-enabler
```

#### 方法 2：使用构建脚本

```bash
./build.sh
# 生成的文件在 bin/ 目录下
```

#### 离线编译说明

本项目使用简洁的模块名称 `xiaomi-router-shell-enabler`，确保：

- ✅ **完全离线编译**：执行 `go mod download` 下载依赖后，可以在无网络环境下编译
- ✅ **任意目录运行**：项目可以放在任何目录，不依赖 `$GOPATH` 或特定路径
- ✅ **跨平台兼容**：导入路径在 Windows/Linux/macOS 上完全一致
- ✅ **易于分发**：整个项目目录可以直接压缩分发，接收者可直接编译

即使没有 GitHub 仓库或网络访问，只要：

1. 本地已安装 Go（1.20+）
2. 执行过一次 `go mod download`（会缓存到 `$GOMODCACHE`，通常是 `~/go/pkg/mod/`）

就可以在完全离线的环境下编译使用。

### 下载预编译版本

请访问 [Releases](https://github.com/bamzest/xiaomi-router-shell-enabler/releases) 页面下载适合您系统的预编译版本。

## 使用方法

### 启用 SSH 和 Telnet

```bash
./xiaomi-router-shell-enabler -model redmi_ax5400pro -host 192.168.31.1 -password YOUR_PASSWORD -enable_shell
```

### 关闭 SSH 和 Telnet

```bash
./xiaomi-router-shell-enabler -model redmi_ax5400pro -host 192.168.31.1 -password YOUR_PASSWORD -disable_shell
```

### 检查 SSH 和 Telnet 状态

```bash
./xiaomi-router-shell-enabler -model redmi_ax5400pro -host 192.168.31.1 -password YOUR_PASSWORD -shell_status
```

### 执行自定义命令

```bash
./xiaomi-router-shell-enabler -model redmi_ax5400pro -host 192.168.31.1 -password YOUR_PASSWORD -exec "cat /etc/passwd"
```

### 计算 SSH 密码

```bash
./xiaomi-router-shell-enabler -sn YOUR_SERIAL_NUMBER -calc-password
```

### 显示支持的路由器型号

```bash
./xiaomi-router-shell-enabler -list
```

### 显示版本信息

```bash
./xiaomi-router-shell-enabler -version
```

### 启用详细日志

在任何命令后添加 `-verbose` 参数可以显示详细的调试信息：

```bash
./xiaomi-router-shell-enabler -model redmi_ax5400pro -host 192.168.31.1 -password YOUR_PASSWORD -enable_shell -verbose
```

## 参数说明

- `-host`: 路由器 IP 地址，默认为 192.168.31.1
- `-password`: 路由器管理密码
- `-model`: 路由器型号，如 redmi_ax5400pro
- `-enable_shell`: 启用 SSH 和 Telnet
- `-disable_shell`: 关闭 SSH 和 Telnet
- `-shell_status`: 检查 SSH 和 Telnet 状态
- `-exec`: 执行自定义命令
- `-sn`: 路由器序列号，用于计算 SSH 密码
- `-calc-password`: 仅计算并显示 SSH 密码
- `-list`: 显示支持的路由器型号
- `-version`: 显示版本信息
- `-verbose`: 显示详细日志

## 注意事项

- 请确保您有合法权限访问和管理路由器
- 启用 SSH 可能会影响路由器的安全性，请谨慎使用
- 本工具仅供学习和研究使用

## 许可证

本项目采用 [GNU Affero General Public License v3.0 (AGPL-3.0)](LICENSE) 许可证，并附加禁止商业使用条款。

### 禁止商业使用

本软件及其衍生作品禁止用于商业目的，包括但不限于：

1. 销售或许可软件或任何衍生作品以获取金钱或其他商业补偿
2. 使用软件提供商业服务或产品
3. 将软件用作商业产品或服务的一部分
4. 将软件纳入商业产品或服务
5. 任何主要用于商业优势或金钱补偿的活动

如需商业使用，请联系版权持有者获取商业许可。