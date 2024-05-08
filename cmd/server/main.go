package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"loong/pkg/controller"
	"loong/pkg/global"
	"loong/pkg/logger"
	"loong/pkg/object/pipeline"
	"loong/pkg/object/trafficgate"
	_ "loong/pkg/register"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// build the api gateway server
func main() {
	controller.DirPath, _ = os.Getwd()
	// init global variable
	global.GlobalZapLog = logger.CreateLogger()
	defer global.GlobalZapLog.Sync()

	global.GlobalValidator = validator.New(validator.WithRequiredStructEnabled())

	serverCfg, err := controller.ReadFromYaml("trafficGate", "server.yml")
	if err != nil {
		global.GlobalZapLog.Fatal("failed to read config", zap.String("error", err.Error()))
	}
	s, err := trafficgate.NewServer(serverCfg)
	if err != nil {
		global.GlobalZapLog.Fatal("failed to new Server", zap.String("error", err.Error()))
	}


	pipelineDir, _ := os.ReadDir(filepath.Join(controller.DirPath, "temp/pipeline"))
	for _, v := range pipelineDir {
		pipelineCfg, err := controller.ReadFromYaml("pipeline", v.Name())
		if err != nil {
			global.GlobalZapLog.Fatal("failed to read config", zap.String("error", err.Error()))
		}
		_, err = pipeline.InitPipeline(pipelineCfg)
		if err != nil {
			global.GlobalZapLog.Fatal("failed to new Pipeline", zap.String("error", err.Error()))
		}
	}
	
	for _, path := range s.GetPath() {
		s.RegisterHandler(path.Path, pipeline.PipelineMap[path.Backend])
	}
	s.RegisterMiddleWare()
	// start the server
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go func() {
		global.GlobalZapLog.Info("server is starting", zap.String("address", fmt.Sprintf("%d", s.GetPort())))
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
