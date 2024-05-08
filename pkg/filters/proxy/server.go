package proxy

import (
	"loong/pkg/filters"
	"net/http"
	"net/http/httputil"
	"net/url"
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
	filters.BaseSpec
}

type Host struct {
	Url    string `json:"url" validate:"required,http_url"`
	Weight int64  `json:"weight,omitempty" validate:"numeric"`
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
	spec  *Spec
}

func (h *HTTPProxy) Init() error {
	return h.newHTTPProxy()
}

func (h *HTTPProxy) Handle(w http.ResponseWriter, r *http.Request) (string, int) {
	host, err := h.lb.Balance(GetClientIP(r))
	if err != nil {
		return ResultInternalError, http.StatusBadGateway
	}
	h.hostMap[host].ServeHTTP(w, r)
	// -1 indicates that the status code has been written.
	return "", -1
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
	return nil
}
