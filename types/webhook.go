package types

import "gopkg.in/square/go-jose.v2"

// StonePublicKeys holds JWK keys by kid (RFC 7517)
type StonePublicKeys map[string]*jose.JSONWebKey

func (s StonePublicKeys) Get(key string) *jose.JSONWebKey {
	return s[key]
}
