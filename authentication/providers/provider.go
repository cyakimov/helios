package providers

import (
	"errors"
	"net/http"
)

// OAuth2Config OAuth2 provider configuration settings
type OAuth2Config struct {
	ClientID     string
	ClientSecret string
	AuthURL      string
	TokenURL     string
	ProfileURL   string
}

// OIDCProfile represents a basic Open ID Connect profile
type OIDCProfile struct {
	Email string
}

// OAuth2 provider interface
type OAuth2 interface {
	GetUserProfile(r *http.Request) (OIDCProfile, error)
	GetLoginURL(callbackURL, state string) string
}

// ErrCodeExchange is returned when the auth code exchange failed
var ErrCodeExchange = errors.New("error on code exchange")

// ErrProfile is returned when user profile fetch failed
var ErrProfile = errors.New("error getting user profile")

// ErrNoEmail is returned when no email is present in the OIDC profile
var ErrNoEmail = errors.New("no email found in user profile")
