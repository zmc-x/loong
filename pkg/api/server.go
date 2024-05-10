package api

import (
	"loong/cmd"
	"loong/pkg/global"
	"net/http"
)

func RegisterRouter() {
	global.GlobalRouter.HandleFunc("/configurations", handle)
}

// this function reloads the configuration file
func handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPatch:
		go cmd.ReloadConfig()
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
