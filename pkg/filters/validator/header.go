package validator

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
)

var (
	ErrNoMatchValue = errors.New("this header don't have match values")
)

type HeaderSpec map[string]ValueValidator

type ValueValidator struct {
	Values []string `json:"values"`
	Regexp string   `json:"regexp"`
	re     *regexp.Regexp
}

type HeaderValidtator struct {
	spec *HeaderSpec
}

func NewHeaderValidator(hs *HeaderSpec) (*HeaderValidtator, error) {
	v := &HeaderValidtator{spec: hs}

	for _, vv := range *hs {
		lenr, lenv := len(vv.Regexp), len(vv.Values)
		if lenr != 0 {
			re, err := regexp.Compile(vv.Regexp)
			if err != nil {
				return nil, fmt.Errorf("compile regexp %s failed", vv.Regexp)
			}
			vv.re = re
		} else if lenr == 0 && lenv == 0 {
			return nil, ErrNoMatchValue
		}
	}
	return v, nil
}

// validate the http request header
func (v *HeaderValidtator) Validate(r *http.Request) error {
	for key, vv := range *v.spec {
		if values, ok := r.Header[key]; !ok {
			return fmt.Errorf("header %s is invalid", key)
		} else {
			// match
			vis := false
			for _, v := range vv.Values {
				for _, reqHeaderValue := range values {
					if v == reqHeaderValue || (vv.re != nil && vv.re.MatchString(reqHeaderValue)) {
						vis = true
						break
					}
				}
				if vis {
					break
				}
			}
			if !vis {
				return fmt.Errorf("header %s is invalid", key)
			}
		}
	}
	return nil
}