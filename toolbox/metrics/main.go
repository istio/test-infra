// Copyright 2017 Istio Authors
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

package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/net/context"
)

const (
	fetchInterval = 5
)

var (
	webAddress = flag.String("listen_port", ":9103", "port on which to expose metrics and web interface.")
	jenkinsURL = flag.String("jenkins_url", "https://testing.istio.io", "URL to Jenkins API")
	gcsBucket  = flag.String("bucket", "istio-code-coverage", "gcs bucket")

	succeededBuilds = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "succeeded_build_durations_seconds",
			Help: "Succeeded build durations seconds.",
		},
		[]string{"build_job", "repo"},
	)

	failedBuilds = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "failed_build_durations_seconds",
			Help: "Failed build durations seconds.",
		},
		[]string{"build_job", "repo"},
	)

	succeededBuildMax = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "succeeded_build_durations_max_seconds",
			Help: "Succeeded max build durations seconds.",
		},
		[]string{"build_job", "repo"},
	)

	succeededBuildMin = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "succeeded_build_durations_min_seconds",
			Help: "Succeeded min build durations seconds.",
		},
		[]string{"build_job", "repo"},
	)

	codeCoverage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "package_code_coverage",
			Help: "Package code coverage.",
		},
		[]string{"package", "repo"},
	)

	jobRegister = map[string][]string{
		"pilot":       {"presubmit", "postsubmit", "e2e-smoketest", "e2e-suite"},
		"mixer":       {"presubmit", "postsubmit", "e2e-smoketest", "e2e-suite"},
		"proxy":       {"presubmit", "postsubmit"},
		"auth":        {"presubmit", "postsubmit"},
		"mixerclient": {"presubmit"},
		"istio":       {"presubmit"},
		"test-infra":  {"presubmit"},
	}

	codeCovRegister = map[string]bool{
		"pilot": true,
		"mixer": true,
		"auth":  true,
	}

	codeCovTrackJob = "presubmit"

	gcsClient *storage.Client

	jk = &jkConst{
		building:      "building",
		result:        "result",
		duration:      "duration",
		number:        "number",
		resultFailure: "FAILURE",
		resultSUCCESS: "SUCCESS",
	}
)

type jkConst struct {
	building      string
	result        string
	duration      string
	number        string
	resultSUCCESS string
	resultFailure string
}

type repo struct {
	name string
	jobs []*job
}

type job struct {
	repoName    string
	jobName     string
	lastBuildID int
	client      *http.Client
}

func newJob(repoName, jobName string) *job {
	return &job{
		repoName:    repoName,
		jobName:     jobName,
		lastBuildID: -1,
		client:      &http.Client{},
	}
}

func newRepo(name string) *repo {
	r := &repo{
		name: name,
		jobs: make([]*job, 0),
	}
	for n := range jobRegister[name] {
		r.jobs = append(r.jobs, newJob(r.name, jobRegister[name][n]))
	}
	return r
}

func init() {
	prometheus.MustRegister(succeededBuilds)
	prometheus.MustRegister(succeededBuildMax)
	prometheus.MustRegister(succeededBuildMin)
	prometheus.MustRegister(failedBuilds)
	prometheus.MustRegister(codeCoverage)
}

