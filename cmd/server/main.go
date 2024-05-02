package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"loong/pkg/controller"
	"loong/pkg/global"
	"loong/pkg/logger"
	"loong/pkg/object/pipeline"
	"loong/pkg/object/trafficgate"
	_ "loong/pkg/register"
	"net/http"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

// build the api gateway server
func main() {
	global.GlobalZapLog = logger.CreateLogger()
	defer global.GlobalZapLog.Sync()

	serverCfg, err := controller.ReadFromYaml("trafficGate")
	if err != nil {
		global.GlobalZapLog.Fatal("failed to read config", zap.String("error", err.Error()))
	}
	pipelineCfg, err := controller.ReadFromYaml("pipeline")
	if err != nil {
		global.GlobalZapLog.Fatal("failed to read config", zap.String("error", err.Error()))
	}

	s, err := trafficgate.NewServer(serverCfg)
	if err != nil {
		global.GlobalZapLog.Fatal("failed to new Server", zap.String("error", err.Error()))
	}
	p, err := pipeline.InitPipeline(pipelineCfg)
	if err != nil {
		global.GlobalZapLog.Fatal("failed to new Pipeline", zap.String("error", err.Error()))
	}
	
	cfg := &trafficgate.Config{}
	json.Unmarshal(serverCfg.([]byte), cfg)
	for _, path := range cfg.Paths {
		s.RegisterHandler(path.Path, p)
	}
	// start the server
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go func() {
		global.GlobalZapLog.Info("server is starting", zap.String("address", fmt.Sprintf("%d", cfg.Port)))
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
