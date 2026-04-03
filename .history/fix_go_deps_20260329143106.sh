#!/bin/bash
echo "=== 一键修复 Go 依赖拉取失败（protobuf/网络超时）==="

# 1. 彻底清理缓存
go clean -modcache -cache
echo "✅ 清理 Go 缓存完成"

# 2. 配置国内代理 + 关闭校验
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=off
echo "✅ 代理与校验配置完成"

# 3. 进入你的项目目录
cd /Users/liunian/Desktop/dnmp/py_project/backend || exit
echo "✅ 已进入项目 backend 目录"

# 4. 修复 protobuf 镜像
go mod edit -dropreplace=google.golang.org/protobuf
go mod edit -replace=google.golang.org/protobuf=github.com/protocolbuffers/protobuf-go@v1.34.1
echo "✅ 已修复 protobuf 依赖路径"

# 5. 拉取指定版本
go get google.golang.org/protobuf@v1.34.1

# 6. 同步所有依赖
go mod tidy

echo -e "\n🎉 一键修复全部完成！"
