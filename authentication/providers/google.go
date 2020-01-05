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

type GoogleProvider struct {
	OAuth2Provider
	oauth2 oauth2.Config
}

const profileURL = "https://www.googleapis.com/oauth2/v2/userinfo"

func NewGoogleProvider(config OAuth2Config) OAuth2Provider {
	authURL := "https://accounts.google.com/o/oauth2/v2/auth"
	tokenURL := "https://oauth2.googleapis.com/token"
	if config.AuthURL != "" {
		authURL = config.AuthURL
	}
	if config.TokenURL != "" {
		tokenURL = config.TokenURL
	}

	return &GoogleProvider{
		oauth2: oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  authURL,
				TokenURL: tokenURL,
			},
			Scopes: []string{"email"},
		},
	}
}

func (provider GoogleProvider) FetchUser(r *http.Request) (UserInfo, error) {
	var userInfo UserInfo
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
		return userInfo, ErrCodeExchange
	}

	// get user profile
	client := provider.oauth2.Client(context.TODO(), token)
	resp, err := client.Get(profileURL)
	if err != nil {
		return userInfo, ErrProfile
	}

	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return userInfo, errors.New("auth0: cannot decode JSON profile")
	}

	if userInfo.Email == "" {
		return userInfo, ErrNoEmail
	}

	return userInfo, nil
}

// GetLoginURL returns OAuth 2 login endpoint used to redirect users
func (provider GoogleProvider) GetLoginURL(callbackURL string, state string) string {
	// @todo encrypt state
	s := base64.StdEncoding.EncodeToString([]byte(state))

	callback := oauth2.SetAuthURLParam("redirect_uri", callbackURL)

	return provider.oauth2.AuthCodeURL(s, callback)
}
