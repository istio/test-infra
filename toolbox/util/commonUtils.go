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
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"
)

const (
	// ReleaseNoteNone is "none" string to indicate release-note is none
	ReleaseNoteNone = "none"
)

var (
	kvSplitters = []string{
		" = ",
		"=",
		":",
	}
)

// Poll executes do() after time interval for a max of numTrials times.
// The bool returned by do() indicates if polling succeeds in that trial
func Poll(interval time.Duration, numTrials int, do func() (bool, error)) error {
	if numTrials < 0 {
		return fmt.Errorf("numTrials cannot be negative")
	}
	for i := 0; i < numTrials; i++ {
		if success, err := do(); err != nil {
			return fmt.Errorf("error during trial %d: %v", i, err)
		} else if success {
			return nil
		} else {
			time.Sleep(interval)
		}
	}
	return fmt.Errorf("max polling iteration reached")
}

// ReadFile reads the file on the given path and
// returns its content as a string
func ReadFile(filePath string) (string, error) {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// WriteTextFile overwrites the file on the given path with content
func WriteTextFile(filePath, content string) error {
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

// updateKeyValueInTomlLines updates all occurrences of key to a new value
func updateKeyValueInTomlLines(lines []string, key, value string) ([]string, bool) {
	// toml dependecies are of the form
	//  name = "istio.io/api"
	//  revision = "b08011c721e03edd61c721e4943607c97b7a9879"

	found := false
	keySearch := fmt.Sprintf("name = %q", key)
	for i, line := range lines {
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		if strings.Contains(line, keySearch) {
			lines[i+1] = fmt.Sprintf("  revision = %q", value)
			found = true
		}
	}
	return lines, found
}

// updateKeyValueInLines updates all occurrences of key to
// a new value, except comments started with # or //
func updateKeyValueInLines(lines []string, key, value string) ([]string, bool) {
	replaceValue := func(line *string, splitter string) {
		idx := strings.Index(*line, splitter) + len(splitter)
		if (*line)[idx] == '"' {
			*line = (*line)[:idx] + "\"" + value + "\""
		} else {
			*line = (*line)[:idx] + value
		}
	}

	found := false
	for i, line := range lines {
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		for _, splitter := range kvSplitters {
			if strings.Contains(line, key+splitter) {
				replaceValue(&lines[i], splitter)
				found = true
				break
			}
		}
	}
	return lines, found
}

// UpdateKeyValueInFile updates in the file all occurrences of key to
// a new value, except comments started with # or //
func UpdateKeyValueInFile(file, key, value string) error {
	input, err := ReadFile(file)
	if err != nil {
		return err
	}
	var found, foundToml bool
	lines := strings.Split(input, "\n")
	lines, found = updateKeyValueInLines(lines, key, value)
	if file == "Gopkg.toml" {
		lines, foundToml = updateKeyValueInTomlLines(lines, key, value)
	}
	if !found && !foundToml {
		return fmt.Errorf("no occurrence of %s found in file %s", key, file)
	}
	output := strings.Join(lines, "\n")
	return WriteTextFile(file, output)
}

// GetMD5Hash generates an MD5 digest of the given string
func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

// Shell runs command on shell and get back output and error if get one,
// it takes a set of environment vaiables that are appended to existing environment
func Shell(env []string, format string, args ...interface{}) (string, error) {
	return sh(env, format, false, args...)
}

// ShellSilent runs command on shell without logging the exact command
// it takes a set of environment vaiables that are appended to existing environment
// useful when command involves secrets
func ShellSilent(env []string, format string, args ...interface{}) (string, error) {
	return sh(env, format, true, args...)
}

// Runs command on shell and get back output and error if get one
func sh(env []string, format string, muted bool, args ...interface{}) (string, error) {
	command := fmt.Sprintf(format, args...)
	parts := strings.Split(command, " ")
	if !muted {
		log.Printf("Running command %s", command)
	}
	c := exec.Command(parts[0], parts[1:]...) // #nosec
	c.Env = append(os.Environ(), env...)
	bytes, err := c.CombinedOutput()
	log.Printf("Command output: \n%s", string(bytes[:]))
	if err != nil {
		return "", fmt.Errorf("command failed: %q %v", string(bytes), err)
	}
	return string(bytes), nil
}

// FillUpTemplate fills up a template from the provided interface
func FillUpTemplate(t string, i interface{}) (string, error) {
	tmpl, err := template.New("tmpl").Parse(t)
	if err != nil {
		return "", err
	}
	wr := bytes.NewBufferString("")
	err = tmpl.Execute(wr, i)
	if err != nil {
		return "", err
	}
	return wr.String(), nil
}

// AssertNotEmpty check if a value is empty, exit if value not specified
func AssertNotEmpty(name string, value *string) {
	if value == nil || *value == "" {
		log.Fatalf("%s must be specified\n", name)
	}
}
