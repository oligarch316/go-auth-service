package token

import (
	"strconv"
	"time"

	"github.com/oligarch316/go-skeleton/pkg/config/types"
)

// StandardClaims TODO
type StandardClaims struct {
	Issuer  string `json:"iss,omitempty"`
	Subject string `json:"sub,omitempty"`
	TokenID string `json:"jti,omitempty"`

	Audience ctype.StringSet `json:"aud,omitempty"`

	Expiration *NumericDate `json:"exp,omitempty"`
	NotBefore  *NumericDate `json:"nbf,omitempty"`
	IssuedAt   *NumericDate `json:"iat,omitempty"`
}

// NumericDate TODO
type NumericDate struct{ time.Time }

// MarshalJSON TODO
func (nd NumericDate) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(nd.Unix(), 10)), nil
}

// UnmarshalJSON TODO
func (nd *NumericDate) UnmarshalJSON(data []byte) error {
	x, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return err
	}

	nd.Time = time.Unix(x, 0)
	return nil
}
