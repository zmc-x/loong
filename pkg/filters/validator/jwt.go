package validator

import (
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type JWTSpec struct {
	Algorithm string `json:"algorithm" validate:"required,oneof=HS256 HS384 HS512 RS256 RS384 RS512 ES256 ES384 ES512 PS256 PS384 PS512"`
	// The PK is hex encoding
	PublicKey string `json:"publickey" validate:"len=0|hexadecimal"`
	// The Secret is hex encoding
	Secret string `json:"secret" validate:"hexadecimal|len=0"`
	CookieName string `json:"cookieName,omitempty"`
}


type JWTValidator struct {
	spec *JWTSpec
	key any
}

func NewJWTValidator(spec *JWTSpec) (*JWTValidator, error) {
	var key any
	if len(spec.PublicKey) > 0 {
		pk, _ := hex.DecodeString(spec.PublicKey)
		p, _ := pem.Decode(pk)
		key, _ = x509.ParsePKIXPublicKey(p.Bytes)
	} else {
		key, _ = hex.DecodeString(spec.Secret)
	}
	return &JWTValidator{spec, key}, nil
}

// Validate validates the JWT token of a http request
func (v *JWTValidator) Validate(r *http.Request) error {
	var token string
	
	if v.spec.CookieName != "" {
		if cookie, e := r.Cookie(v.spec.CookieName); e == nil {
			token = cookie.Value
		}
	}

	if token == "" {
		const prefix = "Bearer "
		authHdr := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHdr, prefix) {
			return fmt.Errorf("unexpected authorization header: %s", authHdr)
		}
		token = authHdr[len(prefix):]
	}
	// jwt.Parse does everything including parsing and verification
	t, e := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if alg := token.Method.Alg(); alg != v.spec.Algorithm {
			return nil, fmt.Errorf("unexpected signing method: %v", alg)
		}
		return v.key, nil
	})
	if e != nil {
		return e
	}
	if !t.Valid {
		return fmt.Errorf("invalid jwt token")
	}
	return nil
}
