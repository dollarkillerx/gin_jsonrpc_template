package middleware

import (
	"errors"
	"runtime"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/google/feitian/pkg/common/resp"
)

func HttpRecover() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stackTrace := debug.Stack()
				runtime.Stack(stackTrace, true)
				log.Error().Msgf("HttpRecover url: %s stackTrace %s", ctx.Request.URL.Path, string(stackTrace))
				resp.ErrorReturn(ctx, "recover", errors.New("Internal Server Error"))
			}
		}()
		ctx.Next()
	}
}
