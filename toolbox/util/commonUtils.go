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
	"io/ioutil"
	"log"
	"os/exec"
	"strings"
)

// ReadFile reads the file on the given path and
// returns its content as a string
func ReadFile(filePath string) (string, error) {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// WriteFile overwrites the file on the given path with content
func WriteFile(filePath, content string) error {
	if len(content) > 0 && content[len(content)-1] != '\n' {
		content += "\n"
	}
	return ioutil.WriteFile(filePath, []byte(content), 0600)
}

// ContainsString finds if target presents in the given slice
func ContainsString(slice []string, target string) bool {
	for _, element := range slice {
		if element == target {
			return true
		}
	}
	return false
}

// UpdateKeyValueInFile updates in the file all occurrences of key to
// a new value. A key-value pair is defined as `key="value"` or `key = "value"`
func UpdateKeyValueInFile(file, key, value string) error {
	replaceValue := func(line *string, value string) {
		idx := strings.Index(*line, "\"")
		*line = (*line)[:idx] + "\"" + value + "\""
	}

	input, err := ReadFile(file)
	if err != nil {
		return err
	}
	lines := strings.Split(input, "\n")
	found := false
	for i, line := range lines {
		if strings.Contains(line, key+" = ") || strings.Contains(line, key+"=") {
			replaceValue(&lines[i], value)
			found = true
		}
	}
	if !found {
		return fmt.Errorf("no occurrence of %s found in %s", key, file)
	}
	output := strings.Join(lines, "\n")
	return WriteFile(file, output)
}

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
