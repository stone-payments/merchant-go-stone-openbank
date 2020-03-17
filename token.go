package openbank

import (
	"io/ioutil"

	jwt "github.com/dgrijalva/jwt-go"
)

func (c *Client) generateToken(claims jwt.MapClaims) (string, error) {
	t := jwt.NewWithClaims(jwt.GetSigningMethod("RS256"), claims)
	signBytes, err := ioutil.ReadFile(c.PrivateKeyPath)
	if err != nil {
		return "", err
	}
	signKey, err := jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		return "", err
	}
	tokenString, err := t.SignedString(signKey)
	if err != nil {
		return "", err
	}
	return tokenString, err
}
