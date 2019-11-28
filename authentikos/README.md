# Authentikos ([αὐθεντικός](https://en.wikipedia.org/wiki/Authentication))

## Description

`authentikos` is a service used to create/refresh a Google oauth token and store the value in a Kubernetes secret.

## Installation

Install using Golang:

```console
$ GO111MODULE="on" go get -u istio.io/test-infra/authentikos
```

Install using Docker:

```console
$ docker pull gcr.io/istio-testing/authentikos:latest
```

## Usage

Run using Golang:
> Ensure `$GOPATH/bin` is on your `$PATH`; or execute `$GOPATH/bin/authentikos` directly.

```console
$ authentikos <options>
```

Run using Docker:

```console
$ docker run gcr.io/istio-testing/authentikos:latest <options>
```

The following is a list of supported options for `authentikos`:

```console
  -c, --creds string           Path to a JSON credentials file.
  -n, --namespace strings      Namespace(s) to create the secret in. (default [default])
  -s, --scopes strings         Oauth scope(s) to request for token.
  -o, --secret string          Name of secret to create. (default "authentikos-token")
  -t, --template string        Template string for the token.
  -f, --template-file string   Path to a template string for the token.
  -v, --verbose                Print verbose output.
```

## Changelog

- 0.0.1: initial release
- 0.0.2: remove `--format` option and add `--template` and `--template-file` options.
