package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/cyakimov/helios/authentication"
	"github.com/cyakimov/helios/authentication/providers"
	"github.com/cyakimov/helios/authentication/providers/auth0"
	"github.com/cyakimov/helios/authentication/providers/azuread"
	"github.com/cyakimov/helios/authentication/providers/google"
	"github.com/cyakimov/helios/authorization"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var (
	configPath string
	config     *Config
	tlsConfig  *tls.Config
	debugMode  bool
)

func init() {
	// Enable TLS 1.3
	_ = os.Setenv("GODEBUG", os.Getenv("GODEBUG")+",tls13=1")

	flag.StringVar(&configPath, "config", "default.yaml", "Configuration file path")
	flag.BoolVar(&debugMode, "verbose", false, "DEBUG level logging")
}

func router() *mux.Router {
	router := mux.NewRouter()
	upstreams := make(map[string]http.Handler, len(config.Upstreams))

	oauth2conf := providers.OAuth2Config{
		ClientID:     config.Identity.ClientID,
		ClientSecret: config.Identity.ClientSecret,
		AuthURL:      config.Identity.OAuth2.AuthURL,
		TokenURL:     config.Identity.OAuth2.TokenURL,
		ProfileURL:   config.Identity.OAuth2.ProfileURL,
	}

	var provider providers.OAuth2Provider
	switch config.Identity.Provider {
	case "aad":
		provider = azuread.NewAzureADProvider(oauth2conf)
	case "auth0":
		provider = auth0.NewAuth0Provider(oauth2conf)
	case "google":
		provider = google.NewGoogleProvider(oauth2conf)
	default:
		log.Fatalf("%q provider is not supported", config.Identity.Provider)
	}

	authN := authentication.NewHeliosAuthentication(provider, config.JWT.Secret, config.JWT.Expires)

	router.PathPrefix("/.well-known/callback").HandlerFunc(authN.CallbackHandler)
	router.PathPrefix("/.well-known/logout").HandlerFunc(authN.Logout)

	for _, up := range config.Upstreams {
		upstreamURL, err := url.Parse(up.URL)
		if err != nil {
			log.Fatalf("Cannot parse upstream %q URL: %v", up.Name, err)
		}

		conf := ReverseProxyConfig{
			ConnectTimeout: up.ConnectTimeout,
			IdleTimeout:    config.Server.IdleTimeout,
			Timeout:        config.Server.Timeout,
		}
		proxy := NewSingleHostReverseProxy(upstreamURL, conf)
		upstreams[up.Name] = proxy
	}

	for _, route := range config.Routes {
		h := router.Host(route.Host).Subrouter()

		for _, path := range route.HTTP.Paths {
			upstream := upstreams[path.Upstream]

			if upstream == nil {
				log.Fatalf("Upstream %q for route %q not found", path.Upstream, route.Host)
				break
			}

			authZ := authorization.NewAuthorization(route.Rules)

			if path.Authentication {
				h.PathPrefix(path.Path).Handler(authN.Middleware(authZ.Middleware(upstream)))
			} else {
				h.PathPrefix(path.Path).Handler(authZ.Middleware(upstream))
			}

		}
	}

	return router
}

func main() {
	flag.Parse()
	if debugMode {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	cb, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	if err = yaml.Unmarshal(cb, &config); err != nil {
		log.Fatalf("Error parsing configuration: %v", err)
	}

	var wait time.Duration

	tlsConfig = &tls.Config{
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS13,
	}

	address := fmt.Sprintf("%s:%d", config.Server.ListenIP, config.Server.ListenPort)
	srv := &http.Server{
		Addr:           address,
		WriteTimeout:   config.Server.Timeout,
		ReadTimeout:    config.Server.Timeout,
		IdleTimeout:    config.Server.IdleTimeout,
		TLSConfig:      tlsConfig,
		MaxHeaderBytes: 1 << 20, // 1mb
		Handler:        router(),
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServeTLS(config.Server.TLSContext.CertificatePath, config.Server.TLSContext.PrivateKeyPath); err != nil {
			log.Fatal(err)
		}
	}()
	log.Infof("Listening on %s", address)

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	if err = srv.Shutdown(ctx); err != nil {
		log.Error(err)
	}

	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Info("shutting down")
	os.Exit(0)
}
