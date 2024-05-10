package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

var (
	host string 
	port uint64
	reload bool
)

func main() {
	flag.BoolVar(&reload, "reload", false, "reload is true and causes the server to reload the configuration")
	flag.StringVar(&host, "host", "http://127.0.0.1", "host is the loong api gateway server address, e.g. http://127.0.0.1")
	flag.Uint64Var(&port, "port", 2459, "port is the loong api gateway server port")
	flag.Parse()

	
	if reload {
		if port >= 1 << 16 {
			log.Fatal("your inputed port is invalid!")
		}
		u, err := url.Parse(host)
		if err != nil || u.Scheme != "http" && u.Scheme != "https" || u.Host == "" {
			log.Fatal("your inputed host is invalid!")
		}

		req, err := http.NewRequest(http.MethodPatch, u.String() + fmt.Sprintf(":%d", port), nil)
		if err != nil {
			log.Fatal(err)
		}
		_, err = http.DefaultClient.Do(req)
		if err != nil {
			log.Fatal(err)
		}
	}
}