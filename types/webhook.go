package types

import (
	"github.com/go-jose/go-jose/v4"
)

// StonePublicKeys holds JWK keys by kid (RFC 7517)
type StonePublicKeys map[string]*jose.JSONWebKey

func (s StonePublicKeys) Get(key string) *jose.JSONWebKey {
	return s[key]
}
