package providers

import (
	"errors"
	"net/http"
)

type OAuth2Config struct {
	ClientID     string
	ClientSecret string
	AuthURL      string
	TokenURL     string
	ProfileURL   string
}

type OIDCProfile struct {
	Email string
}

type OAuth2 interface {
	GetUserProfile(r *http.Request) (OIDCProfile, error)
	GetLoginURL(callbackURL, state string) string
}

var ErrCodeExchange = errors.New("error on code exchange")
var ErrProfile = errors.New("error getting user profile")
var ErrNoEmail = errors.New("no email found in user profile")
