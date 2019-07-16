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
	"io/ioutil"
	"log"
	"os"
	"strings"
)

// Count the number of spaces before a valid character in a string.
func countLeadingSpace(line string) int {
	i := 0
	for _, runeValue := range line {
		if runeValue == ' ' {
			i++
		} else {
			break
		}
	}
	return i
}

// Add element to the slice at the indexed position and move all the strings from
// the index one position to the right.
func appendAtIndex(original []string, element string, index int) []string {
	original = append(original, "")
	copy(original[index+1:], original[index:])
	original[index] = element

	return original
}

// Copy the content of one slice to another.
func copySlice(input []string) []string {
	newSlice := []string{}
	newSlice = append(newSlice, input...)

	return newSlice
}

// Find index of master section in each branch split for the new branch and
// get content from master section with mandatory "merges-blocked-needs-admin" part.
func getContentOfNewBranch(branchLines []string, newBranch string) (int, string) {
	if len(branchLines) < 2 {
		return -1, ""
	}
	firstBranchLine := branchLines[1]
	spacesForBranch := countLeadingSpace(firstBranchLine)
	masterStart := -1
	masterStop := -1
	for j := 1; j < len(branchLines); j++ {
		line := branchLines[j]
		if len(line) > spacesForBranch-1 {
			checkForMaster := line[spacesForBranch:]
			if strings.Contains(checkForMaster, "master:") {
				masterStart = j + 1
			}
		}

		if masterStart != -1 && masterStop == -1 && j > masterStart {
			if countLeadingSpace(line) <= spacesForBranch {
				masterStop = j - 1
			}
		}
	}

	if masterStart != -1 && masterStop == -1 {
		masterStop = len(branchLines) - 1
	}

	if masterStart != -1 && masterStop != -1 {

		contextStart := -1
		contextStop := -1
		contextStrings := -1

		secondBranchSpaces := countLeadingSpace(branchLines[masterStart])

		for m := 0; m < masterStop-masterStart+1; m++ {
			masterLine := branchLines[m+masterStart]
			if strings.Contains(masterLine, "contexts:") {
				contextStart = m + 1
				contextStrings = countLeadingSpace(masterLine)
			}

			if contextStart != -1 && contextStrings != -1 && contextStop == -1 && m > contextStart {

				if strings.Compare(string(masterLine[contextStrings]), "-") != 0 {
					contextStop = m
				}
			}
		}

		if contextStart != -1 && contextStop == -1 {
			contextStop = masterStop - masterStart + 1
		}
		containsNeedsAdmin := false
		newBranchSlice := copySlice(branchLines[masterStart : masterStop+1])
		if contextStart != -1 && contextStop != -1 {
			for n := contextStart; n < contextStop; n++ {
				contextLine := branchLines[n+masterStart]

				if strings.Contains(contextLine, "merges-blocked-needs-admin") {
					containsNeedsAdmin = true
					break
				}
			}

			if !containsNeedsAdmin {
				numSpaces := countLeadingSpace(newBranchSlice[contextStart])
				stringAdmin := strings.Repeat(" ", numSpaces) + "- \"merges-blocked-needs-admin\""

				newBranchSlice = appendAtIndex(newBranchSlice, stringAdmin, contextStop-1)
			}

		} else {

			if len(newBranchSlice[len(newBranchSlice)-1]) == countLeadingSpace(newBranchSlice[len(newBranchSlice)-1]) {
				newBranchSlice = newBranchSlice[:len(newBranchSlice)-1]
			}
			statusCheckString := strings.Repeat(" ", secondBranchSpaces) + "required_status_checks:"
			contextsString := strings.Repeat(" ", secondBranchSpaces+2) + "contexts:"
			needsAdminString := strings.Repeat(" ", secondBranchSpaces+2) + "- \"merges-blocked-needs-admin\""
			newBranchSlice = append(newBranchSlice, statusCheckString, contextsString, needsAdminString)
		}

		newBranchLine := strings.Repeat(" ", spacesForBranch) + newBranch + ":"
		newBranchSlice = appendAtIndex(newBranchSlice, newBranchLine, 0)
		masterBranch := strings.Join(newBranchSlice, "\n")
		return masterStart - 1, masterBranch
	}
	return -1, ""
}

// Convert read result into string and split it based on "branches:" to make it
// simpler to find "master:" and its contents. Get output from getContentOfNewBranch().
// Add new branch content before master section and rejoin the string.
func findMaster(source []byte, newBranch string) string {
	sourceString := string(source)
	eachBranches := strings.Split(sourceString, "branches:")
	for i := 0; i < len(eachBranches); i++ {
		branch := eachBranches[i]
		branchLines := strings.Split(branch, "\n")
		index, branchContent := getContentOfNewBranch(branchLines, newBranch)
		if index == -1 {
			continue
		}
		branchLines = appendAtIndex(branchLines, branchContent, index)

		branchString := strings.Join(branchLines, "\n")
		eachBranches[i] = branchString
	}

	resultString := strings.Join(eachBranches, "branches:")
	return resultString
}

// Call function with go run rewriteConfig.go <new_branch_name>.
func main() {
	if len(os.Args) < 2 {
		log.Println("Please provide new branch name")
		os.Exit(1)
	}
	newBranch := os.Args[1]
	filename := "config.yaml"

	source, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Println("error when reading file")
		os.Exit(1)
	}

	resultString := findMaster(source, newBranch)
	resultBytes := []byte(resultString)

	err = ioutil.WriteFile("config.yaml", resultBytes, 0600)
	if err != nil {
		log.Println("error when writing file")
	}
}
