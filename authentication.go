package openbank

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

func (c *Client) Authenticate() error {
	claims := c.authClaims()
	t := jwt.NewWithClaims(jwt.GetSigningMethod("RS256"), claims)
	signBytes, err := ioutil.ReadFile(c.PrivateKeyPath)
	if err != nil {
		return err
	}
	signKey, err := jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		return err
	}
	tokenString, err := t.SignedString(signKey)
	if err != nil {
		return err
	}

	data := url.Values{}
	data.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
	data.Set("client_assertion", tokenString)
	data.Set("client_id", c.ClientID)
	data.Set("grant_type", "client_credentials")

	u, err := c.AccountURL.Parse("/auth/realms/stone_bank/protocol/openid-connect/token")
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", u.String(), strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("user-agent", c.UserAgent)
	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	var token oauth2.Token

	_, err = c.Do(req, &token)
	if err != nil {
		return err
	}
	ctx := context.Background()
	config := &oauth2.Config{}
	ts := config.TokenSource(ctx, &token)

	c.client = oauth2.NewClient(ctx, ts)

	return nil
}

func (c *Client) authClaims() jwt.MapClaims {
	now := time.Now()
	u, _ := c.AccountURL.Parse("/auth/realms/stone_bank")
	claims := jwt.MapClaims{
		"aud":       u.String(),
		"client_id": c.ClientID,
		"exp":       now.Add(time.Hour * time.Duration(2)).Unix(),
		"iat":       now.Unix(),
		"jti":       uuid.New().String(),
		"nbf":       now.Unix(),
		"realm":     "stone_bank",
		"sub":       c.ClientID,
	}
	return claims
}
