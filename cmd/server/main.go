package main

import (
	"loong/cmd"
	"loong/pkg/controller"
	"loong/pkg/global"
	"loong/pkg/logger"
	_ "loong/pkg/register"
	"os"

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
	cmd.ReloadConfig()
}