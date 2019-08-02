// Copyright 2018 Istio Authors
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
	con "context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func contains(s []string, e string) bool {
	for _, a := range s {
		if strings.Compare(a, e) == 0 {
			return true
		}
	}
	return false
}

type ErrorFinder struct {
	client     *storage.Client
	bucketName string
}

// Read google spreadsheets with credentials, spreadsheet ID and read range to get slice of file paths to build-logs.
func readSpreadSheet(ctx con.Context, apiKey string, spreadsheetID string, readRange string) []string {
	srv, err := sheets.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	if len(resp.Values) == 0 {
		fmt.Println("No data found.")
		return nil
	}
	filePaths := []string{}

	for _, row := range resp.Values {
		filePath := row[0]
		filePathString, err := filePath.(string)
		if err {
			filePaths = append(filePaths, filePathString)
		}
	}

	return filePaths
}

func NewErrorFinder(client *storage.Client, bucketName string) *ErrorFinder {
	return &ErrorFinder{
		client:     client,
		bucketName: bucketName,
	}
}

// Read from gcs the build-log.txt files into slice of lines.
func (f *ErrorFinder) query(ctx context.Context, prefix string) ([]string, error) {
	client := f.client
	bucket := client.Bucket(f.bucketName)
	buildFile := bucket.Object(prefix + "build-log.txt")

	rc, err := buildFile.NewReader(ctx)
	if err != nil {
		return []string{}, err
	}
	defer rc.Close()
	lines := []string{}

	scanner := bufio.NewScanner(rc)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) != 0 {
			lines = append(lines, line)
		}
	}

	return lines, nil
}

// Eliminate unnecessary duplicate '\\s' characters in word slice.
func unduplicate(inputSlice []string) []string {
	output := []string{}
	outputInd := 0
	for ind, element := range inputSlice {
		if ind == 0 {
			output = append(output, element)
		} else if strings.Compare(output[outputInd], "\\s") != 0 {
			outputInd++
			output = append(output, element)
			continue
		}
	}
	return output
}

// Get error contents of files to update error map and warning map with new errors/warnings and path to build logs.
func (f *ErrorFinder) getContent(
	ctx context.Context, filePaths []string, errorMap map[string][]string, warningMap map[string][]string) (map[string][]string, map[string][]string) {
	for _, filePath := range filePaths {
		fileSlice, err := f.query(ctx, filePath)
		if err != nil {
			continue
		}
		output := []string{}
		for _, line := range fileSlice {

			// Ignore command lines to run that starts with "+".
			if strings.Contains(line, "+") {
				continue
			}
			// Generalize digits and content within quotations in current line.
			generalized := f.generalizeDigits(line)
			output = append(output, generalized)
		}

		// Loop through only generalized result lines.
		for _, line := range output {
			toLower := strings.ToLower(line)

			// Only look at lines with key words of "error", "warning" and "fail".
			if strings.Contains(toLower, "error") || strings.Contains(toLower, "failed") || strings.Contains(toLower, "failure") || strings.Contains(toLower, "warn") {

				// Delete date information that are possibly present at the beginning of each line.
				line = strings.TrimLeft(line, "\\d-T:.Z	info")
				lineSplit := strings.Split(line, " ")
				newLineSplit := []string{}

				for indexD, section := range lineSplit {

					// Ignore empty strings
					if strings.Compare(section, "") == 0 {
						continue
					}

					// If word does not contain key words, generalize word to characters necessary.
					if !strings.Contains(strings.ToLower(section), "error") &&
						!strings.Contains(strings.ToLower(section), "fail") && !strings.Contains(strings.ToLower(section), "warn") {
						if (indexD > 0 && strings.Compare(lineSplit[indexD-1], "container") == 0) ||
							strings.Contains(section, "\\d") || strings.Contains(section, "/") {
							newLineSplit = append(newLineSplit, "\\s")
						} else {
							newLineSplit = append(newLineSplit, section)
						}
					} else {
						// If word contains keyword, append it to list.
						newLineSplit = append(newLineSplit, section)
					}
				}
				newLineSplit = unduplicate(newLineSplit)
				// Get general error message not the detailed eplanation that usually comes after colon.
				line = strings.Join(newLineSplit, " ")
				if strings.Contains(line, ":") {
					indexColon := strings.Index(line, ":")
					if indexColon < 20 {
						newLine := strings.Replace(line, ":", "<", 1)
						if strings.Contains(newLine, ":") {
							indColon := strings.Index(newLine, ":")
							if strings.Contains(line[:indColon], "error") || strings.Contains(line[:indColon], "warn") || strings.Contains(line[:indColon], "fail") {
								line = line[:indColon]
							}

						}
					} else if strings.Contains(line[:indexColon], "error") || strings.Contains(line[:indexColon], "warn") || strings.Contains(line[:indexColon], "fail") {
						line = line[:indexColon]
					}
				}

				// If error line contains keyword "warn", it is a warning and should be placed in warning map.
				if strings.Contains(strings.ToLower(line), "warn") {
					var warningPrs []string
					if warningMap[line] == nil {
						warningPrs = []string{filePath}
					} else {
						warningPrs = warningMap[line]
						if !contains(warningPrs, filePath) {
							warningPrs = append(warningPrs, filePath)
						}
					}
					warningMap[line] = warningPrs
					// Else it's an error and should be placed to error map.
				} else {
					var prSlice []string
					if errorMap[line] == nil {
						prSlice = []string{filePath}
					} else {
						prSlice = errorMap[line]
						if !contains(prSlice, filePath) {
							prSlice = append(prSlice, filePath)
						}
					}
					errorMap[line] = prSlice
				}
			}
		}
	}
	return errorMap, warningMap
}

