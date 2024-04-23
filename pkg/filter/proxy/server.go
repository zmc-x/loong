package proxy

import (
	"loong/pkg/object/trafficgate"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var (
	XForwardFor = http.CanonicalHeaderKey("X-Forwarded-For")
)

type HTTPProxy struct {
	// hostMap Indicates the relationship 
	// between the mapping path and the host service
	hostMap map[string]http.Handler
	lb Balancer
	// alive mark host health status
	alive map[string]bool
}

func NewHTTPProxy(targetHosts []trafficgate.Host, algorithm string) (*HTTPProxy, error) {
	// hosts mark all servers
	hosts := []trafficgate.Host{}
	hostMap := make(map[string]http.Handler)
	alive := make(map[string]bool)
	for _, target := range targetHosts {
		url, err := url.Parse(target.Url)
		if err != nil {
			return nil, err 
		}
		proxy := httputil.NewSingleHostReverseProxy(url)
		host := GetHost(url)

		alive[host] = true 
		hostMap[host] = proxy
		hosts = append(hosts, trafficgate.Host{
			Url: host,
			Weight: target.Weight,
		})
	}
	
	lb, err := BuildBalancer(algorithm, hosts)
	if err != nil {
		return nil, err
	}
	return &HTTPProxy{
		alive: alive,
		hostMap: hostMap,
		lb: lb,
	}, nil
}

func (h *HTTPProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host, err := h.lb.Balance(GetClientIP(r))
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		return 
	}
	h.hostMap[host].ServeHTTP(w, r)
}