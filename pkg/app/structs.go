package app

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/mattn/go-mastodon"
	"github.com/rmrfslashbin/mastostart/pkg/mastoclient"
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

type PreflightInput struct {
	jwtToken *jwt.Token
}

type PreflightOutput struct {
	Client      *mastoclient.Config
	Userid      *mastodon.ID
	FQUsername  *string
	InstanceURL *string
	Username    *string
}
