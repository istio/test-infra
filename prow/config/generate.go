package main

import (
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/hashicorp/go-multierror"
	"github.com/kr/pretty"
	"io/ioutil"
	"k8s.io/api/core/v1"
	"k8s.io/test-infra/prow/config"
	"k8s.io/test-infra/prow/kube"
	"os"
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
)

type JobConfig struct {
	Jobs      []Job                              `json:"jobs"`
	Repo      string                             `json:"repo"`
	Branches  []string                           `json:"branches"`
	Resources map[string]v1.ResourceRequirements `json:"resources"`
}

type Job struct {
	Name      string   `json:"name"`
	Command   []string `json:"command"`
	Resources string   `json:"resources"`
}

func main() {
	jobs := readJobConfig("jobs.yaml")
	validateConfig(jobs)
	result := convertJobConfig(jobs)
	writeConfig(result)

	diffConfig(result)
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
	}
	if err != nil {
		exit(err, "validation failed")
	}
}

func diffConfig(result config.JobConfig) {
	pj := readProwJobConfig("../cluster/jobs/istio/istio/istio.istio.master.yaml")
	for _, job := range result.AllPresubmits([]string{"istio/istio"}) {
		current := pj.GetPresubmit("istio/istio", job.Name)
		if current == nil {
			fmt.Println("Could not find job", job.Name)
			continue
		}
		diff := pretty.Diff(current, &job)
		fmt.Println("Diff for", job.Name)
		for _, d := range diff {
			fmt.Println(d)
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
					Name:  fmt.Sprintf("%s-%s", job.Name, branch),
					Agent: string(kube.KubernetesAgent),
					Spec: &v1.PodSpec{
						NodeSelector: map[string]string{"testing": "test-pool"},
						Containers:   createContainer(jobConfig, job),
					},
					UtilityConfig: config.UtilityConfig{
						Decorate:  true,
						PathAlias: "istio.io/istio",
					},
				},
				AlwaysRun: true,
				Brancher: config.Brancher{
					Branches: []string{fmt.Sprintf("^%s$", branch)},
				},
			}
			result.Presubmits[jobConfig.Repo] = append(result.Presubmits[jobConfig.Repo], presubmit)
		}
	}
	return result
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
