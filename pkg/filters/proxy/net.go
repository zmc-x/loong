package proxy

import (
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	ConnectTimeOut = 3 * time.Second
	// defaultHealthCheckInterval
	DefaultInterval = 5 * time.Second
)

// GetHost function return url.Host
func GetHost(url *url.URL) string {
	if _, _, err := net.SplitHostPort(url.Host); err == nil {
		return url.Host
	}
	switch url.Scheme {
	case "http":
		return url.Host + ":80"
	case "https":
		return url.Host + ":443"
	}
	return url.Host
}

func GetClientIP(r *http.Request) string {
	client, _, _ := net.SplitHostPort(r.RemoteAddr)
	xff := r.Header.Get(XForwardFor)
	if len(xff) != 0 {
		pos := strings.Index(xff, ",")
		// don't found
		if pos == -1 {
			pos = len(xff)
		}
		client = xff[:pos]
	}
	return client
}

// this function return backend status
func IsBackendAlive(host string) bool {
	co, err := net.DialTimeout("tcp", host, ConnectTimeOut)
	if err != nil {
		return false
	}
	co.Close()
	return true
}
