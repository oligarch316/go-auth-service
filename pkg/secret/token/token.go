package token

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/oligarch316/go-auth-service/pkg/secret"
)

const jwsType = "JWT"

func unpackKey(key jwk.Key) (interface{}, error) {
	var unpacked interface{}
	if err := key.Raw(&unpacked); err != nil {
		return nil, err
	}
	return unpacked, nil
}

func signKeyToValKey(key jwk.Key) (res jwk.Key, err error) {
	switch t := key.(type) {
	case jwk.SymmetricKey:
		return key, nil
	case jwk.ECDSAPrivateKey:
		res, err = t.PublicKey()
	case jwk.RSAPrivateKey:
		res, err = t.PublicKey()
	default:
		err = fmt.Errorf("unknown key type: %T", t)
	}

	if err != nil {
		return
	}

	secret.SetMeta(res, key.Algorithm(), key.KeyID(), key.KeyUsage())
	return
}

// Validater TODO
type Validater struct {
	headers jws.Headers
	valKey  interface{}
}

// NewValidater TODO
func NewValidater(key jwk.Key) (*Validater, error) {
	if !secret.IsPublic(key) {
		return nil, errors.New("key is not public")
	}

	return newValidater(key)
}

func newValidater(key jwk.Key) (*Validater, error) {
	valKey, err := unpackKey(key)
	if err != nil {
		return nil, err
	}

	headers := jws.NewHeaders()
	headers.Set(jws.AlgorithmKey, key.Algorithm())
	headers.Set(jws.KeyIDKey, key.KeyID())
	headers.Set(jws.TypeKey, jwsType)

	return &Validater{valKey: valKey, headers: headers}, nil
}

// Validate TODO
func (v Validater) Validate(token string, claims interface{}) error {
	payload, err := jws.Verify([]byte(token), v.headers.Algorithm(), v.valKey)
	if err != nil {
		return err
	}

	return json.Unmarshal(payload, claims)
}

// Signer TODO
type Signer struct {
	signKey interface{}
	v       Validater // Don't embed to prevent exposing a "Validater" backed by a private key
}

// NewSigner TODO
func NewSigner(key jwk.Key) (*Signer, error) {
	if !secret.IsPrivate(key) {
		return nil, errors.New("key is not private")
	}

	valKey, err := signKeyToValKey(key)
	if err != nil {
		return nil, err
	}

	v, err := newValidater(valKey)
	if err != nil {
		return nil, err
	}

	signKey, err := unpackKey(key)
	if err != nil {
		return nil, err
	}

	return &Signer{signKey: signKey, v: *v}, nil
}

// Sign TODO
func (s Signer) Sign(claims interface{}) (string, error) {
	data, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	res, err := jws.Sign(data, s.v.headers.Algorithm(), s.signKey, jws.WithHeaders(s.v.headers))
	return string(res), err
}

// Validate TODO
func (s Signer) Validate(token string, claims interface{}) error {
	return s.v.Validate(token, claims)
}
