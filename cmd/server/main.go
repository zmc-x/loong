package main

import (
	"context"
	"errors"
	"fmt"

	"loong/pkg/controller"
	"loong/pkg/filter/proxy"
	"loong/pkg/global"
	"loong/pkg/logger"
	"loong/pkg/object/trafficgate"
	"net/http"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

// build the api gateway server
func main() {
	global.GlobalZapLog = logger.CreateLogger()
	defer global.GlobalZapLog.Sync()

	c, err := controller.ReadFromYaml()
	if err != nil {
		global.GlobalZapLog.Fatal("failed to read config", zap.String("error", err.Error()))
	}

	s := trafficgate.NewServer(c)

	for _, path := range c.Paths {
		handler, err := proxy.NewHTTPProxy(path.Pool, path.Policy)

		if err != nil {
			global.GlobalZapLog.Warn("newHTTPProxy error", zap.Any("error", err.Error()))
			continue
		}
		s.RegisterHandler(path.Path, handler)
		global.GlobalZapLog.Info("proxy is register", zap.Any("proxy", path))
	}
	// start the server
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go func() {
		global.GlobalZapLog.Info("server is starting", zap.String("address", fmt.Sprintf("%d", c.Port)))
		err = s.StartServer()
		if err != nil && errors.Is(err, http.ErrServerClosed) {
			global.GlobalZapLog.Error("failed to start server", zap.String("error", err.Error()))
			cancel()
		}
	}()

	// wait for the shutdown signal and stop the server
	<-ctx.Done()
	err = s.ShutdownServer()
	if err != nil {
		global.GlobalZapLog.Fatal("failed to shutdown server", zap.String("error", err.Error()))
	}
	global.GlobalZapLog.Info("server stopped")
}
