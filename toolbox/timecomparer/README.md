# Find And Compare Time Commands

## Compare Time Commands From Build-log.txt

The file reads dirctory paths to prs that contain build-log.txt in GCS from Google Spreadsheet. It processes each build-log.txt file from the directories and finds places where `time (*)` commands are called. It then checks the output for time command to get the amount of time the commands run for. It finally generates an output that sort the time commands based on the average run time they take.

To run the file, you will need a `credentials.json` file from <https://developers.google.com/sheets/api/quickstart/go> when you click `ENABLE THE GOOGLE SHEETS API` and `DOWNLOAD CLIENT CONFIGURATION` to download the `credentials.json` file and use it as path to flag -CredentialsPath.

You will also need a google spreadsheet id that contains the path to pr folder in a bucket in Google Cloud Storage. For a weblink `https://docs.google.com/spreadsheets/d/1fUmFA2CcnSQuCwEilAewftwitvSOLZGcj2BQiur37rI/edit#gid=935149480`, the spreadsheet id to pass to flag SpreadsheetID is `1fUmFA2CcnSQuCwEilAewftwitvSOLZGcj2BQiur37rI`.

```bash
$ git clone https://github.com/istio/test-infra.git
```

and build it using

```bash
$ go build //toolbox/timecomparer/main.go
```

or

```bash
$ bazel build //toolbox/timecomparer/timecomparer
```

The binary output is located in //toolbox/timecomparer.

Run code using

```bash
$ go run //toolbox/timecomparer/main.go -OutputFileName=<output.csv> -SpreadsheetID=<spreadsheet_id> -CredentialsPath=<credentials.json> -ReadRange=<read_range> -BucketName=<bucket_name> -SplitBy=5
```

and read the result in output file in the directory and filename specified in flag OutputFileName.

A glance of the output csv file:

```bash
$ time (mkdir -p /home/prow/go/out/linux_amd64/release/docker_build/docker.pilot && cp -r pilot/docker/Dockerfile.pilot /home/prow/go/out/linux_amd64/release/pilot-discovery tests/testdata/certs/cacert.pem /home/prow/go/out/linux_amd64/release/docker_build/docker.pilot && cd /home/prow/go/out/linux_amd64/release/docker_build/docker.pilot &&  docker build  --build-arg BASE_DISTRIBUTION=default -t gcr.io/istio-testing/pilot:4f1d5aa1312789046bf81ccba1c7a9a0c8c743e8-e2e_pilotv2_auth_sds -f Dockerfile.pilot . );,2.033630555555556E+01,pr-logs/pull/istio_istio/14927/istio_auth_sds_e2e-master/2349/,real 0m15.107s,user 0m1.658s,sys 0m0.965s
```

In the output, the first column is the `time` command that is being run. The second column is the average time it takes on all build logs of pull requests which contains the command. The next columns are the pull request path in gcs and the time command output in the build-log.
