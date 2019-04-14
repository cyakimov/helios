package providers

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
	OAuth2
	oauth2     oauth2.Config
	profileURL string
}

func NewAuth0Provider(config OAuth2Config) OAuth2 {
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

func (provider Auth0Provider) GetUserProfile(r *http.Request) (OIDCProfile, error) {
	var profile OIDCProfile
	code := r.URL.Query().Get("code")

	// Auth0 requires callback URL
	url := "https://" + r.Host + "/" + r.URL.Path
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

func (provider Auth0Provider) GetLoginURL(callbackURL string, state string) string {
	s := base64.StdEncoding.EncodeToString([]byte(state))

	callback := oauth2.SetAuthURLParam("redirect_uri", callbackURL)

	return provider.oauth2.AuthCodeURL(s, callback)
}
