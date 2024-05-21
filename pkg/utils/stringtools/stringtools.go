package stringtools

import (
	"regexp"
	"strings"
)


type Matcher struct {
	Exact  string `json:"exact,omitempty"`
	Prefix string `json:"prefix,omitempty"`
	Regexp string `json:"regexp,omitempty"`
	Empty  bool   `json:"empty,omitempty" validate:"boolean"`
	re     *regexp.Regexp
}

func (h *Matcher) Init() {
	if h.Regexp != "" {
		h.re = regexp.MustCompile(h.Regexp)
	}
}

func (h *Matcher) Match(value string) bool {
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