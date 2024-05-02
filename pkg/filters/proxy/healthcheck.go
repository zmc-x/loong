package proxy

import (
	"loong/pkg/global"
	"time"

	"go.uber.org/zap"
)

func (p *HTTPProxy) HealthCheck(interval uint64) {
	for host := range p.hostMap {
		go p.healthCheck(host, interval)
	}
}

func (p *HTTPProxy) healthCheck(host string, interval uint64) {
	ticker := time.NewTicker(time.Second * time.Duration(interval))
	for range ticker.C {
		if IsBackendAlive(host) && !p.alive[host] {
			global.GlobalZapLog.Info("this server is unreachable", zap.Any("server address", host))
			p.alive[host] = true
			p.lb.Add(host)
		} else if !IsBackendAlive(host) && p.alive[host] {
			global.GlobalZapLog.Info("this server is reachable", zap.Any("server address", host))
			p.alive[host] = false
			p.lb.Remove(host)
		}
	}
}
