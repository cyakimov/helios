package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"
)

const ServerName = "Helios/0.1"

var (
	port            int
	ip              string
	timeout         int
	certificatePath string
	privateKeyPath  string
	tlsConfig       *tls.Config
)

func init() {
	_ = os.Setenv("GODEBUG", os.Getenv("GODEBUG")+",tls13=1")

	flag.IntVar(&port, "port", 443, "Listen port")
	flag.StringVar(&ip, "ip", "0.0.0.0", "Listen IP")
	flag.IntVar(&timeout, "timeout", 3, "Read/Write timeout")
	flag.StringVar(&certificatePath, "certificate", "localhost.pem", "TLS certificate path")
	flag.StringVar(&privateKeyPath, "key", "localhost-key.pem", "TLS private key path")
	flag.Parse()

	log.SetLevel(log.DebugLevel)
}

func main() {
	var wait time.Duration

	tlsConfig = &tls.Config{
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS13,
	}

	upstream, _ := url.Parse("http://httpbin.org")
	router := mux.NewRouter()
	router.PathPrefix("/").Handler(NewProxy(upstream))

	srv := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", ip, port),
		WriteTimeout:   time.Second * 15,
		ReadTimeout:    time.Second * 15,
		IdleTimeout:    time.Second * 60,
		TLSConfig:      tlsConfig,
		MaxHeaderBytes: 1 << 20,
		Handler:        router,
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServeTLS(certificatePath, privateKeyPath); err != nil {
			log.Fatal(err)
		}
	}()
	log.Infof("Listening on %s:%d", ip, port)

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
