package main

import (
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/hashicorp/go-multierror"
	"github.com/kr/pretty"
	"io/ioutil"
	"k8s.io/api/core/v1"
	"k8s.io/test-infra/prow/config"
	"os"
	"strings"
)

func exit(err error, context string) {
	_, _ = fmt.Fprint(os.Stderr, fmt.Sprintf("%v: %v", context, err))
	os.Exit(1)
}

func writeConfig(c interface{}) {
	bytes, err := yaml.Marshal(c)
	if err != nil {
		exit(err, "failed to write result")
	}
	fmt.Println(string(bytes))
}

const (
	DefaultResource = "default"

	ModifierHidden   = "hidden"
	ModifierOptional = "optional"
	ModifierSkipped  = "skipped"

	RequirementRoot   = "root"
	RequirementKind   = "kind"
	RequirementBoskos = "boskos"
)

type JobConfig struct {
	Jobs      []Job                              `json:"jobs"`
	Repo      string                             `json:"repo"`
	Branches  []string                           `json:"branches"`
	Resources map[string]v1.ResourceRequirements `json:"resources"`
}

type Job struct {
	Name         string   `json:"name"`
	Command      []string `json:"command"`
	Resources    string   `json:"resources"`
	Modifiers    []string `json:"modifiers"`
	Requirements []string `json:"requirements"`
}

func main() {
	jobs := readJobConfig("jobs.yaml")
	validateConfig(jobs)
	result := convertJobConfig(jobs)
	writeConfig(result)

	diffConfig(result)
}

func validate(input string, options []string, description string) error {
	valid := false
	for _, opt := range options {
		if input == opt {
			valid = true
		}
	}
	if !valid {
		return fmt.Errorf("'%v' is not a valid %v. Must be one of %v", input, description, strings.Join(options, ","))
	}
	return nil
}

func validateConfig(jobConfig JobConfig) {
	var err error
	if _, f := jobConfig.Resources[DefaultResource]; !f {
		err = multierror.Append(err, fmt.Errorf("'%v' resource must be provided", DefaultResource))
	}
	for _, job := range jobConfig.Jobs {
		if job.Resources != "" {
			if _, f := jobConfig.Resources[job.Resources]; !f {
				err = multierror.Append(err, fmt.Errorf("job '%v' has nonexistant resource '%v'", job.Name, job.Resources))
			}
		}
		for _, mod := range job.Modifiers {
			if e := validate(mod, []string{ModifierHidden, ModifierOptional, ModifierSkipped}, "status"); e != nil {
				err = multierror.Append(err, e)
			}
		}
		for _, req := range job.Requirements {
			if e := validate(req, []string{RequirementKind, RequirementRoot, RequirementBoskos}, "requirements"); e != nil {
				err = multierror.Append(err, e)
			}
		}
	}
	if err != nil {
		exit(err, "validation failed")
	}
}

func diffConfig(result config.JobConfig) {
	pj := readProwJobConfig("../cluster/jobs/istio/istio/istio.istio.master.yaml")
	known := make(map[string]struct{})
	for _, job := range result.AllPresubmits([]string{"istio/istio"}) {
		known[job.Name] = struct{}{}
		current := pj.GetPresubmit("istio/istio", job.Name)
		if current == nil {
			fmt.Println("Could not find job", job.Name)
			continue
		}
		diff := pretty.Diff(current, &job)
		fmt.Println("\nDiff for", job.Name)
		for _, d := range diff {
			fmt.Println(d)
		}
	}
	for _, job := range pj.AllPresubmits([]string{"istio/istio"}) {
		if _, f := known[job.Name]; !f {
			fmt.Println("Missing", job.Name)
		}
	}

}

func createContainer(config JobConfig, job Job) []v1.Container {
	c := v1.Container{
		Image:           "gcr.io/istio-testing/istio-builder:v20190709-959ee177",
		SecurityContext: &v1.SecurityContext{Privileged: newTrue()},
		Command:         job.Command,
	}
	resource := DefaultResource
	if job.Resources != "" {
		resource = job.Resources
	}
	c.Resources = config.Resources[resource]

	return []v1.Container{c}
}

