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

package pkg

import (
	v1 "k8s.io/api/core/v1"
)

// RequirementPreset can be used to re-use settings across multiple jobs.
type RequirementPreset struct {
	Annotations  map[string]string `json:"annotations"`
	Labels       map[string]string `json:"labels"`
	Env          []v1.EnvVar       `json:"env"`
	Volumes      []v1.Volume       `json:"volumes"`
	VolumeMounts []v1.VolumeMount  `json:"volumeMounts"`
	Args         []string          `json:"args"`
}

func (r RequirementPreset) DeepCopy() RequirementPreset {
	ret := RequirementPreset{
		Annotations: map[string]string{},
		Labels:      map[string]string{},
	}
	for k, v := range r.Annotations {
		ret.Annotations[k] = v
	}
	for k, v := range r.Labels {
		ret.Labels[k] = v
	}
	ret.Env = append(ret.Env, r.Env...)
	ret.Volumes = append(ret.Volumes, r.Volumes...)
	ret.VolumeMounts = append(ret.VolumeMounts, r.VolumeMounts...)
	ret.Args = append(ret.Args, r.Args...)
	return ret
}

func resolveRequirements(annotations, labels map[string]string, spec *v1.PodSpec, requirements []RequirementPreset) {
	if spec != nil {
		for _, req := range requirements {
			mergeRequirement(req, annotations, labels, spec.Containers, &spec.Volumes)
		}
	}
}

// mergeRequirement will overlay the requirement on the existing job spec.
func mergeRequirement(req RequirementPreset, annotations, labels map[string]string, containers []v1.Container, volumes *[]v1.Volume) {
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
}
