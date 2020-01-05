package providers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"

	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// Auth0Provider represents an Auth0 instance
type Auth0Provider struct {
	OAuth2Provider
	oauth2     oauth2.Config
	profileURL string
}

// NewAuth0Provider creates a new Auth0 identity provider with a given config
func NewAuth0Provider(config OAuth2Config) OAuth2Provider {
	return Auth0Provider{
		profileURL: config.ProfileURL,
		oauth2: oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			Scopes:       []string{"openid", "email_verified", "email"},
			RedirectURL:  "", // RedirectURL can vary per route host
			Endpoint: oauth2.Endpoint{
				AuthURL:  config.AuthURL,
				TokenURL: config.TokenURL,
			},
		},
	}
}

// FetchUser fetches user info from Auth0
func (provider Auth0Provider) FetchUser(r *http.Request) (UserInfo, error) {
	var profile UserInfo
	code := r.URL.Query().Get("code")

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	// Auth0 requires callback URL
	url := scheme + "://" + r.Host + r.URL.Path
	callback := oauth2.SetAuthURLParam("redirect_uri", url)

	// get access token
	token, err := provider.oauth2.Exchange(context.TODO(), code, callback)
	if err != nil {
		log.Error(err)
		return profile, ErrCodeExchange
	}

	// get user profile
	client := provider.oauth2.Client(context.TODO(), token)
	resp, err := client.Get(provider.profileURL)
	if err != nil {
		return profile, ErrProfile
	}

	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return profile, errors.New("auth0: cannot decode JSON profile")
	}

	if profile.Email == "" {
		return profile, ErrNoEmail
	}

	return profile, nil
}

// GetLoginURL returns OAuth 2 login endpoint used to redirect users
func (provider Auth0Provider) GetLoginURL(callbackURL string, state string) string {
	// @todo encrypt state
	s := base64.StdEncoding.EncodeToString([]byte(state))

	callback := oauth2.SetAuthURLParam("redirect_uri", callbackURL)

	return provider.oauth2.AuthCodeURL(s, callback)
}
