<h1 align="center"><img src="./docs/images/banner.png" alt="Helios - Identity-aware Proxy"></h1>

Helios is an Identity & Access Proxy (IAP) that authorizes HTTP requests based on sets of rules. 
It is the building block towards [BeyondCorp](https://beyondcorp.com), a model designed by Google to secure applications
in Zero-Trust networks.

My goal is to build an open source alternative to
[Cloudflare Access](https://www.cloudflare.com/products/cloudflare-access/)
and [Cloud IAP](https://cloud.google.com/iap/).

In a nutshell, with Helios you can:

* Identify users using existing identity providers like Google, Auth0, Okta and more
* Secure and authenticate access to any domain or path 
* Configure authorization policies using [CEL](https://github.com/google/cel-spec) expressions
* Use Helios as gateway or reverse proxy 

**Another Go proxy?**

This is a project I started off for 2 reasons: I wanted to exercise and continue improving my Go skills, and because
I'm very interested in BeyondCorp model. This is a side project of mine but I'm determined to reach a stable version.

## Development

### Prerequisites

 - Go 1.12
 - [mkcert](https://github.com/FiloSottile/mkcert)

### Environment Setup

Deploy local CA

```shell
mkcert -install
```

Create a certificate for local development

```shell
mkcert localhost 127.0.0.1
```

### Usage

Install dependencies

```shell
go mod download
```

Run the program

```shell
go run . -config config.example.yaml
```
