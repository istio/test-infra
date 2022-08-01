// Copyright 2020 Istio Authors
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
	"log"

	"github.com/hashicorp/go-multierror"
	"github.com/imdario/mergo"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/test-infra/prow/config"

	"istio.io/test-infra/tools/prowgen/pkg/spec"
)

func ApplyRequirements(job *config.JobBase, requirements, excludedRequirements []string,
	presetMap map[string]spec.RequirementPreset,
) {
	validRequirements := sets.NewString()
	for name := range presetMap {
		validRequirements = validRequirements.Insert(name)
	}
	var err error
	for _, req := range requirements {
		if e := validate(
			req,
			validRequirements,
			"requirements"); e != nil {
			err = multierror.Append(err, e)
		}
	}
	for _, req := range excludedRequirements {
		if e := validate(
			req,
			validRequirements,
			"excluded_requirements"); e != nil {
			err = multierror.Append(err, e)
		}
	}
	if err != nil {
		log.Fatalf("Requirements validation failed: %v", err)
	}

	blocked := sets.NewString(excludedRequirements...)
	presets := make([]spec.RequirementPreset, 0)
	for _, req := range requirements {
		if !blocked.Has(req) {
			presets = append(presets, presetMap[req])
		}
	}
	resolveRequirements(job.Annotations, job.Labels, job.Spec, presets)
}

func resolveRequirements(annotations, labels map[string]string, spec *v1.PodSpec, requirements []spec.RequirementPreset) {
	if spec != nil {
		for _, req := range requirements {
			mergeRequirement(annotations, labels, spec, spec.Containers, &spec.Volumes, req)
		}
	}
}

// mergeRequirement will overlay the requirement on the existing job spec. Use mergo for all keys except containers and metadata
func mergeRequirement(annotations, labels map[string]string, spec *v1.PodSpec, containers []v1.Container, volumes *[]v1.Volume,
	req spec.RequirementPreset) {
	for a, v := range req.Annotations {
		annotations[a] = v
	}
	for l, v := range req.Labels {
		labels[l] = v
	}
	for i := range containers {
		containers[i].Args = append(containers[i].Args, req.Args...)
	}
	for _, e1 := range req.Env {
		for i := range containers {
			exists := false
			for _, e2 := range containers[i].Env {
				if e2.Name == e1.Name {
					exists = true
					break
				}
			}
			if !exists {
				containers[i].Env = append(containers[i].Env, e1)
			}
		}
	}
	for _, vl1 := range req.Volumes {
		exists := false
		for _, vl2 := range *volumes {
			if vl2.Name == vl1.Name {
				exists = true
				break
			}
		}
		if !exists {
			*volumes = append(*volumes, vl1)
		}
	}
	for _, vm1 := range req.VolumeMounts {
		for i := range containers {
			exists := false
			for _, vm2 := range containers[i].VolumeMounts {
				if vm2.MountPath == vm1.MountPath {
					exists = true
					break
				}
			}
			if !exists {
				containers[i].VolumeMounts = append(containers[i].VolumeMounts, vm1)
			}
		}
	}

	if req.PodSpec != nil {
		if err := mergo.Merge(spec, req.PodSpec); err != nil {
			log.Fatalf("Unable to merge PodSpec: %v", err)
		}
	}
}
