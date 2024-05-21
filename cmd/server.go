package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"loong/pkg/controller"
	"loong/pkg/global"
	"loong/pkg/object/pipeline"
	"loong/pkg/object/trafficgate"
	"net/http"

	"go.uber.org/zap"
)

var (
	cancel context.CancelFunc
	// mutex to prevent data race
	mutex sync.Mutex
)


// ReloadConfig function can reloads the configurations
func ReloadConfig() {
	if cancel != nil {
		cancel()
	}
	mutex.Lock()
	defer mutex.Unlock()
	resetConfig()
	startoongApiGateway()
}

// resetConfig function can reset all configuration of loong
func resetConfig() {
	trafficgate.Reset()
	pipeline.Reset()
}


// build the api gateway server
func startoongApiGateway() {
	// read the trafficGate configuration file
	trafficDir, _ := os.ReadDir(filepath.Join(controller.DirPath, "temp/trafficgate"))
	for _, v := range trafficDir {
		fileType := filepath.Ext(v.Name())
		if fileType != `.yaml` && fileType != `.yml` {continue}
		trafficCfg, err := controller.ReadFromYaml("trafficGate", v.Name())
		if err != nil {
			global.GlobalZapLog.Error("failed to read config", zap.String("error", err.Error()))
			continue
		}
		_, err = trafficgate.NewServer(trafficCfg)
		if err != nil {
			global.GlobalZapLog.Error("failed to new TrafficGate", zap.String("error", err.Error()))
			continue
		}
	}

	// read the pipeline configuration file
	pipelineDir, _ := os.ReadDir(filepath.Join(controller.DirPath, "temp/pipeline"))
	for _, v := range pipelineDir {
		fileType := filepath.Ext(v.Name())
		if fileType != `.yaml` && fileType != `.yml` {continue}
		pipelineCfg, err := controller.ReadFromYaml("pipeline", v.Name())
		if err != nil {
			global.GlobalZapLog.Error("failed to read config", zap.String("error", err.Error()))
			continue
		}
		_, err = pipeline.InitPipeline(pipelineCfg)
		if err != nil {
			global.GlobalZapLog.Fatal("failed to new Pipeline", zap.String("error", err.Error()))
			continue
		}
	}

	var ctx context.Context
	// start the server
	ctx, cancel = context.WithCancel(context.Background())
	// register handle
	for _, server := range trafficgate.Servers {
		for _, path := range server.GetPath() {
			server.RegisterHandler(path.Path, pipeline.PipelineMap[path.Backend])
		}
		server.RegisterMiddleWare()
		go runServer(server)
	}

	// wait for the shutdown signal and stop the server
	<-ctx.Done()

	for _, server := range trafficgate.Servers {
		err := server.ShutdownServer()
		if err != nil {
			// https://pkg.go.dev/net/http#Server.Shutdown
			global.GlobalZapLog.Error("failed to shutdown traffic server", zap.String("error", err.Error()))
		}
	}
	global.GlobalZapLog.Info("traffic servers stopped")
}

func runServer(server *trafficgate.Server) {
	global.GlobalZapLog.Info("traffic server " + server.GetName() + " is starting", zap.String("address", fmt.Sprintf("%d", server.GetPort())))
	err := server.StartServer()
	if err != nil && errors.Is(err, http.ErrServerClosed) {
		global.GlobalZapLog.Error("failed to start traffic server", zap.String("error", err.Error()))
		return 
	}
}