func main() {
	flag.Parse()
	var repos []*repo
	for repoName := range jobRegister {
		repos = append(repos, newRepo(repoName))
	}

	var err error
	gcsClient, err = storage.NewClient(context.Background())
	if err != nil {
		log.Printf("Failed to create a gcs client, %v", err)
	}

	go func() {
		for {
			for _, repo := range repos {
				for _, job := range repo.jobs {
					if err := job.publishCIMetrics(); err != nil {
						log.Printf("Failed to process %s/%s: %v", job.repoName, job.jobName, err)
					}
				}
			}
			time.Sleep(time.Duration(fetchInterval) * time.Minute)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	log.Print(http.ListenAndServe(*webAddress, nil))
}

func (j *job) publishCIMetrics() error {
	var max, min float64
	//Reset max/min
	max = -1
	min = math.MaxFloat64

	latestCompletedBuild, err := j.getLatestCompletedBuild()
	if err != nil {
		log.Printf("Failed to get latest complete build id from jenkins.")
		return err
	}

	if j.lastBuildID == -1 {
		j.lastBuildID = latestCompletedBuild
	}

	for i := j.lastBuildID + 1; i <= latestCompletedBuild; i++ {
		build, err := j.getJenkinsObject(fmt.Sprint(i))
		if err != nil {
			log.Printf("Failed to get build No.%d from %s/%s, %v", i, j.repoName, j.jobName, err)
		}

		b, ok := build[jk.building].(bool)
		if !ok {
			log.Printf("Building is not a bool: %s", build[jk.building])
			return errors.New("unexpected jenkins value")
		}

		if b {
			latestCompletedBuild = i - 1 //If this build is still building, set "latestCompletedBuild" to one build before
		} else {
			duration, ok := build[jk.duration].(float64)
			if !ok {
				log.Printf("Duration is not a float64: %s", build[jk.duration])
				return errors.New("unexpected jenkins value")
			}
			t := duration / 1000
			result, ok := build[jk.result].(string)
			if !ok {
				log.Printf("Result is not a string: %s", build[jk.result])
				return errors.New("unexpected jenkins value")
			}
			if result == jk.resultFailure {
				failedBuilds.WithLabelValues(j.jobName, j.repoName).Observe(t)
				log.Printf("%s, %s, %d build failed", j.repoName, j.jobName, i)
			} else if result == jk.resultSUCCESS {
				max = math.Max(max, t)
				min = math.Min(min, t)
				succeededBuilds.WithLabelValues(j.jobName, j.repoName).Observe(t)
				if codeCovRegister[j.repoName] && j.jobName == codeCovTrackJob {
					object := fmt.Sprintf("%s/%s/%d", j.repoName, j.jobName, i)
					coverage, err := getCoverage(object)
					if err != nil {
						log.Printf("Failed to get coverage, target: %s", object)
					} else {
						for p, c := range coverage {
							codeCoverage.WithLabelValues(p, j.repoName).Set(c)
						}
					}
				}
				log.Printf("%s, %s, %d build succeeded", j.repoName, j.jobName, i)
			}
		}
	}

	if max > 0 {
		succeededBuildMax.WithLabelValues(j.jobName, j.repoName).Set(max)
	}
	if min < math.MaxFloat64 {
		succeededBuildMin.WithLabelValues(j.jobName, j.repoName).Set(min)
	}
	j.lastBuildID = latestCompletedBuild

	return nil
}

func (j *job) getLatestCompletedBuild() (int, error) {
	o, err := j.getJenkinsObject("lastCompletedBuild")
	if err != nil {
		return 0, fmt.Errorf("failed to get last build info of build, %v", err)
	}
	if len(o) == 0 {
		return 0, errors.New("get empty result from jenkins")
	}
	id, ok := o[jk.number].(float64)
	if !ok {
		return -1, errors.New(`"number" is not a valid value`)
	}
	return int(id), nil

}

func (j *job) getJenkinsObject(build string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/job/%s/job/%s/%s/api/json", *jenkinsURL, j.repoName, j.jobName, build)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Print("Failed to create new request.")
		return nil, err
	}

	request.SetBasicAuth(os.Getenv("JENKINS_USERNAME"), os.Getenv("JENKINS_TOKEN"))
	resp, err := j.client.Do(request)
	if err != nil {
		log.Print("Failed to accomplish the request.")
		return nil, err
	}
	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Print("Failed to read from http response.")
		return nil, err
	}
	s := string(bodyText)

	var f interface{}
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		log.Print("Failed to unmarshal response to json.")
		return nil, err
	}

	o, ok := f.(map[string]interface{})
	if !ok {
		log.Print(f)
		return nil, errors.New("failed to parse JenkinsObject")
	}
	return o, nil
}

func getCoverage(object string) (map[string]float64, error) {
	cov := make(map[string]float64)
	ctx := context.Background()
	r, err := gcsClient.Bucket(*gcsBucket).Object(object).NewReader(ctx)
	if err != nil {
		log.Printf("Failed to download coverage file from gcs, %v", err)
		return nil, err
	}

	defer func() {
		if err = r.Close(); err != nil {
			log.Printf("Failed to close gcs file reader, %v", err)
		}
	}()

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if parts := strings.Split(scanner.Text(), "\t"); len(parts) == 2 {
			if n, err := strconv.ParseFloat(parts[1], 64); err != nil {
				log.Printf("Failed to parse codecov file: %s, %v", scanner.Text(), err)
			} else {
				cov[parts[0]] = n
			}
		} else {
			log.Printf("Failed to parse codcov file: %s, broken line", scanner.Text())
		}
	}

	return cov, scanner.Err()
}
