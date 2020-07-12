package secret

import (
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/knq/pemutil"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
)

const redactedMsg = "<REDACTED>"

var (
	generateDefaultECCurve          = elliptic.P256()
	generateDefaultRSABitSize       = 2048
	generateDefaultSymmetricKeySize = sha256.BlockSize
)

const (
	typeGenerate = "generate"
	typeInline   = "inline"
	typeEnv      = "env"
	typeFile     = "file"
)

// Source TODO
type Source interface{ Store() (pemutil.Store, error) }

func sourceToType(source Source) (string, error) {
	switch t := source.(type) {
	case nil:
		return "", nil
	case Generate, *Generate:
		return typeGenerate, nil
	case Inline, *Inline:
		return typeInline, nil
	case Env, *Env:
		return typeEnv, nil
	case File, *File:
		return typeFile, nil
	default:
		return "", fmt.Errorf("unknown source '%T'", t)
	}
}

func typeToSource(t string) (Source, error) {
	switch t {
	case typeGenerate:
		return new(Generate), nil
	case typeInline:
		return new(Inline), nil
	case typeEnv:
		return new(Env), nil
	case typeFile:
		return new(File), nil
	}
	return nil, fmt.Errorf("unknown type '%s'", t)
}

// Generate TODO
type Generate jwa.KeyType

// Store TODO
func (g Generate) Store() (pemutil.Store, error) {
	switch jwa.KeyType(g) {
	case jwa.EC:
		return pemutil.GenerateECKeySet(generateDefaultECCurve)
	case jwa.RSA:
		return pemutil.GenerateRSAKeySet(generateDefaultRSABitSize)
	case jwa.OctetSeq:
		return pemutil.GenerateSymmetricKeySet(generateDefaultSymmetricKeySize)
	}

	return nil, fmt.Errorf("generate: invalid key type '%s'", g)
}

// Inline TODO
type Inline []byte

// Store TODO
func (i Inline) Store() (pemutil.Store, error) {
	return pemutil.DecodeBytes([]byte(i))
}

func (i Inline) String() string { return redactedMsg }

// Env TODO
type Env string

// Store TODO
func (e Env) Store() (pemutil.Store, error) {
	str, ok := os.LookupEnv(string(e))
	if !ok {
		return nil, fmt.Errorf("env: variable '%s' not set", e)
	}
	return pemutil.DecodeBytes([]byte(str))
}

// File TODO
type File string

// Store TODO
func (f File) Store() (pemutil.Store, error) {
	bytes, err := ioutil.ReadFile(string(f))
	if err != nil {
		return nil, err
	}

	return pemutil.DecodeBytes(bytes)
}

// Config TODO
type Config struct {
	Algorithm jwa.SignatureAlgorithm
	Source
}

// DefaultConfig TODO
func DefaultConfig() Config {
	return Config{
		Algorithm: jwa.RS256,
		Source:    Generate(jwa.RSA),
	}
}

// PrivateKey TODO
func (c Config) PrivateKey() (jwk.Key, error) {
	pemStore, err := c.Store()
	if err != nil {
		return nil, err
	}

	return NewPrivate(c.Algorithm, pemStore)
}

// PublicKey TODO
func (c Config) PublicKey() (jwk.Key, error) {
	pemStore, err := c.Store()
	if err != nil {
		return nil, err
	}

	return NewPublic(c.Algorithm, pemStore)
}

// UnmarshalJSON TODO
func (c *Config) UnmarshalJSON(data []byte) error {
	var tmp struct {
		Algorithm string          `json:"algorithm"`
		Type      string          `json:"type"`
		Data      json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	if err := c.Algorithm.Accept(tmp.Algorithm); err != nil {
		return fmt.Errorf("%w: '%s'", err, tmp.Algorithm)
	}

	if tmp.Type == "" {
		if tmp.Data != nil {
			return errors.New("secret: missing type")
		}
		return nil
	}

	newSource, err := typeToSource(tmp.Type)
	if err != nil {
		return fmt.Errorf("secret: %w", err)
	}

	c.Source = newSource
	return json.Unmarshal(tmp.Data, c.Source)
}

// MarshalJSON TODO
func (c Config) MarshalJSON() ([]byte, error) {
	t, err := sourceToType(c.Source)
	if err != nil {
		return nil, fmt.Errorf("secret: %w", err)
	}

	tmp := struct {
		Algorithm string      `json:"algorithm,omitempty"`
		Type      string      `json:"type,omitempty"`
		Data      interface{} `json:"data,omitempty"`
	}{
		Algorithm: c.Algorithm.String(),
		Type:      t,
		Data:      c.Source,
	}

	return json.Marshal(tmp)
}
