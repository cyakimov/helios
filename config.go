package main

import (
	"time"
)

type Config struct {
	Server    Server     `yaml:"server"`
	Upstreams []Upstream `yaml:"upstreams"`
	Routes    []Route    `yaml:"routes"`
	Identity  Identity   `yaml:"identity"`
	JWT       JWT        `yaml:"jwt"`
}

type Server struct {
	ListenIP    string        `yaml:"listen_ip"`
	ListenPort  int           `yaml:"listen_port"`
	Timeout     time.Duration `yaml:"timeout"`
	IdleTimeout time.Duration `yaml:"idle_timeout"`
	TLSContext  TLSContext    `yaml:"tls_context"`
}

type TLSContext struct {
	CertificatePath string `yaml:"certificate_path"`
	PrivateKeyPath  string `yaml:"private_key_path"`
}

type Route struct {
	Host string
	HTTP struct {
		Paths []struct {
			Path        string
			Upstream    string
			AuthEnabled bool `yaml:"auth_enabled"`
		}
	}
}

type Upstream struct {
	Name           string `yaml:"name"`
	URL            string `yaml:"url"`
	ConnectTimeout time.Duration
}

type Identity struct {
	Provider     string `yaml:"provider"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	OAuth2       struct {
		CallbackURL string `yaml:"callback_url"`
		AuthURL     string `yaml:"auth_url"`
		TokenURL    string `yaml:"token_url"`
		Domain      string `yaml:"domain"`
	}
}

type JWT struct {
	SharedSecret string
	Expires      time.Duration
}

func (c *Upstream) UnmarshalYAML(unmarshal func(v interface{}) error) error {
	buf := struct {
		ConnectTimeout string `yaml:"connect_timeout"`
		Name           string `yaml:"name"`
		URL            string `yaml:"url"`
	}{}

	if err := unmarshal(&buf); err != nil {
		return err
	}

	timeout, err := time.ParseDuration(buf.ConnectTimeout)
	if err != nil {
		return err
	}

	c.ConnectTimeout = timeout
	c.URL = buf.URL
	c.Name = buf.Name

	return nil
}

func (c *JWT) UnmarshalYAML(unmarshal func(v interface{}) error) error {
	buf := struct {
		SharedSecret string `yaml:"shared_secret"`
		Expires      string `yaml:"expires"`
	}{}

	if err := unmarshal(&buf); err != nil {
		return err
	}

	expires, err := time.ParseDuration(buf.Expires)
	if err != nil {
		return err
	}

	c.Expires = expires
	c.SharedSecret = buf.SharedSecret

	return nil
}

func (c *Server) UnmarshalYAML(unmarshal func(v interface{}) error) error {
	var buf struct {
		ListenIP    string     `yaml:"listen_ip"`
		ListenPort  int        `yaml:"listen_port"`
		Timeout     string     `yaml:"timeout"`
		IdleTimeout string     `yaml:"idle_timeout"`
		TLSContext  TLSContext `yaml:"tls_context"`
	}

	if err := unmarshal(&buf); err != nil {
		return err
	}

	timeout, err := time.ParseDuration(buf.Timeout)
	if err != nil {
		return err
	}
	idleTimeout, err := time.ParseDuration(buf.IdleTimeout)
	if err != nil {
		return err
	}

	c.Timeout = timeout
	c.IdleTimeout = idleTimeout
	c.TLSContext = buf.TLSContext
	c.ListenIP = buf.ListenIP
	c.ListenPort = buf.ListenPort

	return nil
}