func convertJobConfig(jobConfig JobConfig) config.JobConfig {
	result := config.JobConfig{
		Presubmits:  make(map[string][]config.Presubmit),
		Postsubmits: make(map[string][]config.Postsubmit),
	}
	for _, job := range jobConfig.Jobs {
		for _, branch := range jobConfig.Branches {
			job.Command = append([]string{"entrypoint"}, job.Command...)
			presubmit := config.Presubmit{
				JobBase: config.JobBase{
					Name: fmt.Sprintf("%s-%s", job.Name, branch),
					Spec: &v1.PodSpec{
						NodeSelector: map[string]string{"testing": "test-pool"},
						Containers:   createContainer(jobConfig, job),
					},
					UtilityConfig: config.UtilityConfig{
						Decorate:  true,
						PathAlias: "istio.io/istio",
					},
					Labels: make(map[string]string),
				},
				AlwaysRun: true,
				Brancher: config.Brancher{
					Branches: []string{fmt.Sprintf("^%s$", branch)},
				},
			}
			for _, modifier := range job.Modifiers {
				applyModifier(&presubmit, modifier)
			}
			applyRequirements(&presubmit, job.Requirements)
			result.Presubmits[jobConfig.Repo] = append(result.Presubmits[jobConfig.Repo], presubmit)
		}
	}
	return result
}

func applyRequirements(presubmit *config.Presubmit, requirements []string) {
	for _, req := range requirements {
		switch req {
		case RequirementBoskos:
			presubmit.MaxConcurrency = 5
			presubmit.Labels["preset-service-account"] = "true"
		case RequirementRoot:
			presubmit.JobBase.Spec.Containers[0].SecurityContext.Privileged = newTrue()
		case RequirementKind:
			dir := v1.HostPathDirectory
			presubmit.JobBase.Spec.Volumes = append(presubmit.JobBase.Spec.Volumes,
				v1.Volume{
					Name: "modules",
					VolumeSource: v1.VolumeSource{
						HostPath: &v1.HostPathVolumeSource{
							Path: "/lib/modules",
							Type: &dir,
						},
					},
				},
				v1.Volume{
					Name: "cgroup",
					VolumeSource: v1.VolumeSource{
						HostPath: &v1.HostPathVolumeSource{
							Path: "/sys/fs/cgroup",
							Type: &dir,
						},
					},
				},
			)
			presubmit.JobBase.Spec.Containers[0].VolumeMounts = append(presubmit.JobBase.Spec.Containers[0].VolumeMounts,
				v1.VolumeMount{
					MountPath: "/lib/modules",
					Name:      "modules",
					ReadOnly:  true,
				},
				v1.VolumeMount{
					MountPath: "/sys/fs/cgroup",
					Name:      "cgroup",
				},
			)
		}
	}
}

func applyModifier(presubmit *config.Presubmit, jobModifier string) {
	if jobModifier == ModifierOptional {
		presubmit.Optional = true
	} else if jobModifier == ModifierHidden {
		presubmit.SkipReport = true
	} else if jobModifier == ModifierSkipped {
		presubmit.AlwaysRun = false
	}
}

func readProwJobConfig(file string) config.JobConfig {
	yamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		exit(err, "failed to read "+file)
	}
	jobs := config.JobConfig{}
	if err := yaml.Unmarshal(yamlFile, &jobs); err != nil {
		exit(err, "failed to unmarshal "+file)
	}
	return jobs
}

func readJobConfig(file string) JobConfig {
	yamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		exit(err, "failed to read "+file)
	}
	jobs := JobConfig{}
	if err := yaml.Unmarshal(yamlFile, &jobs); err != nil {
		exit(err, "failed to unmarshal "+file)
	}
	return jobs
}

func newTrue() *bool {
	b := true
	return &b
}
