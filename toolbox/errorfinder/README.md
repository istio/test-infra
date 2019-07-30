# Find Build Errors

## Examine Errors From Build-log.txt

The file reads dirctory paths to prs that contain build-log.txt in GCS from Google Spreadsheet. It process each build-log.txt file from the directories and identify errors and warnings within the build files. The output contains all errors and warnings from the build files and the directories in which the errors or warnings exist.

To run the file, you will need a `credentials.json` file from https://developers.google.com/sheets/api/quickstart/go when you click `ENABLE THE GOOGLE SHEETS API` and `DOWNLOAD CLIENT CONFIGURATION` to download the `credentials.json` file and use it as path to flag -CredentialsPath.

You will also need a google spreadsheet id that contains the path to pr folder in a bucket in Google Cloud Storage. For a weblink `https://docs.google.com/spreadsheets/d/1fUmFA2CcnSQuCwEilAewftwitvSOLZGcj2BQiur37rI/edit#gid=935149480`, the spreadsheet id to pass to flag SpreadsheetID is `1fUmFA2CcnSQuCwEilAewftwitvSOLZGcj2BQiur37rI`.

```
$ git clone https://github.com/istio/test-infra.git
```

and build it using

```
$ go build //toolbox/errorfinder/main.go
```

The binary output is located in //toolbox/errorfinder.

Run code using

```
$ go run //toolbox/errorfinder/main.go -OutputFileName=<output.csv> -SpreadsheetID=<spreadsheet_id> -CredentialsPath=<credentials.json> -ReadRange=<read_range> -BucketName=<bucket_name>
```

and read the result in output file in the directory and filename specified in flag OutputFileName.

