package main

import (
	"context"
	"flag"
	"fmt"
	"loong/cmd"
	"loong/pkg/api"
	"loong/pkg/controller"
	"loong/pkg/global"
	"loong/pkg/logger"
	_ "loong/pkg/register"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

func main() {
	controller.DirPath, _ = os.Getwd()
	// init global variable
	global.GlobalZapLog = logger.CreateLogger()
	defer global.GlobalZapLog.Sync()
	global.GlobalValidator = validator.New(validator.WithRequiredStructEnabled())
	global.GlobalRouter = mux.NewRouter()

	flag.Uint64Var(&global.Port, "port", 2459, "port Indicates the location where the loong api gateway provides interface services to the client")
	flag.Parse()

	if global.Port >= 1 << 16 {
		global.GlobalZapLog.Fatal("the port value is greater than 2 ^ 16 - 1")
	}

	go func() {
		api.RegisterRouter()
		http.ListenAndServe(fmt.Sprintf(":%d", global.Port), global.GlobalRouter)
	}()
	
	global.GlobalZapLog.Info("loong API Gateway is running")
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	go cmd.ReloadConfig()

	<- ctx.Done()
	global.GlobalZapLog.Info("loong API Gateway is stoped")
}