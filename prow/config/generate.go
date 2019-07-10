package main

import (
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/kr/pretty"
	"io/ioutil"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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

type JobConfig struct {
	Jobs     []Job    `json:"jobs"`
	Repo     string   `json:"repo"`
	Branches []string `json:"branches"`
}

type Job struct {
	Name    string   `json:"name"`
	Command []string `json:"command"`
}

func main() {
	jobs := readJobConfig("jobs.yaml")
	result := convertJobConfig(jobs)
	writeConfig(result)

	diffConfig(result)
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
						Containers: []v1.Container{{
							Image:           "gcr.io/istio-testing/istio-builder:v20190709-959ee177",
							SecurityContext: &v1.SecurityContext{Privileged: newTrue()},
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									"cpu":    resource.MustParse("3000m"),
									"memory": resource.MustParse("3Gi"),
								},
							},
							Command: job.Command,
						}},
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
