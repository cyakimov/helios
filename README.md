<h1 align="center"><img src="./docs/images/banner.png" alt="Helios - Identity-aware Proxy"></h1>

**⚠ This project is on early stage and it's not ready for production yet⚠**

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

This is a project I started off for 2 reasons:
1. I wanted to exercise and continue improving my Go skills
2. I'm very interested in BeyondCorp model. I believe it's the future of enterprise security.

## Configuring authorization rules

The supported condition attributes are based on details about the request (e.g., its timestamp, originating IP address, 
or destination IP address of the target upstream). Examples and a description attribute types are described below.

### Request Attributes

- `request.host`
- `request.path`
- `request.ip`
- `request.timestamp`

For example, by setting Expression to a CEL expression that uses `request.ip` you can limit access to only members
who have a private IP of 10.0.0.1

```
request.ip == "10.0.0.1"
```

**Example Date/Time Expressions**

Allow access temporarily until a specified expiration date/time:

```request.time < timestamp("2019-01-01T07:00:00Z")```

Allow access only during specified working hours:

```
request.time.getHours("Europe/Berlin") >= 9 &&
request.time.getHours("Europe/Berlin") <= 17 &&
request.time.getDayOfWeek("Europe/Berlin") >= 1 &&
request.time.getDayOfWeek("Europe/Berlin") <= 5
```

Allow access only for a specified month and year:

```
request.time.getFullYear("Europe/Berlin") == 2018
request.time.getMonth("Europe/Berlin") < 6
```

**Example URL Host/Path Expressions**

Allow access only for certain subdomains or URL paths in the request:

```
request.host == "hr.example.com"
request.host.endsWith(".example.com")
request.path == "/admin/payroll.js"
request.path.startsWith("/admin")
```

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
