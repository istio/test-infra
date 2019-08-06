# Find And Compare Time Commands

## Compare Time Commands From Build-log.txt

The file reads dirctory paths to prs that contain build-log.txt in GCS from Google Spreadsheet. It processes each build-log.txt file from the directories and finds places where `time (*)` commands are called. It then checks the output for time command to get the amount of time the commands run for. It finally generates an output that sort the time commands based on the average run time they take.

To run the file, you will need a `credentials.json` file from https://developers.google.com/sheets/api/quickstart/go when you click `ENABLE THE GOOGLE SHEETS API` and `DOWNLOAD CLIENT CONFIGURATION` to download the `credentials.json` file and use it as path to flag -CredentialsPath.

You will also need a google spreadsheet id that contains the path to pr folder in a bucket in Google Cloud Storage. For a weblink `https://docs.google.com/spreadsheets/d/1fUmFA2CcnSQuCwEilAewftwitvSOLZGcj2BQiur37rI/edit#gid=935149480`, the spreadsheet id to pass to flag SpreadsheetID is `1fUmFA2CcnSQuCwEilAewftwitvSOLZGcj2BQiur37rI`.

```
$ git clone https://github.com/istio/test-infra.git
```

and build it using

```
$ go build //toolbox/timecomparer/main.go
```

or 
```
$ bazel build //toolbox/timecomparer/timecomparer
```

The binary output is located in //toolbox/timecomparer.

Run code using

```
$ go run //toolbox/timecomparer/main.go -OutputFileName=<output.csv> -SpreadsheetID=<spreadsheet_id> -CredentialsPath=<credentials.json> -ReadRange=<read_range> -BucketName=<bucket_name> -SplitBy=5
```

and read the result in output file in the directory and filename specified in flag OutputFileName.