// Read destination csv file to get previously generated errors and warnings.
func readCSV(fileName string) (map[string][]string, map[string][]string, error) {
	existingError := map[string][]string{}
	warningMap := map[string][]string{}
	csvFile, err := os.Open(fileName)
	if err != nil {
		return existingError, warningMap, err
	}

	var isError bool

	// Read csv file with bufio and scanner line by line.
	scanner := bufio.NewScanner(csvFile)
	for scanner.Scan() {
		line := strings.Split(scanner.Text(), ",")

		// If header "error" is found, the following lines are all errors that should
		// be stored in map existingError.
		if strings.Compare(line[0], "error") == 0 {
			isError = true
			continue
		}

		if isError {

			// If header "warning" is found, the follwing lines after "warning" are
			// all warnings that should be stored in map warningMap.
			if strings.Compare(line[0], "warning") == 0 {
				isError = false
				continue
			} else {
				errorMessage := line[0]
				ind := 2
				for {
					ele := line[ind]
					if strings.Contains(ele, "/") {
						break
					}
					ind++
				}
				errorAppearance := line[ind:]
				existingError[errorMessage] = errorAppearance
			}
			// If the lines are warnings, update the warning map.
		} else {
			warningMessage := line[0]
			ind := 2
			for {
				ele := line[ind]
				if strings.Contains(ele, "/") {
					break
				}
				ind++
			}

			warningAppearance := line[ind:]
			warningMap[warningMessage] = warningAppearance
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return existingError, warningMap, nil
}

// For each spliter section of file paths, read previously generated errors and warnings.
// Process contents in new files and update the csv.
func (f *ErrorFinder) findErrorForEachSection(ctx context.Context, gcsFilePaths []string, outputFileName string, startInd int, endInd int) {
	curErrors, curWarnings, err := readCSV(outputFileName)

	// If csv file does not contain anything yet, initialize curErrors and curWarnings to be empty maps.
	if err != nil {
		curErrors = map[string][]string{}
		curWarnings = map[string][]string{}
	}
	filesToProcess := append([]string{}, gcsFilePaths[startInd:endInd]...)

	// Get errors and warnings from newly read build logs and update maps of errors and warnings.
	newErrors, newWarnings := f.getContent(ctx, filesToProcess, curErrors, curWarnings)

	// Write new errors and warnings to output csv file.
	writeCSV(newErrors, newWarnings, outputFileName)
}

// Divide the list of file paths read in gcs storage to several spliters to avoid overflooding memory.
func (f *ErrorFinder) divideToSections(ctx context.Context, spliter int, gcsFilePaths []string, outputFileName string) {
	n := 1
	for {
		if n*spliter > len(gcsFilePaths) {
			break
		}
		startInd := (n - 1) * spliter
		endInd := n*spliter - 1
		f.findErrorForEachSection(ctx, gcsFilePaths, outputFileName, startInd, endInd)
		n++
	}
	startInd := (n - 1) * spliter
	endInd := len(gcsFilePaths)
	f.findErrorForEachSection(ctx, gcsFilePaths, outputFileName, startInd, endInd)

}

// Generalize messages from build file.
func (f *ErrorFinder) generalizeDigits(content string) string {
	// Generalize unique numbers in build-log.txt to character '\d'
	reg := regexp.MustCompile("[0-9]+")
	newContent := reg.ReplaceAllString(content, "\\d")

	// Generalize unique messages in build-log.txt surrounded by quotations to character '\d'
	reg = regexp.MustCompile("\\\"(.*?)\\\"|'(.*?)'")
	newContent = reg.ReplaceAllString(newContent, "\\s")

	return newContent
}

// Write error map and warning map to csv file with given file path.
func writeCSV(errorMap map[string][]string, warningMap map[string][]string, fileName string) {
	var file *os.File
	var err error
	// If file already exists in the path, write to original file.
	if _, err := os.Stat(fileName); err == nil {
		file, err = os.OpenFile(fileName, os.O_RDWR, 0644)
		if err != nil {
			log.Fatal("cannot open file", err)
		}
		// If file does not exist, create a new file to write to.
	} else {
		file, err = os.Create(fileName)
		if err != nil {
			log.Fatal("cannot open file", err)
		}
	}

	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()
	errorSlice := []string{"error"}
	err = writer.Write(errorSlice)
	if err != nil {
		log.Fatal("cannot write line", err)
	}

	// Write each line of csv file for error map with a error header.
	for key, value := range errorMap {
		newLine := []string{}
		newLine = append(newLine, key)
		s := strconv.Itoa(len(value))
		newLine = append(newLine, s)
		newLine = append(newLine, value...)
		err := writer.Write(newLine)
		if err != nil {
			log.Fatal("cannot write line", err)
		}

	}

	// Write each line of csv file for warning map with warning header.
	warning := []string{"warning"}
	err = writer.Write(warning)
	if err != nil {
		log.Fatal("cannot write line", err)
	}
	for key, value := range warningMap {
		newLine := []string{}
		newLine = append(newLine, key)
		s := strconv.Itoa(len(value))
		newLine = append(newLine, s)
		newLine = append(newLine, value...)
		err := writer.Write(newLine)
		if err != nil {
			log.Fatal("cannot write line", err)
		}

	}
}

// Optional on whether or not to copy and share output file to gcs.
func (f *ErrorFinder) CopyToGCS(ctx context.Context, source, bucketName, fileName string, public bool) {
	r, err := os.Open(source)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Unable to create new client: %v", err)
	}

	bh := client.Bucket(bucketName)
	// Next check if the bucket exists
	if _, err = bh.Attrs(ctx); err != nil {
		log.Fatalf("Bucket does not exist: %v", err)
	}

	obj := bh.Object(fileName)
	w := obj.NewWriter(ctx)
	if _, err := io.Copy(w, r); err != nil {
		log.Fatalf("Cannot copy file to GCS: %v", err)
	}
	if err := w.Close(); err != nil {
		log.Fatalf("Unable to close writer: %v", err)
	}

	if public {
		if err := obj.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
			log.Fatalf("Unable to set object to public: %v", err)
		}
	}

	_, err = obj.Attrs(ctx)
	if err != nil {
		switch err {
		case storage.ErrBucketNotExist:
			log.Fatal("Please create the bucket first e.g. with `gsutil mb`")
		default:
			log.Fatal(err)
		}
	}

}

