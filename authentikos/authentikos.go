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
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	flag "github.com/spf13/pflag"
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

const (
	secretKey        = "token"                 // secretKey is the kubernetes token secret key.
	defaultSecret    = "authentikos-token"     // defaultSecret is the default kubernetes secret name.
	defaultFormat    = "%v"                    // defaultFormat is the default token format string.
	defaultNamespace = metav1.NamespaceDefault // defaultNamespace is the default kubernetes namespace.
	tickInterval     = 30 * time.Minute        // tickInterval is the default tick interval.
)

// tokenCreator is a function that creates an oauth token.
type tokenCreator func() ([]byte, error)

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

// options are the available command-line flags.
type options struct {
	verbose   bool
	creds     string
	format    string
	secret    string
	namespace []string
	scopes    []string
}

// parseFlags parses the command-line flags.
func (o *options) parseFlags() {
	flag.BoolVarP(&o.verbose, "verbose", "v", false, "Print verbose output.")
	flag.StringVarP(&o.creds, "creds", "c", "", "Path to a JSON credentials file.")
	flag.StringVarP(&o.format, "format", "f", defaultFormat, "Format string for the token.")
	flag.StringVarP(&o.secret, "secret", "o", defaultSecret, "Name of secret to create.")
	flag.StringSliceVarP(&o.namespace, "namespace", "n", []string{defaultNamespace}, "Namespace(s) to create the secret in.")
	flag.StringSliceVarP(&o.scopes, "scopes", "s", []string{}, "Oauth scope(s) to request for token.")

	flag.Parse()
}

// validateFlags validates the command-line flags.
func (o *options) validateFlags() error {
	var err error

	if len(o.format) == 0 {
		o.format = defaultFormat
	}

	if len(o.secret) == 0 {
		o.secret = defaultSecret
	}

	if len(o.namespace) == 0 {
		o.namespace = []string{defaultNamespace}
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
		fmt.Print(formatString)
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
	clientOpts := []option.ClientOption{option.WithScopes(o.scopes...)}

	if len(o.creds) > 0 {
		clientOpts = append(clientOpts, option.WithCredentialsFile(o.creds))
	}

	client, err := transport.Creds(context.Background(), clientOpts...)
	if err != nil {
		return nil, err
	}

	return func() ([]byte, error) {
		token, err := client.TokenSource.Token()
		if err != nil {
			return nil, err
		}
		return []byte(token.AccessToken), nil
	}, nil
}

// createOrUpdateSecret creates or updates a kubernetes secrets.
func createOrUpdateSecret(o options, client v1.SecretsGetter, ns string, secretData []byte) (*corev1.Secret, error) {
	req := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      o.secret,
			Namespace: ns,
		},
		StringData: map[string]string{secretKey: fmt.Sprintf(o.format, string(secretData))},
	}

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
			if secretData, err := create(); err != nil {
				errs = append(errs, &namespacedError{ns, err.Error()})
			} else if secret, err := createOrUpdateSecret(o, client, ns, secretData); err != nil {
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
	ticker := time.NewTicker(tickInterval)

	work := func() {
		secrets, errs := create()

		printVerbose(fmt.Sprintf(
			"%v: reconcile complete; secrets: %v; errors: %v; next reconcile: %vm\n",
			time.Now().Format(time.RFC3339),
			len(secrets),
			len(errs),
			tickInterval.Minutes(),
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
