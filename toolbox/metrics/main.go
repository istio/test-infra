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
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"cloud.google.com/go/storage"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/net/context"
)

const (
	fetchInterval   = 5
	jenkinsUsername = "JENKINS_USERNAME"
	jenkinsToken    = "JENKINS_TOKEN"
)

var (
	webAddress      = flag.String("listen_port", ":9103", "Port on which to expose metrics and web interface.")
	jenkinsURL      = flag.String("jenkins_url", "https://testing.istio.io", "Jenkins URL.")
	gcsBucket       = flag.String("bucket", "istio-code-coverage", "GCS bucket name.")
	codeCovTrackJob = flag.String("coverage_job", "fast-forward", "In which job we are tracking code coverage.")

	metricsSuite *p8sMetricsSuite

	gcsClient  *storage.Client
	httpClient = &http.Client{}

	jk = &jkConst{
		name:               "name",
		jobs:               "jobs",
		building:           "building",
		result:             "result",
		duration:           "duration",
		number:             "number",
		resultFailure:      "FAILURE",
		resultSuccess:      "SUCCESS",
		apiJSON:            "api/json",
		lastCompletedBuild: "lastCompletedBuild",
	}
)

type jkConst struct {
	name               string
	jobs               string
	building           string
	result             string
	duration           string
	number             string
	resultSuccess      string
	resultFailure      string
	apiJSON            string
	lastCompletedBuild string
}

type p8sMetricsSuite struct {
	succeededBuilds   *prometheus.SummaryVec
	failedBuilds      *prometheus.SummaryVec
	succeededBuildMax *prometheus.GaugeVec
	succeededBuildMin *prometheus.GaugeVec
	codeCoverage      *prometheus.GaugeVec
}

type repo struct {
	name string
	jobs map[string]*job
}

type job struct {
	repoName    string
	jobName     string
	lastBuildID int
}

func newJob(repoName, jobName string) *job {
	return &job{
		repoName:    repoName,
		jobName:     jobName,
		lastBuildID: -1,
	}
}

func newRepo(name string) *repo {
	return &repo{
		name: name,
		jobs: make(map[string]*job),
	}
}

func newP8sMetricsSuite() *p8sMetricsSuite {
	return &p8sMetricsSuite{
		succeededBuilds: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name: "succeeded_build_durations_seconds",
				Help: "Succeeded build durations seconds.",
			},
			[]string{"build_job", "repo"},
		),

		failedBuilds: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name: "failed_build_durations_seconds",
				Help: "Failed build durations seconds.",
			},
			[]string{"build_job", "repo"},
		),

		succeededBuildMax: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "succeeded_build_durations_max_seconds",
				Help: "Succeeded max build durations seconds.",
			},
			[]string{"build_job", "repo"},
		),

		succeededBuildMin: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "succeeded_build_durations_min_seconds",
				Help: "Succeeded min build durations seconds.",
			},
			[]string{"build_job", "repo"},
		),

		codeCoverage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "package_code_coverage",
				Help: "Package code coverage.",
			},
			[]string{"package", "repo"},
		),
	}
}

func init() {
	metricsSuite = newP8sMetricsSuite()
	metricsSuite.registerMetricVec()
}