func main() {
	var fileName string
	flag.StringVar(&fileName, "OutputFileName", "output.csv", "A file name for output csv file with .csv at the end.")
	var spreadsheetID string
	flag.StringVar(&spreadsheetID, "SpreadsheetID", "1XK4Nbq7nwqwXMC7waMsEqhe_wOf1J1rqwbECrop9fQI", "Spreadsheet ID for the spreadsheet containing file path")
	var apiKey string
	flag.StringVar(&apiKey, "APIKey", "", "API key to access google sheets.")
	var readRange string
	flag.StringVar(&readRange, "ReadRange", "A1:A5", "ReadRange from the spread sheet such as Sheet1!A1:A reads all elements in column A sheet 1.")
	var bucketName string
	flag.StringVar(&bucketName, "BucketName", "istio-prow", "Bucket Name to read from in GCS.")

	var public bool
	flag.BoolVar(&public, "Public", true, "Whether or not the file should be public in GCS")

	var outputBucketName string
	flag.StringVar(&outputBucketName, "OutputBucketName", "istio-prow", "Bucket Name to write file to in GCS.")

	flag.Parse()

	if strings.Compare(fileName, "") == 0 {
		log.Fatal("Please enter an output file name with .csv extension.")
	}

	if strings.Compare(spreadsheetID, "") == 0 {
		log.Fatal("Please enter a spreadsheet id with paths to the folder containing build-log.txt.")
	}

	if strings.Compare(apiKey, "") == 0 {
		log.Fatal("Please enter a valid API key.")
	}

	if strings.Compare(readRange, "") == 0 {
		log.Fatal("Please enter a valid read range.")
	}

	if strings.Compare(bucketName, "") == 0 {
		log.Fatal("Please enter a bucket name.")
	}

	if strings.Compare(outputBucketName, "") == 0 {
		log.Fatal("Please enter an output bucket name.")
	}

	context := con.Background()
	listOfPrs := readSpreadSheet(context, apiKey, spreadsheetID, readRange)
	client, err := storage.NewClient(context)
	if err != nil {
		log.Fatalf("Unable to create new client: %v", err)
	}

	errorFinder := NewErrorFinder(client, bucketName)
	errorFinder.divideToSections(context, 2, listOfPrs, fileName)
}
