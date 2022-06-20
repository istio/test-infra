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
	"log"
	"regexp"
	"strings"

	"sigs.k8s.io/yaml"

	"istio.io/test-infra/tools/prowgen/pkg/spec"
)

const (
	matrixPrefix = "matrix."
	paramsPrefix = "params."
)

var variableSubstitutionRegex = regexp.MustCompile(`\$\([_a-zA-Z0-9.-]+(\.[_a-zA-Z0-9.-]+)*\)`)

func applyArch(arch string, job spec.Job) spec.Job {
	// For backwards compatibility, amd64 is not suffixed
	if arch != "amd64" {
		job.Name += "-" + arch
	}

	if job.NodeSelector == nil {
		job.NodeSelector = map[string]string{}
	}
	job.NodeSelector["kubernetes.io/arch"] = arch
	return job
}

func ApplyVariables(job spec.Job, architectures []string, params map[string]string, matrix map[string][]string) []spec.Job {
	yamlBS, err := yaml.Marshal(job)
	if err != nil {
		log.Fatalf("Failed to marshal the given Job: %v", err)
	}

	jobs := make([]spec.Job, 0)

	for _, arch := range architectures {
		subsExps := getVarSubstitutionExpressions(string(yamlBS))
		if len(subsExps) == 0 && len(architectures) == 1 {
			jobs = append(jobs, applyArch(arch, job))
			continue
		}
		if params == nil {
			params = map[string]string{}
		}
		params["arch"] = arch

		resolvedYAMLStr := applyParams(string(yamlBS), subsExps, params)
		resolvedYAMLStrs := applyMatrix(resolvedYAMLStr, subsExps, matrix)

		for _, jobYaml := range resolvedYAMLStrs {
			job := spec.Job{}
			if err := yaml.Unmarshal([]byte(jobYaml), &job); err != nil {
				log.Fatalf("Failed to unmarshal the yaml to Job: %v", err)
			}
			jobs = append(jobs, applyArch(arch, job))
		}
	}
	return jobs
}

// applyParams will resolve all the $(params.key) expressions into the
// configured values.
func applyParams(yamlStr string, subsExps []string, params map[string]string) string {
	for _, exp := range subsExps {
		if strings.HasPrefix(exp, paramsPrefix) {
			exp = strings.TrimPrefix(exp, paramsPrefix)
			if val, ok := params[exp]; ok {
				yamlStr = replace(yamlStr, paramsPrefix, exp, val)
			} else {
				log.Fatalf("Param %q not configured in the params map %v", exp, params)
			}
		}
	}
	return yamlStr
}

// applyMatrix will resolve all the $(matrix.dimension) expressions into the
// configured lists of values, and then calculate all the combinations.
func applyMatrix(yamlStr string, subsExps []string, matrix map[string][]string) []string {
	combs := make([]string, 0)
	for _, exp := range subsExps {
		if strings.HasPrefix(exp, matrixPrefix) {
			exp = strings.TrimPrefix(exp, matrixPrefix)
			if _, ok := matrix[exp]; ok {
				combs = append(combs, exp)
			} else {
				log.Fatalf("Dimension %q not configured in the matrix %v", exp, matrix)
			}
		}
	}

	res := &[]string{}
	resolveCombinations(combs, yamlStr, 0, matrix, res)
	return *res
}

func resolveCombinations(combs []string, dest string, start int, matrix map[string][]string, res *[]string) {
	if start == len(combs) {
		*res = append(*res, dest)
		return
	}

	lst := matrix[combs[start]]
	for i := range lst {
		dest := replace(dest, matrixPrefix, combs[start], lst[i])
		resolveCombinations(combs, dest, start+1, matrix, res)
	}
}

// replace replaces the expressions written as $(prefix.expKey) with the expVal
func replace(str, expType, expKey, expVal string) string {
	return strings.ReplaceAll(str, fmt.Sprintf("$(%s%s)", expType, expKey), expVal)
}

// getVarSubstitutionExpressions extracts all the value between "$(" and ")""
func getVarSubstitutionExpressions(yamlStr string) []string {
	allExpressions := validateString(yamlStr)
	return allExpressions
}

func validateString(value string) []string {
	expressions := variableSubstitutionRegex.FindAllString(value, -1)
	if expressions == nil {
		return nil
	}
	var result []string
	set := map[string]bool{}
	for _, expression := range expressions {
		expression = stripVarSubExpression(expression)
		if _, ok := set[expression]; !ok {
			result = append(result, expression)
			set[expression] = true
		}
	}
	return result
}

func stripVarSubExpression(expression string) string {
	return strings.TrimSuffix(strings.TrimPrefix(expression, "$("), ")")
}
