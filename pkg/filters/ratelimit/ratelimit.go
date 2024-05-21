package ratelimit

import (
	"context"
	"fmt"
	"loong/pkg/filters"
	"loong/pkg/utils/stringtools"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

const (
	Kind = "RateLimiter"
	result = "ratelimited"
	DefaultBurst = 500
	DefaultRPS = 500
	DefaultTimeoutDuration = 100 * time.Millisecond
)

var kind = filters.Kind{
	Name: Kind,
	DefaultSpec: func() filters.Spec {
		return &RateLimitSpec{}
	},
	CreateInstance: func(s filters.Spec) filters.Filter {
		return &RateLimiter{
			spec: s.(*RateLimitSpec),
		}
	},
}

func init() {
	filters.Registy(&kind)
}

type RateLimitSpec struct {
	Policies []*Policy `json:"policies" validate:"required,dive,required"`
	filters.BaseSpec
	DefaultPolicy string `json:"defaultPolicy" validate:"required"`
	URLs          []*URL `json:"urls" validate:"dive"`
}

type Policy struct {
	Name            string `json:"name" validate:"required"`
	TimeoutDuration string `json:"timeoutDuration"`
	BurstSize       int `json:"burstSize" validate:"gt=-1,numeric"`
	RPS             float64 `json:"rps" validate:"gt=-1,numeric"`

	timeoutDuration time.Duration
	limit *rate.Limiter
}

type URL struct {
	Methods   []string `json:"methods" validate:"dive,oneof=GET POST HEAD DELEAT PUT PATCH CONNECT OPTIONS TRACE"`
	Url       stringtools.Matcher `json:"url"`
	PolicyRef string `json:"policyRef"`
}

type RateLimiter struct {
	spec *RateLimitSpec
	policyMap map[string]*Policy
}

func (r *RateLimiter) Init() error {
	r.policyMap = make(map[string]*Policy)
	for _, policy := range r.spec.Policies {
		// check whether the policy name is valid
		if r.policyMap[policy.Name] != nil {
			return fmt.Errorf("the policy name is duplicated")
		}
		t, err := time.ParseDuration(policy.TimeoutDuration)
		if err != nil {
			t = DefaultTimeoutDuration
		}
		policy.timeoutDuration = t
		if policy.RPS == 0 {
			policy.RPS = DefaultRPS
		}
		if policy.BurstSize == 0 {
			policy.BurstSize = DefaultBurst
		}
		policy.limit = rate.NewLimiter(rate.Limit(policy.RPS), policy.BurstSize)
		r.policyMap[policy.Name] = policy
	}
	return nil
}

func (r *RateLimiter) Handle(w http.ResponseWriter, req *http.Request) (string, int) {
	// start ratelimit process
	reqMethod := req.Method
	var limit *rate.Limiter
	var timeout time.Duration
	for _, url := range r.spec.URLs {
		// match methods
		if url.PolicyRef == "" && r.spec.DefaultPolicy == "" {continue}
		vis := false
		if len(url.Methods) == 0 {
			vis = true
		}
		for _, method := range url.Methods {
			if method == reqMethod {
				vis = true
			}
		}
		url.Url.Init()
		vis = vis && url.Url.Match(req.URL.Path)
		if vis {
			policy := r.policyMap[url.PolicyRef]
			if policy == nil {
				policy = r.policyMap[r.spec.DefaultPolicy]
			}
			limit, timeout = policy.limit, policy.timeoutDuration
			break
		}
	}
	if limit != nil {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if err := limit.Wait(ctx); err != nil {
			w.WriteHeader(http.StatusTooManyRequests)
			return result, -1
		} 
	}
	return "", http.StatusOK
}

func (r *RateLimiter) Close() {}