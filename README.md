# MediaManager Bot

这是一个基于 Telegram Bot 的多服务器媒体管理工具。

## 功能特点

- 通过 Telegram Bot 控制多种媒体服务器（Audiobookshelf、Emby等）
- 查询服务器信息（版本、运行状态、资源使用情况等）
- 查询用户信息和统计
- 查询媒体库、媒体项信息
- 跨服务器搜索媒体内容
- 管理用户和媒体库
- 访问控制功能，仅允许指定用户使用机器人

## 快速开始

### 环境要求

- Go 1.21 或更高版本
- Telegram Bot Token
- 媒体服务器（Audiobookshelf、Emby等）访问权限

### 安装步骤

1. 克隆项目:
   ```
   git clone <repository-url>
   cd MediaManager
   ```

2. 安装依赖:
   
   由于网络限制，需要使用代理拉取 Go 依赖:
   ```
   HTTPS_PROXY=http://127.0.0.1:7890 HTTP_PROXY=http://127.0.0.1:7890 go mod tidy
   ```

3. 配置环境变量:

   程序会优先从 `conf/.env` 文件加载配置，如果不存在则会尝试从 `.env` 文件加载。推荐使用 `conf` 目录来管理配置文件:
   ```
   # 创建 conf 目录（如果尚不存在）
   mkdir -p conf
   
   # 复制示例配置文件
   cp conf/.env.example conf/.env
   ```
   
   编辑 `conf/.env` 文件，填写以下信息:
   ```
   TELEGRAM_BOT_TOKEN=your_telegram_bot_token
   AUDIOBOOKSHELF_URL=http://localhost:13378         # 可选，默认为 localhost:13378
   AUDIOBOOKSHELF_PORT=13378                         # 可选，默认为 13378
   AUDIOBOOKSHELF_TOKEN=your_audiobookshelf_token
   EMBY_URL=http://localhost:8096                    # 可选，Emby服务器地址
   EMBY_PORT=8096                                    # 可选，默认为 8096
   EMBY_TOKEN=your_emby_token                        # 可选，Emby API 密钥
   PROXY_ADDRESS=127.0.0.1:7890                      # 可选，仅用于 Telegram 和 Go 依赖的代理，默认为 127.0.0.1:7890
   DEBUG=true                                        # 可选，启用调试模式
   ALLOWED_USER_IDS=123456789,987654321              # 可选，允许使用机器人的用户ID列表，多个ID用逗号分隔
   ```

4. 运行程序:
   
   同样需要使用代理拉取依赖:
   ```
   HTTPS_PROXY=http://127.0.0.1:7890 HTTP_PROXY=http://127.0.0.1:7890 go run cmd/bot/main.go
   ```
   
   或者编译后运行:
   ```
   HTTPS_PROXY=http://127.0.0.1:7890 HTTP_PROXY=http://127.0.0.1:7890 go build -o bot cmd/bot/main.go
   ./bot
   ```

## 使用 Docker 部署

项目支持通过 Docker 进行部署，提供了更加简便的部署方式。

### 构建 Docker 镜像

```
make docker-build
```

或者直接使用 Docker 命令:
```
docker build -t audiobookshelf-manager .
```

注意：在国内网络环境下构建镜像可能会遇到网络问题，建议：
1. 使用代理进行构建
2. 配置 Docker 使用国内镜像加速器
3. 使用 `--network host` 参数

### 运行容器

运行容器时需要挂载配置文件或传递环境变量:

#### 方式一：使用环境变量

```
docker run -d \
  --name audiobookshelf-manager \
  -e TELEGRAM_BOT_TOKEN=your_telegram_bot_token \
  -e AUDIOBOOKSHELF_URL=http://your-abs-server:13378 \
  -e AUDIOBOOKSHELF_TOKEN=your_audiobookshelf_token \
  -e EMBY_URL=http://your-emby-server:8096 \
  -e EMBY_TOKEN=your_emby_token \
  audiobookshelf-manager
```

#### 方式二：使用配置文件

首先创建配置文件 `conf/.env`，然后挂载到容器中:

```
docker run -d \
  --name audiobookshelf-manager \
  -v $(pwd)/conf/.env:/root/conf/.env \
  audiobookshelf-manager
```

### 推送镜像到仓库

```
make docker-push REPO=your-repo/audiobookshelf-manager
```

## 使用方法

在 Telegram 中找到你的 Bot 并开始对话。目前已实现的功能包括：

### 菜单系统
机器人提供直观的菜单系统，可通过以下方式访问：
- 发送 `/start` 命令打开主菜单
- 使用 Telegram 的命令菜单（在输入框下方或通过 "/" 触发）

### 服务器信息查询
通过菜单中的「📊 服务器信息」按钮或发送 `/serverinfo` 命令，可以获得：
- 所有已配置媒体服务器的版本信息
- 服务器操作系统和硬件信息
- 运行时间和资源使用情况（内存、磁盘）
- 媒体库概览
- 用户统计信息

## 测试

项目包含多种类型的测试用例，确保各组件正常工作：

### 单元测试
```
go test -v ./internal/config    # 测试配置加载
go test -v ./internal/api       # 测试 API 客户端
```

### 集成测试
```
go test -v ./internal/api/connection_test.go   # 测试 Audiobookshelf 连接
go test -v ./internal/bot/connection_test.go   # 测试 Telegram Bot 连接
```

### 完整集成测试
```
go test -v ./...               # 运行所有测试
```

或者使用 Makefile:
```
make test-unit                 # 运行单元测试
make test-integration          # 运行集成测试
make test                      # 运行所有测试
```

注意：集成测试需要有效的 TOKEN 和网络连接。

## 注意事项

- 代理设置 (`PROXY_ADDRESS`) 仅用于连接 Telegram API 和拉取 Go 依赖
- 连接媒体服务器时不使用代理
- 如果不需要代理访问 Telegram，则可以留空 `PROXY_ADDRESS` 配置
- `ALLOWED_USER_IDS` 用于限制机器人访问，如果不设置则允许所有用户访问

## 项目结构

```
.
├── cmd/
│   └── bot/           # 主程序入口
├── conf/              # 配置文件目录
│   ├── .env           # 实际环境变量文件（需自行配置）
│   └── .env.example   # 环境变量示例文件
├── internal/
│   ├── api/           # 媒体服务器 API 客户端
│   ├── bot/           # Telegram Bot 相关逻辑
│   ├── config/        # 配置管理
│   ├── models/        # 数据模型
│   ├── services/      # 业务逻辑
│   └── util/          # 工具函数
└── .env               # 实际环境变量文件（备选位置）
```

## 开发指南

1. 添加新功能时，请遵循现有的项目结构
2. 在 internal/ 目录中添加相关的组件
3. 保持代码简洁和可维护性

## 贡献

欢迎提交 Issue 和 Pull Request。