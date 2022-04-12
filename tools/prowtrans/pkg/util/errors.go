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

package util

import (
	"fmt"
	"os"
)

// ExitError is a custom error type which stores a message and status code.
type ExitError struct {
	Code    int
	Message string
}

func (err ExitError) Error() string {
	return err.Message
}

// PrintErr prints an error message to stderr.
func PrintErr(msg string) {
	_, _ = fmt.Fprintln(os.Stderr, msg)
}

// PrintErrAndExit prints an error message to stderr and exits with a status code.
func PrintErrAndExit(err error) {
	PrintErr(err.Error())

	exitErr, ok := err.(*ExitError)
	if ok {
		os.Exit(exitErr.Code)
	}
	os.Exit(1)
}
