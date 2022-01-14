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
	"os"
	"path/filepath"
	"regexp"
)

// RenameFile renames a file based on a specified regular expression pattern.
func RenameFile(pat string, src string, repl string) string {
	s := regexp.MustCompile(pat).ReplaceAllString(src, repl)
	return regexp.MustCompile(`^[^\w\d]`).ReplaceAllString(s, "")
}

// HasExtension checks if a file's extension matches a pattern.
func HasExtension(path string, pat string) bool {
	return regexp.MustCompile(pat).MatchString(filepath.Ext(path))
}

// Exists checks if a path exists.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// IsFile checks if a path exists and is a file.
func IsFile(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.Mode().IsRegular()
}

// IsDirectory checks if a path exists and is a directory.
func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.Mode().IsDir()
}
