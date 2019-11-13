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
	"path/filepath"
	"regexp"
	"strings"
)

// GetTopLevelOrg escapes and returns the top-level org from an org string.
func GetTopLevelOrg(s string) string {
	m := regexp.MustCompile(`^http(?:s)://(.+?)(?:/(.+))?$`).FindStringSubmatch(s)

	if len(m) == 2 {
		return strings.Replace(m[1], "/", "-", -1)
	}
	if len(m) == 3 {
		return filepath.Join(strings.Replace(m[1], "/", "-", -1), m[2])
	}

	return s
}

// SplitOrgRepo splits and org/repo string into into two separate strings.
func SplitOrgRepo(s string) (string, string) {
	m := regexp.MustCompile(`^((?:http(?:s)://)?.+)/(.+)$`).FindStringSubmatch(s)

	return m[1], m[2]
}

// RemoveHost removes a host prefix from a string.
func RemoveHost(s string) string {
	return regexp.MustCompile("^http(?:s)?://.+/?$").ReplaceAllString(s, "")
}
