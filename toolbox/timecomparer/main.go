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
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/golang-collections/collections/stack"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type TimeComparer struct {
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

func NewTimeComparer(client *storage.Client, bucketName string) *TimeComparer {
	return &TimeComparer{
		client:     client,
		bucketName: bucketName,
	}
}

// Read from gcs the build-log.txt files into slice of lines.
func (tc *TimeComparer) query(ctx context.Context, prefix string) ([]string, error) {
	client := tc.client
	bucket := client.Bucket(tc.bucketName)
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

type runPair struct {
	name       string
	stringTime string
	time       []float64
}

type runTime struct {
	runTime []runPair
	runPath string
}

func (r runTime) toStringSlice() []string {
	rSlice := []string{r.runPath}
	for _, rTime := range r.runTime {
		rSlice = append(rSlice, rTime.name, rTime.stringTime)
	}
	return rSlice
}

type runCombination struct {
	totalTime float64
	runTimes  []runTime
}

type commandAndTime struct {
	command    string
	runStrings []string
}

// Write error map and warning map to csv file with given file path.
func writeCSV(sortedArray []float64, readResult map[string][]string, timeMap map[float64]commandAndTime, fileName string) {
	var file *os.File
	var err error
	// If file already exists in the path, write to original file.
	if _, err = os.Stat(fileName); err == nil {
		file, err = os.OpenFile(fileName, os.O_RDWR, 0644)
		if err != nil {
			log.Fatal("cannot open file", err)
		}
	} else {
		file, err = os.Create(fileName)
		if err != nil {
			log.Fatal("cannot open file", err)
		}
	}

	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, sortedTime := range sortedArray {
		comAndTime := timeMap[sortedTime]
		emptyCom := commandAndTime{}
		if reflect.DeepEqual(comAndTime, emptyCom) {
			continue
		}

		com := comAndTime.command
		newTimeResult := comAndTime.runStrings
		originalResult := readResult[com]

		completeResult := append(originalResult, newTimeResult...)

		averageTime := strconv.FormatFloat(sortedTime, 'E', -1, 64)

		toWrite := []string{}
		toWrite = append(toWrite, com, averageTime)
		toWrite = append(toWrite, completeResult...)
		_ = writer.Write(toWrite)
	}
}

// Read from original csv file the time and commands that are already present.
// Have one return value hold the time and command in float/string format to combine with files to be read later.
// Have another value holding the file content that already presents in the original csv to be appended to new values.
func readCSV(fileName string) (map[string][]string, map[string]runCombination, error) {
	// Read actual contents from file and store them.
	// First map key is the same float time that would appear in the result of actual processing.
	// Second map key is command. The map value is the string slice that are in each csv file line.

	commandMap := map[string][]string{}
	commandToTime := map[string]runCombination{}
	csvFile, err := os.Open(fileName)
	if err != nil {
		return commandMap, commandToTime, err
	}
	scanner := bufio.NewScanner(csvFile)
	for scanner.Scan() {
		completeLine := scanner.Text()
		sections := strings.Split(completeLine, ",")
		if len(sections) < 2 {
			continue
		}
		command := sections[0]
		averageTime := sections[1]

		averageTimeFloat, err := strconv.ParseFloat(averageTime, 64)
		if err != nil {
			continue
		}
		numPaths := (len(sections) - 2) / 3

		undefinedRuntimes := []runTime{}

		for i := 0; i < numPaths; i++ {
			undefinedRuntimes = append(undefinedRuntimes, runTime{})
		}
		rCombination := runCombination{
			totalTime: averageTimeFloat * float64(numPaths),
			runTimes:  undefinedRuntimes,
		}
		commandToTime[command] = rCombination

		commandMap[command] = sections[2:]
	}

	return commandMap, commandToTime, nil
}

func processCommand(commandToTime map[string]runCombination) ([]float64, map[float64]commandAndTime) {
	timeArray := []float64{}
	commandsBreakDown := map[float64]commandAndTime{}
	for command, rCombination := range commandToTime {
		time := rCombination.totalTime
		rTimes := rCombination.runTimes
		averageTime := time / float64(len(rTimes))
		cAndTime := commandAndTime{
			command:    command,
			runStrings: []string{},
		}
		for _, r := range rTimes {
			cAndTime.runStrings = append(cAndTime.runStrings, r.toStringSlice()...)
		}
		timeArray = append(timeArray, averageTime)
		commandsBreakDown[averageTime] = cAndTime
	}
	return timeArray, commandsBreakDown
}

// Convert minutes and sesconds ("0m1.061s") to float object.
func convertStringToTime(timeString string) []float64 {
	timeSlice := []float64{}
	if strings.Contains(timeString, "m") {
		timeSplit := strings.Split(timeString, "m")
		minString := timeSplit[0]
		fmt.Println("get minutes")
		fmt.Println(minString)
		minInt, err := strconv.Atoi(minString)
		if err != nil {
			fmt.Println("error parsing float minute")
			timeSlice = append(timeSlice, -1.0)
		} else {
			minFloat := float64(minInt)
			timeSlice = append(timeSlice, minFloat)
		}
		if len(timeSplit) > 1 {
			secString := timeSplit[len(timeSplit)-1]
			secValueString := strings.Split(secString, "s")[0]
			fmt.Println("get seconds")
			fmt.Println(secValueString)
			secFloat, err := strconv.ParseFloat(secValueString, 64)
			if err != nil {
				fmt.Println("error parsing float second")
				timeSlice = append(timeSlice, -1.0)
			} else {
				timeSlice = append(timeSlice, secFloat)
			}
		}
	}
	return timeSlice
}

// Extract time from runtime string with the format of "real	0m1.061s"
func extractTime(runTimeString string) runPair {
	timeSplit := strings.Fields(runTimeString)
	length := len(timeSplit)
	var ind int
	for ind = length - 1; ind > 0; ind-- {
		if strings.Compare(timeSplit[ind], "") != 0 {
			break
		}
	}
	timePart := timeSplit[ind]
	timeSlice := convertStringToTime(timePart)

	for ind = 0; ind < length; ind++ {
		if strings.Compare(timeSplit[ind], "") != 0 {
			break
		}
	}
	namePart := timeSplit[ind]

	rPair := runPair{
		name:       namePart,
		time:       timeSlice,
		stringTime: timePart,
	}

	return rPair

}

// Compute average user/real/sys time to compare with others.
func getAverageTime(realTime, userTime, sysTime runPair) float64 {
	averageSlice := []float64{}
	realT := realTime.time
	userT := userTime.time
	sysT := sysTime.time

	for i := 0; i < 2; i++ {
		averagePart := (realT[i] + userT[i] + sysT[i]) / 3
		averageSlice = append(averageSlice, averagePart)
	}
	average := averageSlice[0]*60 + averageSlice[1]

	return average
}

// Get error contents of files to update command-time map with commands and path to build logs.
func (tc *TimeComparer) findTimeCommands(
	ctx context.Context, filePaths []string, commandToTime map[string]runCombination) map[string]runCombination {
	timeCommand := regexp.MustCompile(`^time (.*);`)
	realOutput := regexp.MustCompile(`^real(\s)*([0-9])*m([0-9])*\.([0-9])*s`)
	userOutput := regexp.MustCompile(`^user(\s)*([0-9])*m([0-9])*\.([0-9])*s`)
	sysOutput := regexp.MustCompile(`^sys(\s)*([0-9])*m([0-9])*\.([0-9])*s`)
	for _, filePath := range filePaths {
		fileSlice, err := tc.query(ctx, filePath)
		if err != nil {
			continue
		}

		fileStack := stack.New()
		for start := 0; start < len(fileSlice); start++ {
			line := fileSlice[start]
			if timeCommand.MatchString(line) {
				fileStack.Push(line)
			}

			if realOutput.MatchString(fileSlice[start]) && userOutput.MatchString(fileSlice[start+1]) && sysOutput.MatchString(fileSlice[start+2]) {
				if fileStack.Len() != 0 {
					corespondingCommand := fileStack.Pop()
					stringCommand := corespondingCommand.(string)
					// Find each section of time in the output
					realTime := extractTime(fileSlice[start])
					userTime := extractTime(fileSlice[start+1])
					sysTime := extractTime(fileSlice[start+2])
					average := getAverageTime(realTime, userTime, sysTime)

					rTime := runTime{
						runTime: []runPair{
							realTime, userTime, sysTime,
						},
						runPath: filePath,
					}

					empty := runCombination{}
					if reflect.DeepEqual(commandToTime[stringCommand], empty) {
						commandToTime[stringCommand] = runCombination{
							totalTime: average,
							runTimes:  []runTime{rTime},
						}
					} else {
						rCombination := commandToTime[stringCommand]
						rCombination.totalTime += average
						rCombination.runTimes = append(rCombination.runTimes, rTime)
						commandToTime[stringCommand] = rCombination
					}
				}
			}
		}
	}
	return commandToTime
}

// For each spliter section of file paths, read previously generated errors and warnings.
// Process contents in new files and update the csv.
func (tc *TimeComparer) findErrorForEachSection(ctx context.Context, gcsFilePaths []string, outputFileName string, startInd int, endInd int) {
	readResult, commandToTime, err := readCSV(outputFileName)

	// If csv file does not contain anything yet, initialize curErrors and curWarnings to be empty maps.
	if err != nil {
		commandToTime = map[string]runCombination{}
		readResult = map[string][]string{}
	}
	filesToProcess := append([]string{}, gcsFilePaths[startInd:endInd]...)

	// Get time/command and pull request paths from build log.
	commandToTime = tc.findTimeCommands(ctx, filesToProcess, commandToTime)

	sortedArray, timeMap := processCommand(commandToTime)

	// Sort the time values
	sort.Float64s(sortedArray)
	// Write combined content to output csv file.
	writeCSV(sortedArray, readResult, timeMap, outputFileName)
}

// Divide the list of file paths read in gcs storage to several spliters to avoid overflooding memory.
func (tc *TimeComparer) divideToSections(ctx context.Context, spliter int, gcsFilePaths []string, outputFileName string) {
	n := 1
	for {
		if n*spliter > len(gcsFilePaths) {
			break
		}
		startInd := (n - 1) * spliter
		endInd := n*spliter - 1
		tc.findErrorForEachSection(ctx, gcsFilePaths, outputFileName, startInd, endInd)
		n++
	}
	startInd := (n - 1) * spliter
	endInd := len(gcsFilePaths)
	tc.findErrorForEachSection(ctx, gcsFilePaths, outputFileName, startInd, endInd)

}

// Optional on whether or not to copy and share output file to gcs.
func (tc *TimeComparer) CopyToGCS(ctx context.Context, source, bucketName, fileName string, public bool) {
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
	var credentialsPath string
	flag.StringVar(&credentialsPath, "CredentialsPath", "credentials.json", "Path to Credentials.json file.")
	var readRange string
	flag.StringVar(&readRange, "ReadRange", "A1:A5", "ReadRange from the spread sheet such as Sheet1!A1:A reads all elements in column A sheet 1.")
	var bucketName string
	flag.StringVar(&bucketName, "BucketName", "istio-prow", "Bucket Name to read from in GCS.")

	var public bool
	flag.BoolVar(&public, "Public", true, "Whether or not the file should be public in GCS")

	var outputBucketName string
	flag.StringVar(&outputBucketName, "OutputBucketName", "istio-prow", "Bucket Name to write file to in GCS.")

	var apiKey string
	flag.StringVar(&apiKey, "APIKey", "", "API key to create new google sheets service")

	var spliter int
	flag.IntVar(&spliter, "SplitBy", 2, "The number of run paths to process for each read/write section.")

	flag.Parse()

	if strings.Compare(fileName, "") == 0 {
		log.Fatal("Please enter an output file name with .csv extension.")
	}

	if strings.Compare(spreadsheetID, "") == 0 {
		log.Fatal("Please enter a spreadsheet id with paths to the folder containing build-log.txt.")
	}

	if strings.Compare(credentialsPath, "") == 0 {
		log.Fatal("Please enter the path to credentials file.")
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

	if strings.Compare(apiKey, "") == 0 {
		log.Fatal("Please enter a valid API Key.")
	}

	context := con.Background()
	listOfPrs := readSpreadSheet(context, apiKey, spreadsheetID, readRange)

	client, err := storage.NewClient(context)
	if err != nil {
		log.Fatalf("Unable to create new client: %v", err)
	}

	timeComparer := NewTimeComparer(client, bucketName)
	timeComparer.divideToSections(context, spliter, listOfPrs, fileName)
}
