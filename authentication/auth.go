package authentication

import (
	"encoding/base64"
	"errors"
	"net/http"
	"time"

	"github.com/cyakimov/helios/authentication/providers"
	log "github.com/sirupsen/logrus"
)

// CookieName is the name of the cookie that contains the JWT token
const CookieName = "Helios_Authorization"
// HeaderName is the name of the cookie that contains the JWT token
const HeaderName = "Helios-Jwt-Assertion"

// ErrUnauthorized is returned by the middleware when a request is not authorized
var ErrUnauthorized = errors.New("unauthorized request")

// JWTConfig JWT configuration
type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

// Helios represents a middleware instance that can authenticate requests
type Helios struct {
	provider  providers.OAuth2
	jwtConfig JWTConfig
}

// NewHeliosAuthentication creates a new authentication middleware instance
func NewHeliosAuthentication(provider providers.OAuth2, jwtSecret string, jwtExpiration time.Duration) Helios {
	return Helios{
		provider: provider,
		jwtConfig: JWTConfig{
			Secret:     jwtSecret,
			Expiration: jwtExpiration,
		},
	}
}

// Middleware checks if a request is authentic
func (helios Helios) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := authenticate(helios.jwtConfig.Secret, r); err != nil {

			// dynamically build callback URL based on current domain
			callback := "https://" + r.Host + "/.oauth2/callback"

			url := helios.provider.GetLoginURL(callback, r.RequestURI)

			log.Debugf("Redirecting to %s", url)

			http.Redirect(w, r, url, http.StatusTemporaryRedirect)
			return
		}

		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

// CallbackHandler handles OAuth 2 callback flow
func (helios Helios) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	// decode and decrypt state to recover original request url
	encodedState := r.URL.Query().Get("state")

	// @todo decrypt state (see GetLoginURL)
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

	exp := time.Now().Add(helios.jwtConfig.Expiration)
	jwt, err := IssueJWTWithSecret(helios.jwtConfig.Secret, profile.Email, exp)
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
