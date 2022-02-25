// Copyright Istio Authors
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

package decorator

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
)

func validate(input string, options sets.String, description string) error {
	if !options.Has(input) {
		return fmt.Errorf("'%v' is not a valid %v. Must be one of %v", input, description, strings.Join(options.List(), ", "))
	}
	return nil
}
