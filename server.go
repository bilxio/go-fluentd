package concator

import (
	"context"
	"net/http"
	"time"

	middlewares "github.com/Laisky/gin-middlewares"
	"github.com/Laisky/go-fluentd/libs"
	utils "github.com/Laisky/go-utils"
	"github.com/Laisky/zap"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

var (
	server                   = gin.New()
	defaultGraceShutdownWait = 3 * time.Second
)

// RunServer starting http server
func RunServer(ctx context.Context, addr string) {
	if !utils.Settings.GetBool("debug") {
		gin.SetMode(gin.ReleaseMode)
	}

	httpSrv := http.Server{
		Addr:    addr,
		Handler: server,
	}

	server.Use(gin.Recovery())
	server.Any("/health", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "hello, world")
	})

	// supported action:
	// cmdline, profile, symbol, goroutine, heap, threadcreate, block
	pprof.Register(server, "pprof")
	middlewares.BindPrometheus(server)

	libs.Logger.Info("listening on http", zap.String("addr", addr))
	go func() {
		libs.Logger.Panic("server exit", zap.Error(httpSrv.ListenAndServe()))
	}()

	<-ctx.Done()
	srvCtx, cancel := context.WithTimeout(ctx, defaultGraceShutdownWait)
	defer cancel()
	if err := httpSrv.Shutdown(srvCtx); err != nil {
		libs.Logger.Error("shutdown monitor server", zap.Error(err))
	}

}
