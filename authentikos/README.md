# Authentikos ([αὐθεντικός](https://en.wikipedia.org/wiki/Authentication))

## Description

`authentikos` is a service used to create/refresh a Google oauth token and store the value in a Kubernetes secret.

## Installation

Install using Golang:

```shell
GO111MODULE="on" go get -u istio.io/test-infra/authentikos
```

Install using Docker:

```shell
docker pull gcr.io/istio-testing/authentikos:latest
```

## Usage

Run using Golang:
> Ensure `$GOPATH/bin` is on your `$PATH`; or execute `$GOPATH/bin/authentikos` directly.

```shell
authentikos <options>
```

Run using Docker:

```shell
docker run gcr.io/istio-testing/authentikos:latest <options>
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
- 0.0.3: add new `TimeToUnix`, `UnixToTime`, and `Parse` template variable and change method signature for math template variables from `(a, b time.Duration) time.Duration` to `(a, b int64) int64`.
