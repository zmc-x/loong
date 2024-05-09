package api

import (
	"loong/pkg/global"
	"net/http"
)

func RegisterRouter() {
	global.GlobalRouter.HandleFunc("/configurations", handle)
}

func handle(w http.ResponseWriter, r *http.Request) {
	
}
