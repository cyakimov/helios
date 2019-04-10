package authentication

import (
	"encoding/base64"
	"errors"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type OAuth2Config struct {
	ClientID     string
	ClientSecret string
	CallbackURL  string
	AuthURL      string
	TokenURL     string
	Domain       string
}

type OIDCProfile struct {
	Email string
}

type OAuth2Provider interface {
	GetUserProfile(r *http.Request) (OIDCProfile, error)
	GetLoginURL(state string) string
}

const CookieName = "Helios_Authorization"
const HeaderName = "Helios-Jwt-Assertion"

var ErrUnauthorized = errors.New("unauthorized request")
var ErrCodeExchange = errors.New("error on code exchange")
var ErrProfile = errors.New("error getting user profile")
var ErrNoEmail = errors.New("no email found in user profile")

func authenticate(r *http.Request) error {
	// look for Token in both Cookies and Headers
	_, err := r.Cookie(CookieName)
	htoken := r.Header.Get(HeaderName)

	if err == http.ErrNoCookie && htoken == "" {
		return ErrUnauthorized
	}

	return nil
}

func Middleware(provider OAuth2Provider, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := authenticate(r); err != nil {
			url := provider.GetLoginURL(r.RequestURI)
			log.Debugf("Redirecting to %s", url)
			http.Redirect(w, r, url, http.StatusTemporaryRedirect)
			return
		}

		log.Println(r.RequestURI)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

func CallbackHandler(provider OAuth2Provider, jwtSecret string, jwtDuration time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		state := r.URL.Query().Get("state")

		dstate, err := base64.StdEncoding.DecodeString(state)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		profile, err := provider.GetUserProfile(r)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Debugf("Authorized. Redirecting to %s", string(dstate))

		exp := time.Now().Add(jwtDuration)
		jwt, err := IssueJWT(jwtSecret, profile.Email, exp)
		http.SetCookie(w, &http.Cookie{
			Name:    CookieName,
			Value:   jwt,
			Expires: exp,
			Path:    "/",
			Secure:  true,
		})

		http.Redirect(w, r, string(dstate), http.StatusFound)
		return
	})
}
