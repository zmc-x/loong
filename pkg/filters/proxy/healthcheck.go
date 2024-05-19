package proxy

import (
	"loong/pkg/global"
	"time"

	"go.uber.org/zap"
)

func (p *HTTPProxy) HealthCheck() {
	for host := range p.hostMap {
		go p.healthCheck(host)
	}
}

func (p *HTTPProxy) healthCheck(host string) {
	ticker := time.NewTicker(time.Second * time.Duration(p.spec.HealthCheckInterval))

	for {
		select {
		case <-p.done:
			global.GlobalZapLog.Info("health check is stoped", zap.Any("backend name", p.spec.Name()))
			return
		case <-ticker.C:
			newAlive := IsBackendAlive(host)
			if newAlive && !p.alive[host] {
				global.GlobalZapLog.Info("this server is reachable", zap.Any("server address", host))
				p.alive[host] = true
				p.lb.Add(host)
			} else if !newAlive && p.alive[host] {
				global.GlobalZapLog.Info("this server is unreachable", zap.Any("server address", host))
				p.alive[host] = false
				p.lb.Remove(host)
			} else if newAlive && p.alive[host] {
				global.GlobalZapLog.Info("this server is healthy", zap.Any("server address", host))
			} else if !newAlive && !p.alive[host] {
				global.GlobalZapLog.Info("this server is not healthy", zap.Any("server address", host))
			}
		}
	}
}
