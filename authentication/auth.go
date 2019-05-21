package authentication

import (
	"encoding/base64"
	"errors"
	"github.com/cyakimov/helios/authentication/providers"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

const CookieName = "Helios_Authorization"
const HeaderName = "Helios-Jwt-Assertion"

var ErrUnauthorized = errors.New("unauthorized request")

type JWTOpts struct {
	Secret     string
	Expiration time.Duration
}

type Helios struct {
	provider providers.OAuth2
	jwtOpts  JWTOpts
}

func NewHeliosAuthentication(provider providers.OAuth2, jwtSecret string, jwtExpiration time.Duration) Helios {
	return Helios{
		provider: provider,
		jwtOpts: JWTOpts{
			Secret:     jwtSecret,
			Expiration: jwtExpiration,
		},
	}
}

func (helios Helios) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := authenticate(helios.jwtOpts.Secret, r); err != nil {

			// dynamically build callback URL based on current domain
			callback := "https://" + r.Host + "/.oauth2/callback"

			url := helios.provider.GetLoginURL(callback, r.RequestURI)

			log.Debugf("Redirecting to %s", url)

			http.Redirect(w, r, url, http.StatusTemporaryRedirect)
			return
		}

		log.Println(r.RequestURI)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

func (helios Helios) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	// decode and decrypt state to recover original request url
	encodedState := r.URL.Query().Get("state")

	state, err := base64.StdEncoding.DecodeString(encodedState)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	profile, err := helios.provider.GetUserProfile(r)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Debugf("Authorized. Redirecting to %s", string(state))

	exp := time.Now().Add(helios.jwtOpts.Expiration)
	jwt, err := IssueJWTWithSecret(helios.jwtOpts.Secret, profile.Email, exp)
	if err != nil {
		log.Error(err)
	}
	http.SetCookie(w, &http.Cookie{
		Name:    CookieName,
		Value:   jwt,
		Expires: exp,
		Path:    "/",
		Secure:  true,
	})

	http.Redirect(w, r, string(state), http.StatusFound)
}

func authenticate(jwtSecret string, r *http.Request) error {
	// look for Token in both Cookies and Headers
	cookie, err := r.Cookie(CookieName)
	token := r.Header.Get(HeaderName)

	if err == http.ErrNoCookie && token == "" {
		return ErrUnauthorized
	}

	if token == "" {
		token = cookie.Value
	}

	if !ValidateJWTWithSecret(jwtSecret, token) {
		return ErrUnauthorized
	}

	return nil
}
