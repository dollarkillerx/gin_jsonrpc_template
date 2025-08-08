package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/feitian/internal/conf"
	"github.com/google/feitian/internal/middleware"
	"github.com/google/feitian/internal/storage"
)

type ApiServer struct {
	storage    *storage.Storage
	conf       conf.Config
	app        *gin.Engine
	rpcHandler *RpcHandler
}

func NewApiServer(port string) *ApiServer { // kept for backward-compat in case of external usage
	return NewApiServerWithDeps(nil, conf.Config{ServiceConfiguration: struct{ Port string `mapstructure:"Port"`; Debug bool `mapstructure:"Debug"` }{Port: port}})
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