func main() {
	flag.Parse()

	var err error
	gcsClient, err = storage.NewClient(context.Background())
	if err != nil {
		log.Printf("Failed to create a gcs client, %v", err)
	}

	repoRegister := map[string]*repo{}
	go func() {
		for {
			updateRepoList(repoRegister)
			for _, repo := range repoRegister {
				repo.updateJobList()
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

func (m *p8sMetricsSuite) registerMetricVec() {
	prometheus.MustRegister(m.succeededBuilds)
	prometheus.MustRegister(m.succeededBuildMax)
	prometheus.MustRegister(m.succeededBuildMin)
	prometheus.MustRegister(m.failedBuilds)
	prometheus.MustRegister(m.codeCoverage)
}

func updateRepoList(repoRegister map[string]*repo) {
	l, err := listJenkinsItems("")
	if err != nil {
		log.Printf("Failed to update repo list: %s", err)
		return
	}

	for _, repoName := range l {
		if _, existed := repoRegister[repoName]; !existed {
			repoRegister[repoName] = newRepo(repoName)
		}
	}
}

func (r *repo) updateJobList() {
	l, err := listJenkinsItems(fmt.Sprintf("job/%s", r.name))
	if err != nil {
		log.Printf("Failed to update job list for repo %s: %s", r.name, err)
		return
	}

	for _, jobName := range l {
		if _, existed := r.jobs[jobName]; !existed {
			r.jobs[jobName] = newJob(r.name, jobName)
		}
	}
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
		object := fmt.Sprintf("%s/%s/%d", j.repoName, j.jobName, i)
		build, err := getJenkinsObject(fmt.Sprintf("job/%s/job/%s/%d", j.repoName, j.jobName, i))
		if err != nil {
			log.Printf("Failed to get build %s, %v", object, err)
			continue
		}

		b, ok := build[jk.building].(bool)
		if !ok {
			log.Printf("Unexpected jenkins value, %s. Building is not a bool: %s", object, build[jk.building])
			continue
		}

		if b {
			latestCompletedBuild = i - 1 //If this build is still building, set "latestCompletedBuild" to one build before
		} else {
			duration, ok := build[jk.duration].(float64)
			if !ok {
				log.Printf("Unexpected jenkins value, %s. Duration is not a float64: %s", object, build[jk.duration])
				continue
			}
			t := duration / 1000
			result, ok := build[jk.result].(string)
			if !ok {
				log.Printf("Unexpected jenkins value, %s. Result is not a string: %s", object, build[jk.result])
				continue
			}
			if result == jk.resultFailure {
				metricsSuite.failedBuilds.WithLabelValues(j.jobName, j.repoName).Observe(t)
				log.Printf("%s build failed", object)
			} else if result == jk.resultSuccess {
				max = math.Max(max, t)
				min = math.Min(min, t)
				metricsSuite.succeededBuilds.WithLabelValues(j.jobName, j.repoName).Observe(t)
				if j.jobName == *codeCovTrackJob {
					coverage, err := getCoverage(object)
					if err != nil {
						log.Printf("Failed to get coverage, target: %s", object)
					} else {
						for p, c := range coverage {
							metricsSuite.codeCoverage.WithLabelValues(p, j.repoName).Set(c)
						}
					}
				}
				log.Printf("%s build succeeded", object)

			}
		}
	}

	if max > 0 {
		metricsSuite.succeededBuildMax.WithLabelValues(j.jobName, j.repoName).Set(max)
	}
	if min < math.MaxFloat64 {
		metricsSuite.succeededBuildMin.WithLabelValues(j.jobName, j.repoName).Set(min)
	}
	j.lastBuildID = latestCompletedBuild

	return nil
}

func (j *job) getLatestCompletedBuild() (int, error) {
	o, err := getJenkinsObject(fmt.Sprintf("job/%s/job/%s/%s", j.repoName, j.jobName, jk.lastCompletedBuild))
	if err != nil {
		return 0, fmt.Errorf("failed to get last build info of build, %v", err)
	}
	if len(o) == 0 {
		return 0, errors.New("got empty result from jenkins")
	}
	id, ok := o[jk.number].(float64)
	if !ok {
		return -1, errors.New(`"number" is not a valid value`)
	}
	return int(id), nil

}

func getJenkinsObject(object string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/%s", *jenkinsURL, filepath.Join(object, jk.apiJSON))
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Print("Failed to create new request.")
		return nil, err
	}

	request.SetBasicAuth(os.Getenv(jenkinsUsername), os.Getenv(jenkinsToken))
	resp, err := httpClient.Do(request)
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

func listJenkinsItems(object string) ([]string, error) {
	o, err := getJenkinsObject(object)
	if err != nil {
		return nil, err
	}
	var list []string
	l, ok := o[jk.jobs].([]interface{})
	if !ok {
		log.Printf("can't parse from jenkins jobs to list of interface: %s", o[jk.jobs])
		return list, errors.New("unexpected jenkins value")
	}

	for _, i := range l {
		j, ok := i.(map[string]interface{})
		if !ok {
			log.Printf("can't parse to map: %s", j)
		} else {
			jobName, ok := j[jk.name].(string)
			if !ok {
				log.Printf("can't parse to string: %s", j[jk.name])
			}
			list = append(list, jobName)
		}
	}

	return list, nil
}

func getCoverage(object string) (map[string]float64, error) {
	cov := make(map[string]float64)
	ctx := context.Background()
	r, err := gcsClient.Bucket(*gcsBucket).Object(object).NewReader(ctx)
	if err != nil {
		if err == storage.ErrBucketNotExist || err == storage.ErrObjectNotExist {
			return cov, nil
		}
		log.Printf("Failed to download coverage file from gcs, %v", err)
		return nil, err
	}

	defer func() {
		if err = r.Close(); err != nil {
			log.Printf("Failed to close gcs file reader, %v", err)
		}
	}()

	//Line example: "istio.io/mixer/adapter/denyChecker	99"
	scanner := bufio.NewScanner(r)
	reg := regexp.MustCompile(`(.*)\t(.*)`)
	for scanner.Scan() {
		if m := reg.FindStringSubmatch(scanner.Text()); len(m) == 3 {
			if n, err := strconv.ParseFloat(m[2], 64); err != nil {
				log.Printf("Failed to parse codecov file: %s, %v", scanner.Text(), err)
			} else {
				cov[m[1]] = n
			}
		} else {
			log.Printf("Failed to parse codecov file: %s, broken line", scanner.Text())
		}
	}

	return cov, scanner.Err()
}
