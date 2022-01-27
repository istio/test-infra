/*
Copyright 2019 Istio Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Masterminds/sprig"
	flag "github.com/spf13/pflag"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/transport"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth" // Enable all auth provider plugins
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Command-line options.
const (
	defaultKey       = "token"                         // defaultKey is the kubernetes secret data key.
	defaultSecret    = "authentikos-token"             // defaultSecret is the default kubernetes secret name.
	defaultTemplate  = "{{.Token}}"                    // defaultTemplate is the default token template string.
	defaultNamespace = metav1.NamespaceDefault         // defaultNamespace is the default kubernetes namespace.
	defaultInterval  = 30 * time.Minute                // defaultInterval is the default tick interval.
	minInterval      = 1 * time.Minute                 // minInterval is the minimum tick interval (inclusive).
	maxInterval      = tokenExpiration - expiryDelta*2 // maxInterval is the maximum tick interval (exclusive).
)

// OAuth2 scopes.
const (
	userinfoEmailScope = "https://www.googleapis.com/auth/userinfo.email" // View your email address
	cloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform" // View and manage your data across Google Cloud Platform services
	openIDScope        = "openid"                                         // Authenticate using OpenID Connect
)

var defaultScopes = []string{userinfoEmailScope, cloudPlatformScope, openIDScope} // defaultScopes is the default OAuth2 scopes.

// Token expiration parameters.
const (
	tokenExpiration = 60 * time.Minute // tokenExpiration is the oauth token expiration.
	maxTries        = 5                // maxTries is the maximum number of consecutive attempts to force refresh a token.
	expiryDelta     = 5 * time.Minute  // expiryDelta is how early a token is be considered expired before its actual expiration time.
)

var timeNow = time.Now

// tokenCreator is a function that creates an oauth token.
type tokenCreator func(forceRefresh bool, tries int) ([]byte, *time.Time, error)

// secretCreator is a function that creates a kubernetes secret.
type secretCreator func() ([]*corev1.Secret, namespacedErrors)

// namespacedError is a custom error type which stores a message and a kubernetes namespace.
type namespacedError struct {
	namespace string
	message   string
}

// namespacedError returns the string representation of the error.
func (err namespacedError) Error() string {
	return fmt.Sprintf("%v:%v", err.namespace, err.message)
}

// namespacedError is a list of custom namespaced errors.
type namespacedErrors []*namespacedError

// namespacedErrors returns the string representation of the error(s).
func (errs namespacedErrors) Errors() string {
	var errMsgs []string

	for _, err := range errs {
		errMsgs = append(errMsgs, err.Error())
	}

	return strings.Join(errMsgs, ", ")
}

// tokenTemplate is the template data structure.
type tokenTemplate struct {
	Token  string
	Expire int64
}

// options are the available command-line flags.
type options struct {
	forceRefresh bool
	verbose      bool
	interval     time.Duration
	creds        string
	key          string
	secret       string
	template     string
	templateFile string
	namespace    []string
	scopes       []string
}

// parseFlags parses the command-line flags.
func (o *options) parseFlags() {
	flag.BoolVarP(&o.forceRefresh, "force-refresh", "r", false, "Force a token refresh. Otherwise, the token will only refresh when necessary.")
	flag.BoolVarP(&o.verbose, "verbose", "v", false, "Print verbose output.")
	flag.DurationVarP(&o.interval, "interval", "i", defaultInterval, fmt.Sprintf("Token refresh interval [%v - %v).", minInterval, maxInterval))
	flag.StringVarP(&o.creds, "creds", "c", "", "Path to a JSON credentials file.")
	flag.StringVarP(&o.secret, "secret", "o", defaultSecret, "Name of secret to create.")
	flag.StringVarP(&o.key, "key", "k", defaultKey, "Name of secret data key.")
	flag.StringVarP(&o.template, "template", "t", "", "Template string for the token.")
	flag.StringVarP(&o.templateFile, "template-file", "f", "", "Path to a template string for the token.")
	flag.StringSliceVarP(&o.namespace, "namespace", "n", []string{defaultNamespace}, "Namespace(s) to create the secret in.")
	flag.StringSliceVarP(&o.scopes, "scopes", "s", []string{}, "Oauth scope(s) to request for token (see: https://developers.google.com/identity/protocols/oauth2/scopes).")

	flag.Parse()
}

// validateFlags validates the command-line flags.
func (o *options) validateFlags() error {
	var err error

	// Ensure both `template` and `templateFile` are not set.
	if len(o.template) > 0 && len(o.templateFile) > 0 {
		return errors.New("-t, --template and -f, --template-file are mutually exclusive options")
	}

	// Default to `defaultTemplate` if a template is not specified.
	if len(o.template) == 0 && len(o.templateFile) == 0 {
		o.template = defaultTemplate
	}

	// Read in `templateFile` as template if both set and valid.
	if len(o.templateFile) > 0 {
		data, err := ioutil.ReadFile(o.templateFile)
		if err != nil {
			return fmt.Errorf("-f, --template-file option invalid: %v", o.templateFile)
		}
		o.template = string(data)
	}

	if len(o.scopes) == 0 {
		o.scopes = defaultScopes
	}

	// Secrets must have a name, so if unset then default to `defaultSecret`.
	if len(o.secret) == 0 {
		o.secret = defaultSecret
	}

	// Secrets must have a key, so if unset then default to `defaultKey`.
	if len(o.key) == 0 {
		o.key = defaultKey
	}

	// Secrets must have a namespace, so if unset then default to `defaultNamespace`.
	if len(o.namespace) == 0 {
		o.namespace = []string{defaultNamespace}
	}

	// Tick interval must be [1m - 50m), where 60m is the oauth token expiration, 5m is the token expiry delta, and another 5m for processing delta.
	if o.interval < minInterval || o.interval >= maxInterval {
		return fmt.Errorf("-i, --interval option must be in range [%v, %v): %v", minInterval, maxInterval, o.interval)
	}

	if len(o.creds) > 0 {
		if o.creds, err = filepath.Abs(o.creds); err != nil || !fileExists(o.creds) {
			return fmt.Errorf("-c, --creds option invalid: %v", o.creds)
		}
	}

	return nil
}

// printErrAndExit prints an error message to stderr and exits with a status code.
func printErrAndExit(err error, code int) {
	_, _ = fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(code)
}

// printVerbose prints output based on verbosity level.
func printVerbose(formatString string, verbose bool) {
	if verbose {
		log.Print(formatString)
	}
}

// fileExists checks if a path exists and is a regular file.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.Mode().IsRegular()
}

// isExpired determines if a token is expired and needs to be refreshed.
func isExpired(o options, token *oauth2.Token) bool {
	nextRec := timeNow().Add(o.interval)
	expiry := token.Expiry.Add(-expiryDelta)
	isExpired := expiry.Before(nextRec)
	printVerbose(fmt.Sprintf("expired: %t; token expiry (minus delta): %v; next reconcile: %v\n", isExpired, expiry, nextRec), o.verbose)
	return isExpired
}

// getBackoffTime returns a backoff time calculated using formula: `{backoff factor} * 2 ^ {# of retries}`.
func getBackoffTime(factor float64, retry int) time.Duration {
	return time.Duration(math.Max(factor*math.Exp2(float64(retry)), 0)) * time.Second
}

// withBackoff waits with a backoff and runs a function.
func withBackoff(factor float64, retry int, f interface{}) interface{} {
	time.Sleep(getBackoffTime(factor, retry))
	return f
}

func generateTokenData(o options, data []byte, expiration *time.Time) ([]byte, error) {
	var b bytes.Buffer

	tmpl, err := template.New("TokenData").Funcs(sprig.FuncMap()).Parse(o.template)
	if err != nil {
		return nil, err
	}

	err = tmpl.Execute(&b, &tokenTemplate{
		Token:  string(data),
		Expire: expiration.Unix(),
	})
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

// createClusterConfig creates kubernetes cluster configuration.
func createClusterConfig() (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
}

// loadClusterConfig loads kubernetes cluster configuration.
func loadClusterConfig() (*rest.Config, error) {
	if clusterConfig, err := rest.InClusterConfig(); err == nil {
		return clusterConfig, nil
	} else if clusterConfig, err := createClusterConfig(); err == nil {
		return clusterConfig, nil
	} else {
		return nil, err
	}
}

// getOauthTokenCreator returns a function that creates/refreshes an oauth token.
func getOauthTokenCreator(o options) (tokenCreator, error) {
	var create tokenCreator

	clientOpts := []option.ClientOption{option.WithScopes(o.scopes...)}

	if len(o.creds) > 0 {
		clientOpts = append(clientOpts, option.WithCredentialsFile(o.creds))
	}

	// Reusing the client leverages the token source cache.
	client, err := transport.Creds(context.Background(), clientOpts...)

	clientCreator := func(forceRefresh bool) (*google.Credentials, error) {
		if forceRefresh {
			// Recreating the client invalidates the token source cache.
			client, err = transport.Creds(context.Background(), clientOpts...)
			printVerbose("force refreshing token\n", o.verbose)
		}

		return client, err
	}

	create = func(forceRefresh bool, tries int) ([]byte, *time.Time, error) {
		if tries <= 0 {
			return nil, nil, fmt.Errorf("maximum tries: %d exceeded to force refresh token", maxTries)
		}

		client, err := clientCreator(forceRefresh)
		if err != nil {
			printVerbose(fmt.Sprintf("Failed to create oauth token client: %v.", err), o.verbose)
			return withBackoff(1, maxTries-tries, create).(tokenCreator)(forceRefresh, tries-1)
		}

		token, err := client.TokenSource.Token()
		if err != nil {
			printVerbose(fmt.Sprintf("Failed to get oauth token from client: %v.", err), o.verbose)
			return withBackoff(1, maxTries-tries, create).(tokenCreator)(forceRefresh, tries-1)
		}

		if isExpired(o, token) {
			// Force recreate the token if it will expire before the next reconciliation.
			printVerbose("Token will expire before next reconciliation.", o.verbose)
			return withBackoff(1, maxTries-tries, create).(tokenCreator)(true, tries-1)
		}

		return []byte(token.AccessToken), &token.Expiry, nil
	}

	return create, nil
}

// createOrUpdateSecret creates or updates a kubernetes secrets.
func createOrUpdateSecret(o options, client v1.SecretsGetter, ns string, secretData []byte, expiration *time.Time) (*corev1.Secret, error) {
	data, err := generateTokenData(o, secretData, expiration)
	if err != nil {
		return nil, err
	}

	req := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      o.secret,
			Namespace: ns,
		},
		Data: map[string][]byte{o.key: data},
	}

	printVerbose(fmt.Sprintf("md5sum of token created is: %x", md5.Sum(data)), o.verbose)
	if secret, err := client.Secrets(ns).Create(req); err == nil {
		printVerbose(fmt.Sprintf("creating secret: %v in namespace: %v\n", o.secret, ns), o.verbose)
		return secret, nil
	} else if secret, err := client.Secrets(ns).Update(req); err == nil {
		printVerbose(fmt.Sprintf("updating secret: %v in namespace: %v\n", o.secret, ns), o.verbose)
		return secret, nil
	} else {
		return nil, err
	}
}

// getSecretCreator returns a function that creates a kubernetes secret(s).
func getSecretCreator(o options, create tokenCreator) (secretCreator, error) {
	config, err := loadClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	client := clientset.CoreV1()

	return func() ([]*corev1.Secret, namespacedErrors) {
		var (
			secrets []*corev1.Secret
			errs    namespacedErrors
		)

		for _, ns := range o.namespace {
			if secretData, expiration, err := create(o.forceRefresh, maxTries); err != nil {
				errs = append(errs, &namespacedError{ns, err.Error()})
			} else if secret, err := createOrUpdateSecret(o, client, ns, secretData, expiration); err != nil {
				errs = append(errs, &namespacedError{ns, err.Error()})
			} else {
				secrets = append(secrets, secret)
			}
		}

		return secrets, errs
	}, nil
}

// reconcile runs a reconciliation loop in order to achieve desired secret state.
func reconcile(o options, create secretCreator) {
	ticker := time.NewTicker(o.interval)

	work := func() {
		secrets, errs := create()

		printVerbose(fmt.Sprintf(
			"%v: reconcile complete; secrets: %v; errors: %v; next reconcile: %vm\n",
			timeNow().Format(time.RFC3339),
			len(secrets),
			len(errs),
			o.interval.Minutes(),
		), true)

		if len(errs) > 0 {
			printVerbose(fmt.Sprintf("errors: %v\n", errs.Errors()), o.verbose)
		}
	}

	for ; true; <-ticker.C {
		work()
	}
}

// main entry point.
func main() {
	var o options

	o.parseFlags()

	if err := o.validateFlags(); err != nil {
		printErrAndExit(err, 1)
	}

	tokenCreator, err := getOauthTokenCreator(o)
	if err != nil {
		printErrAndExit(err, 1)
	}

	secretCreator, err := getSecretCreator(o, tokenCreator)
	if err != nil {
		printErrAndExit(err, 1)
	}

	reconcile(o, secretCreator)
}
