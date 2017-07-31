// Copyright 2017 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

// GetMD5Hash generates an MD5 digest of the given string
func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

// Shell runs command on shell and get back output and error if get one
func Shell(format string, args ...interface{}) (string, error) {
	return sh(format, false, args...)
}

// ShellSilent runs command on shell without logging the exact command
// useful when command involves secrets
func ShellSilent(format string, args ...interface{}) (string, error) {
	return sh(format, true, args...)
}

// Runs command on shell and get back output and error if get one
func sh(format string, muted bool, args ...interface{}) (string, error) {
	command := fmt.Sprintf(format, args...)
	parts := strings.Split(command, " ")
	if !muted {
		log.Printf("Running command %s", command)
	}
	c := exec.Command(parts[0], parts[1:]...) // #nosec
	bytes, err := c.CombinedOutput()
	log.Printf("Command output: \n%s", string(bytes[:]))
	if err != nil {
		return "", fmt.Errorf("command failed: %q %v", string(bytes), err)
	}
	return string(bytes), nil
}
