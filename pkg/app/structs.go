package app

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/mattn/go-mastodon"
)

// JWTClaims is the JWT claims struct
type JWTClaims struct {
	AccessToken string `json:"access_token"`
	jwt.RegisteredClaims
}

type AuthVerifyReturn struct {
	Account    *mastodon.Account `json:"account"`
	LastStatus *mastodon.Status  `json:"last_status"`
}
