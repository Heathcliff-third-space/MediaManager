# Makefile for MediaManager

.PHONY: test test-unit test-integration test-all clean build run deps help docker-build docker-push

# 运行所有测试
test: test-unit test-integration

# 运行单元测试
test-unit:
	@echo "运行单元测试..."
	go test -v ./internal/config
	go test -v ./internal/api

# 运行集成测试
test-integration:
	@echo "运行集成测试..."
	go test -v ./internal/api/connection_test.go
	go test -v ./internal/bot/connection_test.go

# 运行完整集成测试
test-all:
	@echo "运行所有测试..."
	go test -v ./...

# 清理构建产物
clean:
	@echo "清理中..."
	rm -f MediaManager
	rm -f MediaManager

# 构建项目
build:
	@echo "构建项目..."
	go build -ldflags="-s -w" -o MediaManager cmd/bot/main.go

# 运行项目
run: build
	@echo "运行项目..."
	./MediaManager

# 安装依赖
deps:
	@echo "安装依赖..."
	HTTPS_PROXY=http://127.0.0.1:7890 HTTP_PROXY=http://127.0.0.1:7890 go mod tidy

# 构建 Docker 镜像
docker-build:
	@echo "构建 Docker 镜像..."
	docker build -t MediaManager .

# 推送 Docker 镜像到仓库 (需要设置 REPO 变量)
docker-push:
	@echo "推送 Docker 镜像到仓库..."
ifndef REPO
	@echo "错误: 请设置 REPO 变量，例如: make docker-push REPO=myrepo/MediaManager"
	@exit 1
endif
	docker tag MediaManager $(REPO):latest
	docker push $(REPO):latest

# 帮助信息
help:
	@echo "可用命令:"
	@echo "  make test-unit       - 运行单元测试"
	@echo "  make test-integration - 运行集成测试"
	@echo "  make test-all        - 运行所有测试"
	@echo "  make test            - 运行单元测试和集成测试"
	@echo "  make build           - 构建项目"
	@echo "  make run             - 构建并运行项目"
	@echo "  make deps            - 安装依赖"
	@echo "  make clean           - 清理构建产物"
	@echo "  make docker-build    - 构建 Docker 镜像"
	@echo "  make docker-push     - 推送 Docker 镜像到仓库 (需要设置 REPO 变量)"