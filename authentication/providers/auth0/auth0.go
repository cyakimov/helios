package auth0

import (
	"context"
	"encoding/base64"
	"github.com/cyakimov/helios/authentication/providers"
	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"net/http"
)

// Auth0Provider represents an Auth0 instance
type Provider struct {
	providers.OAuth2Provider
	oauth2     oauth2.Config
	profileURL string
}

// NewAuth0Provider creates a new Auth0 identity provider with a given config
func NewAuth0Provider(config providers.OAuth2Config) providers.OAuth2Provider {
	return Provider{
		profileURL: config.ProfileURL,
		oauth2: oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			Scopes:       []string{"openid", "email_verified", "email"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  config.AuthURL,
				TokenURL: config.TokenURL,
			},
		},
	}
}

// FetchUser fetches user info from Auth0
func (provider Provider) FetchUser(r *http.Request) (providers.UserInfo, error) {
	var userInfo providers.UserInfo
	code := r.URL.Query().Get("code")

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	url := scheme + "://" + r.Host + r.URL.Path
	callback := oauth2.SetAuthURLParam("redirect_uri", url)

	// get access token
	token, err := provider.oauth2.Exchange(context.TODO(), code, callback)
	if err != nil {
		log.Error(err)
		return userInfo, providers.ErrCodeExchange
	}

	// Parse the token
	jwtToken, _ := jwt.ParseWithClaims(token.Extra("id_token").(string), &providers.OIDClaims{}, nil)
	if jwtToken == nil {
		return userInfo, providers.ErrJWTParse
	}

	claims, ok := jwtToken.Claims.(*providers.OIDClaims)
	if !ok {
		return userInfo, providers.ErrJWTClaims
	}

	if claims.Email == "" {
		return userInfo, providers.ErrNoEmail
	}

	userInfo.Email = claims.Email

	return userInfo, nil
}

// GetLoginURL returns OAuth 2 login endpoint used to redirect users
func (provider Provider) GetLoginURL(callbackURL string, state string) string {
	// @todo encrypt state
	s := base64.StdEncoding.EncodeToString([]byte(state))

	callback := oauth2.SetAuthURLParam("redirect_uri", callbackURL)

	return provider.oauth2.AuthCodeURL(s, callback)
}
