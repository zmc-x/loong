package proxy

import (
	"loong/pkg/filters"
	"loong/pkg/global"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

const (
	Kind = "Proxy"

	ResultInternalError = "internalError"
	ResultServerError   = "serverError"
)

var (
	XForwardFor = http.CanonicalHeaderKey("X-Forwarded-For")

	kind = filters.Kind{
		Name: Kind,
		DefaultSpec: func() filters.Spec {
			return &Spec{}
		},
		CreateInstance: func(s filters.Spec) filters.Filter {
			return &HTTPProxy{
				spec: s.(*Spec),
			}
		},
	}
)

func init() {
	filters.Registy(&kind)
}

type Spec struct {
	Pool        []Host `json:"pool" validate:"required,dive,required"`
	LoadBalance `json:"loadBalance,omitempty"`
	HealthCheckInterval string `json:"healthCheckInterval,omitempty"`
	filters.BaseSpec
}

type Host struct {
	Url    string `json:"url" validate:"required,http_url"`
	Weight uint64  `json:"weight,omitempty" validate:"numeric"`
}

type LoadBalance struct {
	Policy string `json:"policy,omitempty" validate:"len=0|oneof=random roundRobin weightRoundRobin ipHash"`
}

type HTTPProxy struct {
	// hostMap Indicates the relationship
	// between the mapping path and the host service
	hostMap map[string]http.Handler
	lb      Balancer
	// alive mark host health status
	alive map[string]bool
	// done channel is used to notify the end of the service 
	done chan struct{}
	// health check interval
	interval time.Duration
	spec  *Spec
}

func (h *HTTPProxy) Init() error {
	return h.newHTTPProxy()
}

func (h *HTTPProxy) Handle(w http.ResponseWriter, r *http.Request) (string, int) {
	host, err := h.lb.Balance(GetClientIP(r))
	if err != nil {
		global.GlobalZapLog.Error("the back-end servers is unavaliable")
		return ResultInternalError, http.StatusServiceUnavailable
	}
	h.hostMap[host].ServeHTTP(w, r)
	// -1 indicates that the status code has been written.
	return "", -1
}

func (h *HTTPProxy) Close() {
	h.done <- struct{}{}
}

func (h *HTTPProxy) newHTTPProxy() error {
	// hosts mark all servers
	hosts := []Host{}
	hostMap := make(map[string]http.Handler)
	alive := make(map[string]bool)
	for _, target := range h.spec.Pool {
		url, err := url.Parse(target.Url)
		if err != nil {
			return err
		}
		proxy := httputil.NewSingleHostReverseProxy(url)
		host := GetHost(url)
		alive[host] = true
		hostMap[host] = proxy
		hosts = append(hosts, Host{
			Url:    host,
			Weight: target.Weight,
		})
	}

	lb, err := BuildBalancer(h.spec.Policy, hosts)
	if err != nil {
		return err
	}
	h.alive = alive
	h.lb = lb
	h.hostMap = hostMap
	h.done = make(chan struct{})
	
	if h.spec.HealthCheckInterval == "" {
		h.interval = DefaultInterval
	} else {
		interval, err := time.ParseDuration(h.spec.HealthCheckInterval)
		if err != nil {
			return err
		}
		h.interval = interval
	}
	h.HealthCheck()
	return nil
}
