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
  -r, --force-refresh          Force a token refresh. Otherwise, the token will only refresh when necessary.
  -i, --interval duration      Token refresh interval [1m0s - 50m0s). (default 30m0s)
  -k, --key string             Name of secret data key. (default "token")
  -n, --namespace strings      Namespace(s) to create the secret in. (default [default])
  -s, --scopes strings         Oauth scope(s) to request for token (see: https://developers.google.com/identity/protocols/oauth2/scopes).
  -o, --secret string          Name of secret to create. (default "authentikos-token")
  -t, --template string        Template string for the token.
  -f, --template-file string   Path to a template string for the token.
  -v, --verbose                Print verbose output.
```

## Changelog

- 0.0.1: Initial release
- 0.0.2: Remove `--format` option and add `--template` and `--template-file` options.
- 0.0.3: Add new `TimeToUnix`, `UnixToTime`, and `Parse` template variable and change method signature for math template variables from `(a, b time.Duration) time.Duration` to `(a, b int64) int64`.
- 0.0.4: Add `--key` option for specifying the name of the data key in the created Kubernetes secret.
- 0.0.5: Use [Sprig](http://masterminds.github.io/sprig/) as the library for template functions.
- 0.0.6: Add `--force-refresh` option for forcing a token refresh. If this option is omitted or false, the token will only refresh when necessary. Add `--interval` option for customizing the token refresh interval. If unspecified, default scopes to _userinfo.email_, _cloud-platform_, and _openid_.
- 0.0.7: Add more descriptive error logging for token creation failure.
