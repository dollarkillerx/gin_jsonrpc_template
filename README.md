### Gin JSON-RPC Template (Feitian)

基于 Gin 的 JSON-RPC 2.0 模板工程，内置：统一配置加载、日志滚动、Redis/GORM(Postgres) 客户端封装、基础中间件（CORS/Recover）、健康检查、JSON-RPC 路由与方法注册机制，以及项目脚手架 `ftinit`，用于快速生成同结构的新项目。

- **框架**: Gin
- **RPC**: JSON-RPC 2.0（`/api/rpc`）
- **存储**: Postgres（GORM）、Redis
- **日志**: zerolog + lumberjack 滚动
- **配置**: Viper（支持多目录查找与环境变量）
- **中间件**: CORS、Panic Recover
- **脚手架**: `cmd/ftinit` 一键生成新项目

---

### 目录结构

```
.
├── cmd/                    # 主程序入口（支持 -c/-cPath 加载配置）
│   ├── main.go
│   └── ftinit/             # 项目脚手架入口
│       └── main.go
├── configs/
│   └── config.toml         # 默认配置文件
├── internal/
│   ├── api/                # HTTP/JSON-RPC 服务
│   │   ├── api.go          # Gin Engine 初始化、路由与启动
│   │   ├── rpc_handler.go  # JSON-RPC 路由器与方法调度
│   │   └── rpc_methods.go  # 示例方法：ping / echo
│   ├── conf/               # 业务配置结构体
│   ├── middleware/         # 通用中间件（CORS/Recover）
│   ├── scaffold/           # ftinit 模板与生成逻辑
│   ├── server/             # Server 聚合
│   └── storage/            # 统一存储聚合（Redis/GORM）
├── pkg/common/
│   ├── client/             # Redis / Postgres 客户端
│   ├── config/             # 配置加载（Viper 封装）
│   ├── log/                # 日志初始化
│   └── resp/               # JSON-RPC 请求/响应结构与返回助手
├── Makefile                # 常用构建/运行命令
├── go.mod / go.sum
└── README.md
```

---

### 快速开始

- **准备**
  - Go（建议 1.22+；本模板 `go.mod` 声明为 1.24.5）
  - 本地 Postgres 与 Redis（或修改配置连接你自己的实例）

- **拉取依赖**
  - 使用 Make：`make tidy`
  - 或直接：`go mod tidy`

- **本地运行**
  - 使用 Make：`make run`
  - 或直接：
    ```bash
    go run ./cmd -c config -cPath "./,./configs/"
    ```
  - 默认监听：`127.0.0.1:8080`

- **健康检查**
  - `GET http://127.0.0.1:8080/health`

---

### 配置

默认从 `-c` 指定的名称（不含扩展名）与 `-cPath` 指定的目录列表中查找 TOML 配置（默认 `-c config` 与 `-cPath "./,./configs/"`）。配置结构参见 `internal/conf/config.go` 与 `pkg/common/config/config.go`。

示例 `configs/config.toml`（仓库自带）：

```toml
[ServiceConfiguration]
Port = "8080"
Debug = true

[LoggerConfiguration]
Filename = "./logs/feitian.log"
MaxSize = 10

[PostgresConfiguration]
Host = "127.0.0.1"
Port = 6532
User = "root"
Password = "root"
DBName = "feitian"

[RedisConfiguration]
Addr = "127.0.0.1:6779"
Db = 1
Password = ""
```

提示：上述端口/账号仅为示例，请改为你的环境参数。Server 运行时绑定 `127.0.0.1:{Port}`，如需对外暴露，可在 `internal/api/api.go` 的 `Run()` 中调整绑定地址。

---

### 运行与构建（Makefile）

- `make tidy`：执行 `go mod tidy`
- `make build`：构建主程序与 `ftinit` 到 `bin/`
- `make build-app`：仅构建主程序（默认可执行名 `feitian`）
- `make build-ftinit`：仅构建脚手架 `ftinit`
- `make run`：本地运行（使用默认配置参数）
- `make test`：运行全部单测（如有）
- `make clean`：清理 `bin/`

构建产物：
- `bin/feitian`：服务二进制
- `bin/ftinit`：脚手架二进制

---

### JSON-RPC 接口

- **Endpoint**: `POST /api/rpc`
- **Content-Type**: `application/json`
- **协议**: JSON-RPC 2.0

请求体结构（简化）：
```json
{
  "jsonrpc": "2.0",
  "method": "<method-name>",
  "params": { ... },
  "id": "<request-id>"
}
```

返回体结构（简化）：
```json
{
  "jsonrpc": "2.0",
  "id": "<request-id>",
  "result": { ... },
  "error": { "code": -32000, "message": "..." }
}
```

内置方法：见 `internal/api/rpc_methods.go`

- `ping`（无鉴权）示例：
  ```bash
  curl -s http://127.0.0.1:8080/api/rpc \
    -H 'Content-Type: application/json' \
    -d '{"jsonrpc":"2.0","method":"ping","params":{},"id":"1"}'
  ```

- `echo`（无鉴权）示例：
  ```bash
  curl -s http://127.0.0.1:8080/api/rpc \
    -H 'Content-Type: application/json' \
    -d '{"jsonrpc":"2.0","method":"echo","params":{"msg":"hello"},"id":"2"}'
  ```

如需新增方法：
- 实现接口 `RpcMethod`（见 `internal/api/rpc_handler.go`）
- 在 `ApiServer.registerRpcMethods()` 中注册。

---

### 日志

- 初始化：`pkg/common/log`（zerolog + lumberjack）
- 配置：`[LoggerConfiguration]` 中设置日志文件与最大大小，日志同时输出到控制台与滚动文件。

---

### 存储

- Redis 客户端：`pkg/common/client/redis.go`
- Postgres(GORM) 客户端：`pkg/common/client/pgsql.go`
- 统一注入：`internal/storage/storage.go`，通过 `server.NewServer()` -> `api.NewApiServerWithDeps()` 传递到业务层。

---

### 配置与参数说明

- 启动参数：
  - `-c`：配置名（不含扩展名），默认 `config`
  - `-cPath`：配置搜索目录（逗号分隔），默认 `"./,./configs/"`
  - 额外地，`-dev_config` 与 `-c` 等价（保留兼容）

- 配置加载：`pkg/common/config.InitConfiguration()` 使用 Viper 读取 TOML，并支持环境变量覆盖。

---

### 使用脚手架生成新项目（ftinit）

你可以用当前模板作为“母体”，通过 `ftinit` 在指定目录生成一个同结构的新项目：

- 构建脚手架：`make build-ftinit`（生成 `bin/ftinit`）
- 或直接运行：
  ```bash
  go run ./cmd/ftinit \
    -module github.com/you/yourapp \
    -name app \
    -out ./out_dir \
    -port 8080
  ```

完成后会在 `-out` 指定目录生成：
- `go.mod`（替换为你的 `module`）
- `configs/config.toml`（端口、日志名等自动替换）
- 完整的 `internal/`、`pkg/common/` 代码骨架

进入新项目后：
```bash
cd ./out_dir
go mod tidy
go run ./cmd -c config -cPath "./,./configs/"
```

---

### 开发建议

- 新增 JSON-RPC 方法时，遵循接口 `RpcMethod`，并在集中注册处进行注册，方便统一管理与鉴权接入。
- 生产环境请将监听地址改为 `0.0.0.0` 并完善鉴权、限流、追踪等。
- 合理拆分业务逻辑到 `internal` 与 `pkg`，保持公共能力的可复用与边界清晰。
