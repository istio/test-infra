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
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
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

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	err = json.NewEncoder(f).Encode(token)
	if err != nil {
		log.Fatalf("Unable to encode token to file path: %v", err)
	}
}

func readSpreadSheet(credentialsPath string, spreadsheetID string, readRange string) []string {
	b, err := ioutil.ReadFile(credentialsPath)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := sheets.New(client)
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

func NewErrorFinder(client *storage.Client, bucketName string) (*ErrorFinder, error) {
	return &ErrorFinder{
		client:     client,
		bucketName: bucketName,
	}, nil
}

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

func (f *ErrorFinder) getContent(ctx context.Context, filePaths []string) (map[string]string, map[string][]string, map[string][]string, error) {
	errorMap := map[string][]string{}
	warningMap := map[string][]string{}
	contentMap := map[string]string{}
	for _, filePath := range filePaths {
		fileSlice, err := f.query(ctx, filePath)
		if err != nil {
			continue
		}

		output := []string{}
		for _, line := range fileSlice {
			if !strings.Contains(line, "+") {
				output = append(output, line)
			}
		}
		outputString := strings.Join(output, "\n")
		outputString = f.generalizeDigits(outputString)

		processedString := outputString
		errorString := ""
		for {
			ind := strings.Index(processedString, "exit status")

			if ind == -1 {
				break
			}
			splitSlice := strings.SplitAfterN(processedString, "exit status", 2)
			beforeExit := strings.Split(splitSlice[0], "\n")
			afterExit := strings.Split(splitSlice[1], "\n")

			var tenLinesBefore []string
			if len(beforeExit) > 20 {
				tenLinesBefore = beforeExit[len(beforeExit)-20:]

			} else {
				tenLinesBefore = beforeExit
			}

			var tenLinesAfter []string
			if len(afterExit) > 20 {
				tenLinesAfter = afterExit[:20]
			} else {
				tenLinesAfter = afterExit[:]
			}

			total := append(tenLinesBefore, tenLinesAfter...)
			errorLines := []string{}

			for _, line := range total {
				if strings.Contains(line, "error") || strings.Contains(line, "fail") {
					line = strings.TrimLeft(line, "\\d-T:.Z	info")
					lineSplit := strings.Split(line, " ")
					newLineSplit := []string{}

					for indexD, section := range lineSplit {
						if strings.Compare(section, "") != 0 && !strings.Contains(section, "error") && !strings.Contains(section, "fail") {
							if indexD > 0 && strings.Compare(lineSplit[indexD-1], "container") == 0 {
								newLineSplit = append(newLineSplit, "\\s")
							} else if strings.Contains(section, "\\d") {
								newLineSplit = append(newLineSplit, "\\s")
							} else if strings.Contains(section, "/") {
								newLineSplit = append(newLineSplit, "\\s")
							} else {
								newLineSplit = append(newLineSplit, section)
							}
						} else if strings.Compare(section, "") != 0 {
							newLineSplit = append(newLineSplit, section)
						}
					}

					line = strings.Join(newLineSplit, " ")
					if strings.Contains(line, ":") {
						indexColon := strings.Index(line, ":")
						if indexColon < 20 {
							newLine := strings.Replace(line, ":", "<", 1)
							if strings.Contains(newLine, ":") {
								indColon := strings.Index(newLine, ":")
								line = line[:indColon]
							}
						} else {
							line = line[:indexColon]
						}
					}
					errorLines = append(errorLines, line)

					if strings.Contains(line, "warn") {
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

			errorString = errorString + "\n" + strings.Join(errorLines, "\n")
			processedString = splitSlice[1]
		}
		contentMap[filePath] = errorString
	}
	return contentMap, errorMap, warningMap, nil
}

func (f *ErrorFinder) generalizeDigits(content string) string {
	reg, _ := regexp.Compile("[0-9]+")
	newContent := reg.ReplaceAllString(content, "\\d")
	reg, _ = regexp.Compile("\\\"(.*?)\\\"|'(.*?)'")
	newContent = reg.ReplaceAllString(newContent, "\\s")

	return newContent
}

func writeCSV(errorMap map[string][]string, warningMap map[string][]string, fileName string) {
	file, err := os.Create(fileName)
	if err != nil {
		log.Fatal("cannot open file", err)
	}

	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()
	error := []string{"error"}
	err = writer.Write(error)
	if err != nil {
		log.Fatal("cannot write line", err)
	}
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

func main() {
	var fileName string
	flag.StringVar(&fileName, "OutputFileName", "", "A file name for output csv file with .csv at the end.")
	var spreadsheetID string
	flag.StringVar(&spreadsheetID, "SpreadsheetID", "", "Spreadsheet ID for the spreadsheet containing file path")
	var credentialsPath string
	flag.StringVar(&credentialsPath, "CredentialsPath", "", "Path to Credentials.json file.")
	var readRange string
	flag.StringVar(&readRange, "ReadRange", "", "ReadRange from the spread sheet such as Sheet1!A1:A reads all elements in column A sheet 1.")
	var bucketName string
	flag.StringVar(&bucketName, "BucketName", "", "Bucket Name to read from in GCS.")

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

	listOfPrs := readSpreadSheet(credentialsPath, spreadsheetID, readRange)
	context := con.Background()
	client, err := storage.NewClient(context)
	if err != nil {
		log.Fatalf("Unable to create new client: %v", err)
	}

	errorFinder, err := NewErrorFinder(client, bucketName)
	if err != nil {
		log.Fatalf("Unable to create new error finder: %v", err)
	}

	_, errorMap, warningMap, err := errorFinder.getContent(context, listOfPrs)
	if err != nil {
		fmt.Println(err)
	}
	writeCSV(errorMap, warningMap, fileName)
}
