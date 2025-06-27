# Makefile for gotty-piko-client
SHELL=/bin/bash
# 支持多平台编译

# 变量定义
BINARY_NAME=webssh-piko
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go 相关变量
GO=go
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)
CGO_ENABLED?=0

# 编译参数
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT} -s -w"

# 输出目录
DIST_DIR=dist
BUILD_DIR=dist

# 支持的平台
PLATFORMS=linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64

# 默认目标
.PHONY: all
all: clean build

# 清理构建文件
.PHONY: clean
clean:
	@echo "清理构建文件..."
	@rm -rf ${DIST_DIR} ${BUILD_DIR}
	@echo "清理完成"

# 构建当前平台
.PHONY: build
build:
	@echo "构建 ${BINARY_NAME} for ${GOOS}/${GOARCH}..."
	@mkdir -p ${BUILD_DIR}
	cd web && npm run build
	CGO_ENABLED=${CGO_ENABLED} ${GO} build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME} ./main.go
	@echo "构建完成: ${BUILD_DIR}/${BINARY_NAME}"

# 构建所有平台
.PHONY: build-all
build-all: 
	@echo "构建所有平台的 ${BINARY_NAME}..."
	@mkdir -p ${DIST_DIR}
	@for platform in ${PLATFORMS}; do \
		IFS='/' read -r os arch <<< "$$platform"; \
		echo "构建 $$os/$$arch..."; \
		GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 ${GO} build ${LDFLAGS} -o ${DIST_DIR}/${BINARY_NAME}-$$os-$$arch ./main.go; \
		if [ "$$os" = "windows" ]; then \
			mv ${DIST_DIR}/${BINARY_NAME}-$$os-$$arch ${DIST_DIR}/${BINARY_NAME}-$$os-$$arch.exe; \
		fi; \
	done
	@echo "所有平台构建完成，输出目录: ${DIST_DIR}"

# 构建特定平台
.PHONY: build-linux
build-linux:
	@echo "构建 Linux 版本..."
	@mkdir -p ${DIST_DIR}
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 ${GO} build ${LDFLAGS} -o ${DIST_DIR}/${BINARY_NAME}-linux-amd64 ./main.go
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 ${GO} build ${LDFLAGS} -o ${DIST_DIR}/${BINARY_NAME}-linux-arm64 ./main.go
	@echo "Linux 版本构建完成"

.PHONY: build-darwin
build-darwin:
	@echo "构建 macOS 版本..."
	@mkdir -p ${DIST_DIR}
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 ${GO} build ${LDFLAGS} -o ${DIST_DIR}/${BINARY_NAME}-darwin-amd64 ./main.go
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 ${GO} build ${LDFLAGS} -o ${DIST_DIR}/${BINARY_NAME}-darwin-arm64 ./main.go
	@echo "macOS 版本构建完成"

.PHONY: build-windows
build-windows:
	@echo "构建 Windows 版本..."
	@mkdir -p ${DIST_DIR}
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 ${GO} build ${LDFLAGS} -o ${DIST_DIR}/${BINARY_NAME}-windows-amd64.exe ./main.go
	GOOS=windows GOARCH=arm64 CGO_ENABLED=0 ${GO} build ${LDFLAGS} -o ${DIST_DIR}/${BINARY_NAME}-windows-arm64.exe ./main.go
	@echo "Windows 版本构建完成"

# 帮助
.PHONY: help
help:
	@echo "可用的 Make 目标:"
	@echo "  all          - 清理并构建当前平台版本"
	@echo "  build        - 构建当前平台版本"
	@echo "  build-all    - 构建所有平台版本"
	@echo "  build-linux  - 构建 Linux 版本"
	@echo "  build-darwin - 构建 macOS 版本"
	@echo "  build-windows - 构建 Windows 版本"
	@echo "  clean        - 清理构建文件"
	@echo "  help         - 显示此帮助信息" 