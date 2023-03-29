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
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/pflag"
)

const (
	testDir = "testdata"
)

func resolvePath(t *testing.T, filename string) string {
	name := strings.ToLower(filepath.Base(t.Name()))
	return filepath.Join(testDir, strings.ToLower(name), name+filename)
}

func parseConfigTmpl(input, output, config, dir string) (string, error) {
	var b bytes.Buffer

	cfg, err := os.ReadFile(config)
	if err != nil {
		return "", fmt.Errorf("failed reading config file %v: %v", config, err)
	}

	tmpl, err := template.New("test").Parse(string(cfg))
	if err != nil {
		return "", fmt.Errorf("failed parsing config template %v: %v", config, err)
	}

	if err := tmpl.Execute(&b, struct {
		Input  string
		Output string
	}{
		Input:  input,
		Output: output,
	}); err != nil {
		return "", fmt.Errorf("failed executing config template %v: %v", config, err)
	}

	cfgO := filepath.Join(dir, "cfg.yaml")

	if err := os.WriteFile(cfgO, b.Bytes(), 0o644); err != nil {
		return "", fmt.Errorf("failed writing config file %v: %v", cfgO, err)
	}

	return cfgO, nil
}

func TestProwTrans(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		args    []string
		configs bool
	}{
		{
			name: "simple transform",
			args: []string{"--mapping=istio=istio-private"},
		},
		{
			name: "branches-out",
			args: []string{"--mapping=istio=istio-private", "--branches-out=custom-1,^custom-2$"},
		},
		{
			name: "refs exists",
			args: []string{"--mapping=istio=istio-private", "--refs"},
		},
		{
			name: "refs not exists",
			args: []string{"--mapping=istio=istio-private", "--refs"},
		},
		{
			name: "rerun-orgs",
			args: []string{"--mapping=istio=istio-private", "--rerun-orgs=istio-private,istio-secret"},
		},
		{
			name: "rerun-users",
			args: []string{"--mapping=istio=istio-private", "--rerun-users=clarketm,scoobydoo"},
		},
		{
			name: "override annotations",
			args: []string{"--mapping=istio=istio-private", "--annotations=testgrid-create-test-group=false"},
		},
		{
			name: "sort ascending",
			args: []string{"--mapping=istio=istio-private", "--sort=asc"},
		},
		{
			name: "sort descending",
			args: []string{"--mapping=istio=istio-private", "--sort=desc"},
		},
		{
			name: "env denylist",
			args: []string{"--mapping=istio=istio-private", "--env-denylist=bad-env"},
		},
		{
			name: "volume denylist",
			args: []string{"--mapping=istio=istio-private", "--volume-denylist=bad-volume"},
		},
		{
			name:    "config file",
			configs: true,
		},
		{
			name:    "image_rename",
			configs: true,
		},
		{
			name:    "tag_rename",
			configs: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			in := resolvePath(t, "_in.yaml")
			outE := resolvePath(t, "_out.yaml")

			expected, err := os.ReadFile(outE)
			if err != nil {
				t.Fatalf("failed reading expected output file %v: %v", outE, err)
			}

			tmpDir, err := os.MkdirTemp("", "")
			if err != nil {
				t.Fatalf("failed creating temp file: %v", err)
			}
			defer os.Remove(tmpDir)
			outA := filepath.Join(tmpDir, "out.yaml")

			os.Args = []string{"prowtrans"}
			pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
			os.Args = append(os.Args, test.args...)
			if test.configs {
				cfg, err := parseConfigTmpl(in, outA, resolvePath(t, "_cfg.yaml"), tmpDir)
				if err != nil {
					t.Fatal(err)
				}
				os.Args = append(os.Args, "--configs="+cfg)
			} else {
				os.Args = append(os.Args, "--input="+in, "--output="+outA)
			}
			Main()

			actual, err := os.ReadFile(outA)
			if err != nil {
				t.Fatalf("failed reading actual output file %v: %v", outA, err)
			}

			if os.Getenv("REFRESH_GOLDEN") == "true" {
				if err = os.WriteFile(outE, actual, 0o644); err != nil {
					t.Fatalf("failed writing expected output file %v: %v", outE, err)
				}
				expected = actual
			}

			if diff := cmp.Diff(expected, actual); diff != "" {
				t.Error("TestProwTrans (-want, +got):", diff)
			}
		})
	}
}
