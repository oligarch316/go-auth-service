package secret

import (
	"crypto"
	"encoding/base64"
	"errors"

	"github.com/knq/pemutil"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
)

const thumbprintHash = crypto.SHA256

// NewPrivate TODO
func NewPrivate(alg jwa.SignatureAlgorithm, store pemutil.Store) (jwk.Key, error) {
	res, err := newKey(alg, store, blockIsPrivate)
	if err != nil {
		return nil, err
	}

	res.Set(jwk.KeyOpsKey, jwk.KeyOperationList{jwk.KeyOpSign, jwk.KeyOpVerify})
	return res, nil
}

// NewPublic TODO
func NewPublic(alg jwa.SignatureAlgorithm, store pemutil.Store) (jwk.Key, error) {
	// TODO: Not so great that we're mutating an argument that's not soley ours...
	store.AddPublicKeys()

	res, err := newKey(alg, store, blockIsPublic)
	if err != nil {
		return nil, err
	}

	res.Set(jwk.KeyOpsKey, jwk.KeyOperationList{jwk.KeyOpVerify})

	return res, nil
}

// SetMeta TODO
func SetMeta(key jwk.Key, alg, kid, use string) {
	key.Set(jwk.AlgorithmKey, alg)
	key.Set(jwk.KeyIDKey, kid)
	key.Set(jwk.KeyUsageKey, use)
}

// IsPrivate TODO
func IsPrivate(key jwk.Key) bool {
	// Be explicit, no !IsPublic here

	switch key.(type) {
	case jwk.ECDSAPrivateKey, jwk.RSAPrivateKey, jwk.SymmetricKey:
		return true
	}
	return false
}

// IsPublic TODO
func IsPublic(key jwk.Key) bool {
	// Be explicit, no !IsPrivate here

	switch key.(type) {
	case jwk.ECDSAPublicKey, jwk.RSAPublicKey:
		return true
	}
	return false
}

func newKey(alg jwa.SignatureAlgorithm, store pemutil.Store, filter func(pemutil.BlockType) bool) (jwk.Key, error) {
	var key interface{}

	for blockType, blockKey := range store {
		if filter(blockType) {
			if key != nil {
				return nil, errors.New("multiple keys found in pem store")
			}

			key = blockKey
		}
	}

	if key == nil {
		return nil, errors.New("no valid keys found in pem store")
	}

	res, err := jwk.New(key)
	if err != nil {
		return nil, err
	}

	thumbprint, err := res.Thumbprint(thumbprintHash)
	if err != nil {
		return nil, err
	}

	SetMeta(res, alg.String(), base64.RawURLEncoding.EncodeToString(thumbprint), string(jwk.ForSignature))

	// TODO: handle certificate chains

	return res, nil
}

func blockIsPrivate(blockType pemutil.BlockType) bool {
	return blockType == pemutil.PrivateKey ||
		blockType == pemutil.RSAPrivateKey ||
		blockType == pemutil.ECPrivateKey
}

func blockIsPublic(blockType pemutil.BlockType) bool {
	return blockType == pemutil.PublicKey
}
