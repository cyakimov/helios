<h1 align="center"><img src="./docs/images/banner.png" alt="Helios - Identity-aware Proxy"></h1>

**⚠ This project is on early stage and it's not ready for production yet ⚠**

Helios is an Identity & Access Proxy (IAP) that authorizes HTTP requests based on sets of rules. 
It is the building block towards [BeyondCorp](https://beyondcorp.com), a model designed by Google to secure applications
in Zero-Trust networks.

In a nutshell, with Helios you can:

* Identify users using existing identity providers like Google, Auth0, Okta, etc.
* Secure and authenticate access to any domain or path 
* Configure authorization policies using [CEL](https://github.com/google/cel-spec) expressions
* Use Helios as gateway or reverse proxy 

## Motivation

My goal is to build an open source alternative to
[Cloudflare Access](https://www.cloudflare.com/products/cloudflare-access/)
and [Cloud IAP](https://cloud.google.com/iap/).

Beyond that, I started this project off for 2 reasons:

1. I wanted to exercise and continue improving my Go skills
2. I'm very interested in BeyondCorp. I believe it's the future of enterprise security.

## Install
First, install Go, set your `GOPATH`, and make sure `$GOPATH/bin` is on your PATH.

```shell
$ export GOPATH="$HOME/go"
$ export PATH="$PATH:$GOPATH/bin"
```

Next download the project and build the binary file.

```shell
$ go get -u github.com/cyakimov/helios
```

## Usage

```shell
helios -config config.example.yaml
```

List flags with

```shell
helios -help
```

### Configuring authorization rules

The supported condition attributes are based on details about the request (e.g., its timestamp, originating IP address
, etc).
Examples and a description attribute types are described below.

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

Alternatively, you can check if a request comes from a particular network:

```
request.ip.network("192.168.0.0/24")
```

**Example Date/Time Expressions**

Allow access temporarily until a specified expiration date/time:

```timestamp(request.time) < timestamp("2019-01-01T07:00:00Z")```

Allow access only during specified working hours:

```
timestamp(request.time).getHours("America/Santiago") >= 9 &&
timestamp(request.time).getHours("America/Santiago") <= 17 &&
timestamp(request.time).getDayOfWeek("America/Santiago") >= 1 &&
timestamp(request.time).getDayOfWeek("America/Santiago") <= 5
```

Allow access only for a specified month and year:

```
timestamp(request.time).getFullYear("America/Santiago") == 2018
timestamp(request.time).getMonth("America/Santiago") < 6
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

Install dependencies

```shell
go mod download
```

Run the program

```shell
go run . -config config.example.yaml
```
