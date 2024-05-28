package validator

import (
	"loong/pkg/filters"
	"net/http"
)

const (
	Kind = "Validator"

	ResultInvalid = "invalid"
)

var (
	kind = filters.Kind{
		Name:        Kind,
		DefaultSpec: func() filters.Spec { return &Spec{} },
		CreateInstance: func(s filters.Spec) filters.Filter {
			return &Validator{
				spec: s.(*Spec),
			}
		},
	}
)

func init() {
	filters.Registy(&kind)
}

type Spec struct {
	filters.BaseSpec
	Headers *HeaderSpec `json:"headers,omitempty"`
	JWT     *JWTSpec    `json:"jwt,omitempty"`
}

type Validator struct {
	spec *Spec
	// header part
	headers *HeaderValidtator
	// jwt part
	jwt *JWTValidator
}

func (v *Validator) Init() error {
	if v.spec.Headers != nil {
		headers, err := NewHeaderValidator(v.spec.Headers)
		if err != nil {
			return err
		}
		v.headers = headers
	}
	if v.spec.JWT != nil {
		jwt, err := NewJWTValidator(v.spec.JWT)
		if err != nil {
			return err
		}
		v.jwt = jwt
	}
	return nil
}

func (v *Validator) Handle(w http.ResponseWriter, r *http.Request) (string, int) {
	if v.headers != nil {
		if err := v.headers.Validate(r); err != nil {
			return ResultInvalid, http.StatusBadRequest
		}
	}
	if v.jwt != nil {
		if err := v.jwt.Validate(r); err != nil {
			return ResultInvalid, http.StatusBadRequest
		}
	}
	return "", http.StatusOK
}

func (v *Validator) Close() {}
