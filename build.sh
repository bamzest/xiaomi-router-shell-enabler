#!/bin/bash

# 构建脚本

# 设置版本信息
VERSION="0.1.0"
BUILD_TIME=$(date -u '+%Y-%m-%d %H:%M:%S')
GIT_COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "unknown")

# 创建输出目录
mkdir -p bin

echo "开始构建 Xiaomi Router SSH Enabler v${VERSION}..."
echo "构建时间: ${BUILD_TIME}"
echo "Git提交: ${GIT_COMMIT}"

# 使用正确的ldflags传递版本信息
go build -ldflags "-X 'xiaomi-router-shell-enabler/pkg/version.Version=${VERSION}' -X 'xiaomi-router-shell-enabler/pkg/version.BuildTime=${BUILD_TIME}' -X 'xiaomi-router-shell-enabler/pkg/version.GitCommit=${GIT_COMMIT}'" -o bin/xiaomi-router-shell-enabler

if [ $? -eq 0 ]; then
    echo "构建成功! 可执行文件位于 bin/xiaomi-router-shell-enabler"
    chmod +x bin/xiaomi-router-shell-enabler
else
    echo "构建失败!"
    exit 1
fi
