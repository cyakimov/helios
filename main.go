package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"os"
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
	flag.StringVar(&ip, "ip", "", "Listen IP")
	flag.IntVar(&timeout, "timeout", 3, "Read/Write timeout")
	flag.StringVar(&certificatePath, "certificate", "localhost.pem", "TLS certificate path")
	flag.StringVar(&privateKeyPath, "key", "localhost-key.pem", "TLS private key path")
	flag.Parse()

	log.SetLevel(log.DebugLevel)
}

func serveConn(conn net.Conn) {
	log.Debugf("Got a new connection from %s", conn.RemoteAddr().String())
	defer conn.Close()

	// Set read/write timeout
	err := conn.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
	// drop connection on error
	if err != nil {
		return
	}

	tlsConn := tls.Server(conn, tlsConfig)
	if err := tlsConn.Handshake(); err != nil {
		log.Errorf("TLS handshake error: %q", err)
		return
	}
	proxy(tlsConn)
}

func proxy(conn net.Conn) {
	remote, err := net.Dial("tcp", "httpbin.org:443")
	if err != nil {
		return
	}

	defer remote.Close()

	tlsConn := tls.Client(remote, &tls.Config{
		ServerName: "httpbin.org",
	})
	if err = tlsConn.Handshake(); err != nil {
		log.Error(err)
		return
	}

	if err = conn.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second)); err != nil {
		log.Debugf("Connection setTimeout error: %q", err)
		return
	}

	go io.Copy(tlsConn, conn)
	io.Copy(conn, tlsConn)
}

func main() {
	// load TLS certificate
	cert, err := tls.LoadX509KeyPair(certificatePath, privateKeyPath)
	if err != nil {
		log.Fatal(err)
	}
	tlsConfig = &tls.Config{
		MinVersion:   tls.VersionTLS12,
		MaxVersion:   tls.VersionTLS13,
		Certificates: []tls.Certificate{cert},
		//NextProtos:   []string{"h2"},
	}

	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("Listening on %s:%d", ip, port)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Error(err)
		}

		// TODO: limit open connections
		go serveConn(conn)
	}
}
