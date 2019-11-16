# Authentikos ([αὐθεντικός](https://en.wikipedia.org/wiki/Authentication))

## Description

`authentikos` is a service used to create/refresh a Google oauth token and store the value in a Kubernetes secret.

## Installation

```console
$ go get -u istio.io/test-infra/authentikos
```

## Usage

Run using Golang:

```console
$ go run istio.io/test-infra/authentikos <options>
```

The following is a list of supported options for `authentikos`:

```console
  -c, --creds string        Path to a JSON credentials file.
  -f, --format string       Format string for the token. (default "%v")
  -n, --namespace strings   Namespace(s) to create the secret in. (default [default])
  -s, --scopes strings      Oauth scope(s) to request for token.
  -o, --secret string       Name of secret to create. (default "authentikos-token")
  -v, --verbose             Print verbose output.
```
