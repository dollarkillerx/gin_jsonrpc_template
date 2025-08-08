package scaffold

var templates = map[string]string{
	"go.mod": `module {{.Module}}

go 1.24.5

require (
	github.com/gin-gonic/gin v1.10.1
	github.com/natefinch/lumberjack v2.0.0+incompatible
	github.com/redis/go-redis/v9 v9.12.0
	github.com/rs/zerolog v1.34.0
	github.com/spf13/viper v1.20.1
	gorm.io/driver/postgres v1.6.0
	gorm.io/gorm v1.30.1
)
`,
	"configs/config.toml": `[ServiceConfiguration]
Port = "{{.HTTPPort}}"
Debug = true

[LoggerConfiguration]
Filename = "./logs/{{.AppName}}.log"
MaxSize = 10

[PostgresConfiguration]
Host = "127.0.0.1"
Port = 5432
User = "postgres"
Password = "postgres"
DBName = "{{.AppName}}"
SSLMode = false
TimeZone = "Asia/Shanghai"

[RedisConfiguration]
Addr = "127.0.0.1:6379"
Db = 0
Password = ""
`,
	"cmd/main.go": `package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"{{.Module}}/internal/conf"
	"{{.Module}}/internal/server"
	"{{.Module}}/internal/storage"
	"{{.Module}}/pkg/common/client"
	"{{.Module}}/pkg/common/config"
	logs "{{.Module}}/pkg/common/log"
)

var configFilename string
var configDirs string

func init() {
	const (
		defaultConfigFilename = "config"
		defaultConfigDirs     = "./,./configs/"
	)
	flag.StringVar(&configFilename, "c", defaultConfigFilename, "Name of the config file, without extension")
	flag.StringVar(&configFilename, "dev_config", defaultConfigFilename, "Name of the config file, without extension")
	flag.StringVar(&configDirs, "cPath", defaultConfigDirs, "Directories to search for config file, separated by ','")
}

func main() {
	flag.Parse()

	var appConfig conf.Config
	if err := config.InitConfiguration(configFilename, strings.Split(configDirs, ","), &appConfig); err != nil {
		panic(err)
	}
	if b, err := json.MarshalIndent(appConfig, "", "  "); err == nil {
		fmt.Println(string(b))
	}
	fmt.Println("Config loaded successfully!")

	// Logger
	logs.InitLog(appConfig.LoggerConfiguration)

	// Postgres
	pg, err := client.PostgresClient(appConfig.PostgresConfiguration, nil)
	if err != nil {
		log.Error().Msg("Failed to connect to postgres")
		panic(err)
	}

	// Redis
	rc, err := client.RedisClient(appConfig.RedisConfiguration)
	if err != nil {
		log.Error().Msg("Failed to connect to redis")
		panic(err)
	}

	st := storage.NewStorage(rc, pg)
	log.Info().Msg("Storage initialized")

	s := server.NewServer(st, appConfig)
	if err := s.Run(); err != nil {
		log.Error().Msgf("Failed to start server %s", err)
	}
}
`,
	"internal/conf/config.go": `package conf

import "{{.Module}}/pkg/common/config"

type Config struct {
	ServiceConfiguration  config.ServiceConfiguration  ` + "`mapstructure:\"ServiceConfiguration\"`" + `
	PostgresConfiguration config.PostgresConfiguration ` + "`mapstructure:\"PostgresConfiguration\"`" + `
	RedisConfiguration    config.RedisConfiguration    ` + "`mapstructure:\"RedisConfiguration\"`" + `
	LoggerConfiguration   config.LoggerConfig          ` + "`mapstructure:\"LoggerConfiguration\"`" + `
}
`,
	"internal/server/server.go": `package server

import (
	"{{.Module}}/internal/api"
	"{{.Module}}/internal/conf"
	"{{.Module}}/internal/storage"
)

type Server struct {
	storage   *storage.Storage
	apiServer *api.ApiServer
	conf      conf.Config
}

func NewServer(storage *storage.Storage, conf conf.Config) *Server {
	return &Server{storage: storage, apiServer: api.NewApiServerWithDeps(storage, conf), conf: conf}
}

func (s *Server) Run() error { return s.apiServer.Run() }
`,
	"internal/api/api.go": `package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"{{.Module}}/internal/conf"
	"{{.Module}}/internal/middleware"
	"{{.Module}}/internal/storage"
)

type ApiServer struct {
	storage    *storage.Storage
	conf       conf.Config
	app        *gin.Engine
	rpcHandler *RpcHandler
}

func NewApiServerWithDeps(storage *storage.Storage, conf conf.Config) *ApiServer {
	server := &ApiServer{
		storage:    storage,
		conf:       conf,
		rpcHandler: NewRpcHandler(),
	}
	server.registerRpcMethods()
	return server
}

func (a *ApiServer) Run() error {
	if a.app == nil {
		a.app = gin.New()
		a.app.Use(middleware.HttpRecover())
		a.app.Use(gin.Logger())
		a.app.Use(middleware.Cors())
	}
	a.Router()
	return a.app.Run(fmt.Sprintf("127.0.0.1:%s", a.conf.ServiceConfiguration.Port))
}

func (a *ApiServer) Router() {
	a.app.GET("/health", a.HealthCheck)
	a.app.POST("/api/rpc", a.Rpc)
}

func (a *ApiServer) HealthCheck(ctx *gin.Context) {
	ctx.JSON(200, gin.H{"status": "healthy", "message": "ok"})
}

func (a *ApiServer) Rpc(ctx *gin.Context) {
	a.rpcHandler.HandleRpcRequest(ctx)
}

func (a *ApiServer) registerRpcMethods() {
	a.rpcHandler.RegisterMethod(&PingMethod{})
	a.rpcHandler.RegisterMethod(&EchoMethod{})
}
`,
	"internal/api/rpc_handler.go": `package api

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
	"{{.Module}}/pkg/common/resp"
)

type RpcMethod interface {
	Name() string
	Execute(ctx context.Context, params json.RawMessage) (interface{}, error)
	RequireAuth() bool
}

type RpcHandler struct {
	methods map[string]RpcMethod
	mu      sync.RWMutex
}

func NewRpcHandler() *RpcHandler { return &RpcHandler{methods: make(map[string]RpcMethod)} }

func (h *RpcHandler) RegisterMethod(method RpcMethod) {
	h.mu.Lock(); defer h.mu.Unlock(); h.methods[method.Name()] = method
}

func (h *RpcHandler) getMethod(name string) (RpcMethod, bool) {
	h.mu.RLock(); defer h.mu.RUnlock(); m, ok := h.methods[name]; return m, ok
}

func (h *RpcHandler) HandleRpcRequest(ctx *gin.Context) {
	var request resp.RpcRequest
	if err := ctx.ShouldBindJSON(&request); err != nil { resp.ErrorReturn(ctx, request.Id, fmt.Errorf("invalid request: %v", err)); return }
	if request.JsonRPC != "2.0" { resp.ErrorReturn(ctx, request.Id, fmt.Errorf("unsupported jsonrpc version: %s", request.JsonRPC)); return }
	method, exists := h.getMethod(request.Method)
	if !exists { resp.ErrorReturn(ctx, request.Id, fmt.Errorf("method not found: %s", request.Method)); return }
	result, err := method.Execute(ctx, request.Params)
	if err != nil { resp.ErrorReturn(ctx, request.Id, err); return }
	resp.SimpleReturn(ctx, request.Id, result)
}
`,
	"internal/api/rpc_methods.go": `package api

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type PingMethod struct{}
func (m *PingMethod) Name() string { return "ping" }
func (m *PingMethod) Execute(ctx context.Context, params json.RawMessage) (interface{}, error) {
	return map[string]any{"pong": true, "time": time.Now().Unix(), "message": "pong"}, nil
}
func (m *PingMethod) RequireAuth() bool { return false }

type EchoMethod struct{}
func (m *EchoMethod) Name() string { return "echo" }
func (m *EchoMethod) Execute(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var input map[string]any
	if len(params) > 0 { if err := json.Unmarshal(params, &input); err != nil { return nil, fmt.Errorf("invalid params: %v", err) } }
	return map[string]any{"echo": input, "time": time.Now().Unix()}, nil
}
func (m *EchoMethod) RequireAuth() bool { return false }
`,
	"internal/middleware/cors.go": `package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Accept, Content-Type,AccessToken,X-CSRF-Token, Authorization, Token,X-Token,X-UserID-Id")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT,DELETE,OPTIONS,PATCH")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")
		if method == http.MethodOptions { c.AbortWithStatus(http.StatusNoContent); return }
		c.Next()
	}
}
`,
	"internal/middleware/recover.go": `package middleware

import (
	"errors"
	"runtime"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"{{.Module}}/pkg/common/resp"
)

func HttpRecover() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stackTrace := debug.Stack(); runtime.Stack(stackTrace, true)
				log.Error().Msgf("HttpRecover url: %s stackTrace %s", ctx.Request.URL.Path, string(stackTrace))
				resp.ErrorReturn(ctx, "recover", errors.New("Internal Server Error"))
			}
		}()
		ctx.Next()
	}
}
`,
	"internal/storage/storage.go": `package storage

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Storage struct { redis *redis.Client; db *gorm.DB }
func NewStorage(redisConn *redis.Client, db *gorm.DB) *Storage { return &Storage{redis: redisConn, db: db} }
func (s *Storage) GetRedis() *redis.Client { return s.redis }
func (s *Storage) GetDB() *gorm.DB { return s.db }
`,
	"pkg/common/config/config.go": `package config

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type PostgresConfiguration struct { Host string ` + "`mapstructure:\"Host\"`" + `; Port int ` + "`mapstructure:\"Port\"`" + `; User string ` + "`mapstructure:\"User\"`" + `; Password string ` + "`mapstructure:\"Password\"`" + `; DBName string ` + "`mapstructure:\"DBName\"`" + `; SSLMode bool ` + "`mapstructure:\"SSLMode\"`" + `; TimeZone string ` + "`mapstructure:\"TimeZone\"`" + ` }

type ServiceConfiguration struct { Port string ` + "`mapstructure:\"Port\"`" + `; Debug bool ` + "`mapstructure:\"Debug\"`" + ` }

type RedisConfiguration struct { Addr string ` + "`mapstructure:\"Addr\"`" + `; Db int ` + "`mapstructure:\"Db\"`" + `; Password string ` + "`mapstructure:\"Password\"`" + ` }

type LoggerConfig struct { Filename string ` + "`mapstructure:\"Filename\"`" + `; MaxSize int ` + "`mapstructure:\"MaxSize\"`" + ` }

func InitConfiguration(configName string, configPaths []string, config interface{}) error {
	vp := viper.New(); vp.SetConfigName(configName); vp.AutomaticEnv();
	for _, p := range configPaths { vp.AddConfigPath(p) }
	if err := vp.ReadInConfig(); err != nil { if _, ok := err.(viper.ConfigFileNotFoundError); !ok { return errors.WithStack(err) } }
	if err := vp.Unmarshal(config); err != nil { return errors.WithStack(err) }
	for _, key := range vp.AllKeys() { if err := vp.BindEnv(key); err != nil { return errors.WithStack(err) } }
	return nil
}
`,
	"pkg/common/client/pgsql.go": `package client

import (
	"fmt"
	"log"
	"os"
	"time"

	"{{.Module}}/pkg/common/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func PostgresClient(conf config.PostgresConfiguration, gormConfig *gorm.Config) (*gorm.DB, error) {
	if conf.TimeZone == "" { conf.TimeZone = "Asia/Tokyo" }
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d TimeZone=%s", conf.Host, conf.User, conf.Password, conf.DBName, conf.Port, conf.TimeZone)
	if !conf.SSLMode { dsn += " sslmode=disable" }
	if gormConfig == nil {
		gormConfig = &gorm.Config{ Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{ SlowThreshold: time.Second, LogLevel: logger.Info, IgnoreRecordNotFoundError: true, Colorful: true }) }
	}
	return gorm.Open(postgres.Open(dsn), gormConfig)
}
`,
	"pkg/common/client/redis.go": `package client

import (
	"context"

	"github.com/redis/go-redis/v9"
	"{{.Module}}/pkg/common/config"
)

func RedisClient(conf config.RedisConfiguration) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{ Addr: conf.Addr, Password: conf.Password, DB: conf.Db })
	ctx := context.Background(); if _, err := client.Ping(ctx).Result(); err != nil { return nil, err }
	return client, nil
}
`,
	"pkg/common/log/log.go": `package logs

import (
	"os"
	"path/filepath"
	"time"

	"github.com/natefinch/lumberjack"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"{{.Module}}/pkg/common/config"
)

func InitLog(loggerConfig config.LoggerConfig) {
	if loggerConfig.Filename != "" {
		_ = os.MkdirAll(filepath.Dir(loggerConfig.Filename), 0o755)
	}

	rotatingLogger := &lumberjack.Logger{ Filename: loggerConfig.Filename, MaxSize: loggerConfig.MaxSize, MaxBackups: 1, MaxAge: 28, Compress: true }
	consoleWriter := zerolog.ConsoleWriter{ Out: os.Stdout, TimeFormat: time.RFC3339 }
	multi := zerolog.MultiLevelWriter(consoleWriter, rotatingLogger)
	log.Logger = zerolog.New(multi).With().Caller().Timestamp().Logger()
	log.Info().Msg("Logger initialized")
}
`,
	"pkg/common/resp/resp.go": `package resp

import ( "encoding/json"; "github.com/gin-gonic/gin" )

type RpcRequest struct { JsonRPC string ` + "`json:\"jsonrpc\"`" + `; Method string ` + "`json:\"method\"`" + `; Params json.RawMessage ` + "`json:\"params\"`" + `; Id string ` + "`json:\"id\"`" + ` }

type RpcError struct { Code int ` + "`json:\"code\"`" + `; Message string ` + "`json:\"message\"`" + `; Data interface{} ` + "`json:\"data,omitempty\"`" + ` }

type RpcResponse struct { JsonRPC string ` + "`json:\"jsonrpc\"`" + `; Id string ` + "`json:\"id\"`" + `; Result json.RawMessage ` + "`json:\"result,omitempty\"`" + `; Error *RpcError ` + "`json:\"error,omitempty\"`" + ` }

func SimpleReturn(ctx *gin.Context, id string, data interface{}) { Return(ctx, 200, id, data, nil) }
func ErrorReturn(ctx *gin.Context, id string, err error) { Return(ctx, 200, id, nil, err) }
func Return(ctx *gin.Context, code int, id string, data interface{}, err error) {
	response := RpcResponse{ JsonRPC: "2.0", Id: id }
	if err != nil { response.Error = &RpcError{ Code: -32000, Message: err.Error() } } else { jsonData, _ := json.Marshal(data); response.Result = jsonData }
	ctx.JSON(code, response)
}
`,
    "Makefile": `SHELL := /bin/bash

APP_NAME ?= {{.AppName}}
FTINIT_NAME ?= ftinit
BIN_DIR ?= bin
GO ?= go

.PHONY: help deps tidy build build-app build-ftinit run test clean

help:
	@echo "Available targets:"
	@echo "  tidy          - go mod tidy"
	@echo "  build         - build app and ftinit binaries into $(BIN_DIR)/"
	@echo "  build-app     - build main app binary into $(BIN_DIR)/$(APP_NAME)"
	@echo "  build-ftinit  - build scaffold tool into $(BIN_DIR)/$(FTINIT_NAME)"
	@echo "  run           - run the server with default config path"
	@echo "  test          - run unit tests"
	@echo "  clean         - remove $(BIN_DIR)/"

# Aliases
 deps: tidy

tidy:
	$(GO) mod tidy

build: build-app build-ftinit

build-app:
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/$(APP_NAME) ./cmd

build-ftinit:
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/$(FTINIT_NAME) ./cmd/ftinit

run:
	$(GO) run ./cmd -c config -cPath "./,./configs/"

test:
	$(GO) test ./...

clean:
	rm -rf $(BIN_DIR)
`,
    ".gitignore": `# Binaries
/ftinit
*.exe
*.dll
*.so
*.dylib

# Build
/bin/
/dist/

# Logs
/logs/*
!/logs/.gitkeep

# IDE
.idea/
.vscode/

# OS
.DS_Store
Thumbs.db

# Env
.env
`,
    "logs/.gitkeep": ``,
}
