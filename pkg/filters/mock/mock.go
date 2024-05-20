package mock

import (
	"loong/pkg/filters"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	Kind = "Mock"
	result = "mocked"
)

var (
	kind = filters.Kind{
		Name: Kind,
		DefaultSpec: func() filters.Spec {
			return &MockSpec{}
		},
		CreateInstance: func(s filters.Spec) filters.Filter {
			return &Mock{
				spec: s.(*MockSpec),
			}
		},
	}
)

func init() {
	filters.Registy(&kind)
}

type MockSpec struct {
	filters.BaseSpec
	// match rules
	Rules []*Rule `json:"rules" validate:"required"`
}

type Rule struct {
	// match rule
	Match MatchRule `json:",inline"`
	// response header info
	Header map[string]string `json:"header,omitempty"`
	Code int `json:"code" validate:"gte=100,lt=600"`
	// response data
	Body string `json:"body,omitempty"`
	// process duration
	Delay string `json:"delay,omitempty"`

	delay time.Duration
}

type MatchRule struct {
	Path           string                  `json:"path,omitempty" validate:"uri"`
	PathPrefix     string                  `json:"pathPrefix,omitempty" validate:"uri"`
	Headers        map[string]*HeaderMatch `json:"headers,omitempty"`
	MatchAllHeader bool                    `json:"matchAllHeaders,omitempty" validate:"boolean"`
}

type HeaderMatch struct {
	Exact  string `json:"exact,omitempty"`
	Prefix string `json:"prefix,omitempty"`
	Regexp string `json:"regexp,omitempty"`
	Empty  bool   `json:"empty,omitempty" validate:"boolean"`
	re     *regexp.Regexp
}

func (h *HeaderMatch) Init() {
	if h.Regexp != "" {
		h.re = regexp.MustCompile(h.Regexp)
	}
}

func (h *HeaderMatch) Match(value string) bool {
	if h.Empty && value == "" {
		return true
	}

	if h.Exact != "" && h.Exact == value {
		return true 
	}

	if h.Prefix != "" && strings.HasPrefix(value, h.Prefix) {
		return true 
	}
	
	if h.re == nil {
		return false 
	}
	return h.re.MatchString(value)
}

type Mock struct {
	spec *MockSpec
}

func (m *Mock) Init() error {
	for _, rule := range m.spec.Rules {
		if rule.Delay == "" {
			continue 
		}
		delay, err := time.ParseDuration(rule.Delay)
		if err != nil {
			return err 
		}
		rule.delay = delay
	}
	return nil
}

func (m *Mock) Handle(w http.ResponseWriter, r *http.Request) (string, int) {
	if idx, ok := m.Match(r); ok {
		rule := m.spec.Rules[idx]
		wait := time.After(rule.delay)
		<- wait
		// imitate backend process
		for k, v := range rule.Header {
			w.Header().Set(k, v)
		}
		w.WriteHeader(rule.Code)
		w.Write([]byte(rule.Body))
		return result, -1
	}
	return "", http.StatusBadRequest
}

func (m *Mock) Close() {}

func (m *Mock) Match(r *http.Request) (int, bool) {
	// determine whether the request meets the requirment
	matchPath := func(rule *Rule) bool {
		if rule.Match.Path == "" && rule.Match.PathPrefix == "" {
			return true 
		}
		u := r.URL
		if rule.Match.Path == u.Path {
			return true
		}

		if rule.Match.PathPrefix == "" {
			return false 
		}
		return strings.HasPrefix(u.Path, rule.Match.PathPrefix)
	}

	matchHeader := func(key string, rule *Rule) bool {
		m := rule.Match.Headers[key]
		m.Init()
		for _, str := range r.Header[key] {
			res := m.Match(str)
			if res {
				return true 
			}
		}
		return false
	}

	matchAllHeader := func(rule *Rule) bool {
		for k := range rule.Match.Headers {
			res := matchHeader(k, rule) 
			if res && !rule.Match.MatchAllHeader {
				return true 
			} else if rule.Match.MatchAllHeader && !res {
				return false
			}
		}
		return true
	}

	// start to match
	for idx, rule := range m.spec.Rules {
		// there only needs to be one match
		if matchPath(rule) && matchAllHeader(rule) {
			return idx, true
		}
	}

	return -1, false
}