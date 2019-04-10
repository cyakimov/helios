package authentication

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"net/http"
)

type Auth0Provider struct {
	OAuth2Provider
	oauth2 oauth2.Config
	domain string
}

func NewAuth0Provider(config OAuth2Config) OAuth2Provider {
	return Auth0Provider{
		domain: config.Domain,
		oauth2: oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			RedirectURL:  config.CallbackURL,
			Scopes:       []string{"openid", "email_verified", "email"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  config.AuthURL,
				TokenURL: config.TokenURL,
			},
		},
	}
}

func (provider Auth0Provider) GetUserProfile(r *http.Request) (OIDCProfile, error) {
	var profile OIDCProfile
	code := r.URL.Query().Get("code")

	token, err := provider.oauth2.Exchange(context.TODO(), code)
	if err != nil {
		return profile, ErrCodeExchange
	}

	// Get user profile
	client := provider.oauth2.Client(context.TODO(), token)
	resp, err := client.Get(provider.domain + "/userinfo")
	if err != nil {
		return profile, ErrProfile
	}

	defer func() {
		if resp.Body != nil {
			if err := resp.Body.Close(); err != nil {
				log.Error(err)
			}
		}
	}()

	if err = json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return profile, errors.New("auth0: cannot decode JSON profile")
	}

	if profile.Email == "" {
		return profile, ErrNoEmail
	}

	return profile, nil
}

func (provider Auth0Provider) GetLoginURL(state string) string {
	s := base64.StdEncoding.EncodeToString([]byte(state))
	return provider.oauth2.AuthCodeURL(s)
}
