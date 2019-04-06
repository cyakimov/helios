package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"
)

const ServerName = "Helios/0.1"

var (
	configPath string
	config     *Config
	tlsConfig  *tls.Config
)

func init() {
	_ = os.Setenv("GODEBUG", os.Getenv("GODEBUG")+",tls13=1")

	flag.StringVar(&configPath, "config", "default.yaml", "Configuration file path")
	flag.Parse()

	log.SetLevel(log.DebugLevel)
}

func setupRoutes() *mux.Router {
	router := mux.NewRouter()
	upstreams := map[string]*http.Handler{}

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
		upstreams[up.Name] = &proxy
	}

	for _, route := range config.Routes {
		h := router.Host(route.Host)

		for _, path := range route.HTTP.Paths {
			up := upstreams[path.Upstream]

			if up == nil {
				log.Fatalf("Upstream %q for route %q not found", path.Upstream, route.Host)
			}

			h.PathPrefix(path.Path).Handler(*up)
		}
	}

	return router
}

func main() {
	// Set sane defaults
	cb, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	if err = yaml.UnmarshalStrict(cb, &config); err != nil {
		log.Fatalf("Error parsing configuration: %v", err)
	}

	var wait time.Duration

	tlsConfig = &tls.Config{
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS13,
	}

	router := setupRoutes()

	address := fmt.Sprintf("%s:%d", config.Server.ListenIP, config.Server.ListenPort)
	srv := &http.Server{
		Addr:           address,
		WriteTimeout:   config.Server.Timeout,
		ReadTimeout:    config.Server.Timeout,
		IdleTimeout:    config.Server.IdleTimeout,
		TLSConfig:      tlsConfig,
		MaxHeaderBytes: 1 << 20,
		Handler:        router,
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
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Info("shutting down")
	os.Exit(0)
}
