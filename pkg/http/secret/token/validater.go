package token

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/oligarch316/go-skeleton/pkg/config/types"
	"go.uber.org/zap/zapcore"
)

// ConfigValidater TODO.
type ConfigValidater struct {
	AllowedIssuers ctype.StringSet `json:"allowedIssuers"`
	AudienceName   string          `json:"audienceName"`
}

// MarshalLogObject TODO.
func (cv ConfigValidater) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("audienceName", cv.AudienceName)
	return enc.AddArray("allowedIssuers", cv.AllowedIssuers)
}

// Validater TODO.
type Validater struct {
	Claims ConfigValidater
	Secret interface {
		Validate(token string, claims interface{}) error
	}
}

// Validate TODO.
func (v Validater) Validate(r *http.Request) (string, error) {
	tokenStr, err := loadTokenString(r)
	if err != nil {
		return "", err
	}

	var claims StandardClaims
	if err := v.Secret.Validate(tokenStr, &claims); err != nil {
		return "", err
	}

	now := time.Now()

	switch {
	// ---- Required claim names
	case claims.Subject == "":
		return "", errors.New("sub: missing subject")
	case claims.Issuer == "":
		return "", errors.New("iss: missing issuer")
	case claims.Audience == nil:
		return "", errors.New("aud: missing audience")
	case claims.Expiration == nil:
		return "", errors.New("exp: missing expiration")

		// ----- Issuer/Audience allowances
	case !v.Claims.AllowedIssuers.Contains(claims.Issuer):
		return "", errors.New("iss: not from an allowed issuer")
	case !claims.Audience.Contains(v.Claims.AudienceName):
		return "", errors.New("aud: not for this audience")

		// ----- Time requirements
	case now.After(claims.Expiration.Time):
		return "", errors.New("exp: expired")
	case claims.NotBefore != nil && now.Before(claims.NotBefore.Time):
		return "", errors.New("nbf: not yet valid")
	}

	return claims.Subject, nil
}

func loadTokenString(r *http.Request) (string, error) {
	hdrVal := r.Header.Get("Authorization")
	if hdrVal == "" {
		return "", errors.New("empty authorization header")
	}

	fields := strings.Fields(hdrVal)
	switch {
	case len(fields) != 2:
		return "", errors.New("invalid authorization header format")
	case fields[0] != "Bearer":
		return "", fmt.Errorf("invalid authorization header authentication scheme: %s", fields[0])
	}

	return fields[1], nil
}
