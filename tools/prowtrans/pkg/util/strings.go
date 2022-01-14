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
	"sort"
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
	return regexp.MustCompile("^https?://.*?(?:/(.*)|$)").ReplaceAllString(s, "$1")
}

// NormalizeOrg removes a host prefix from a string.
func NormalizeOrg(s, sep string) string {
	s = RemoveHost(s)
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "/")
	s = strings.ReplaceAll(s, "/", sep)
	return s
}

// SortedKeys returns a sorted list of keys for a given map.
func SortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))

	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}
