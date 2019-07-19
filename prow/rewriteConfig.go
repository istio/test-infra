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
	"flag"
	"io/ioutil"
	"log"
	"regexp"
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
func appendAtIndex(original string, element string, index int) string {
	prev := original[:index]
	apres := original[index:]
	return prev + element + apres
}

// Find index of master section in each branch split for the new branch and
// get content from master section with mandatory "merges-blocked-needs-admin" part.
func getContentOfNewBranch(branchContent string, newBranch string) (int, string) {
	branchLines := strings.Split(branchContent, "\n")
	if len(branchLines) < 2 {
		return -1, ""
	}

	firstBranchLine := branchLines[1]
	spacesForBranchNum := countLeadingSpace(firstBranchLine)

	spacesForBranch := strings.Repeat(" ", spacesForBranchNum)
	masterLine := spacesForBranch + "master:\n"
	masterStart := strings.Index(branchContent, masterLine)

	nonMasterContent := branchContent[masterStart+1:]

	generalBranchRe := regexp.MustCompile("\n" + spacesForBranch + "\\S+:\n")
	otherBranches := generalBranchRe.FindStringSubmatchIndex(nonMasterContent)

	var masterContent string
	if otherBranches == nil {
		masterContent = branchContent[masterStart:]
	} else {
		masterContent = branchContent[masterStart:otherBranches[0]]
	}

	masterContent = strings.Join(strings.Split(masterContent, "\n")[1:], "\n")
	contextRe := regexp.MustCompile("\\s+contexts:\n")
	contextInd := contextRe.FindStringSubmatchIndex(masterContent)
	if contextInd == nil {
		masterContentSplit := strings.Split(masterContent, "\n")
		masterContentLine := masterContentSplit[0]
		spacesInMaster := countLeadingSpace(masterContentLine)

		tab := spacesInMaster - spacesForBranchNum
		statusCheckString := strings.Repeat(" ", spacesInMaster) + "required_status_checks:\n"
		contextsString := strings.Repeat(" ", spacesInMaster+tab) + "contexts:\n"

		if strings.Compare(masterContentSplit[len(masterContentSplit)-1], "") != 0 {
			masterContent = masterContent + "\n"
		}

		masterContent = masterContent + statusCheckString + contextsString
	}

	contextInd = contextRe.FindStringSubmatchIndex(masterContent)
	contextStart := contextInd[0] + 1
	contextContent := masterContent[contextStart:]
	if !strings.Contains(contextContent, "- \"merges-blocked-needs-admin\"") {
		contextSpaces := countLeadingSpace(masterContent[contextStart:contextInd[1]])
		adminLine := strings.Repeat(" ", contextSpaces) + "- \"merges-blocked-needs-admin\"\n"

		if strings.Compare(string(masterContent[len(masterContent)-1]), "\n") != 0 {
			masterContent = masterContent + "\n"
		}

		masterContent = masterContent + adminLine
	}

	withinMaster := spacesForBranch + newBranch + ":\n" + masterContent
	return masterStart, withinMaster
}

func findRepo(source []byte, repos []string, newBranch string) string {
	sourceString := string(source)
	eachRepos := strings.Split(sourceString, "repos:")
	if len(eachRepos) < 1 {
		return sourceString
	}
	for i := 1; i < len(eachRepos); i++ {
		repoSection := eachRepos[i]
		repoLines := strings.Split(repoSection, "\n")
		var numSpaceForRepo int
		if strings.Compare(repoLines[0], "") == 0 && len(repoLines) > 1 {
			numSpaceForRepo = countLeadingSpace(repoLines[1])
		} else {
			numSpaceForRepo = countLeadingSpace(repoLines[0])
		}
		leadingSpace := strings.Repeat(" ", numSpaceForRepo)
		for _, repo := range repos {
			repoLine := leadingSpace + repo + ":\n"
			repoSplit := strings.Split(repoSection, repoLine)
			if len(repoSplit) == 1 {
				continue
			}

			generalReposRe := regexp.MustCompile("\n" + leadingSpace + "\\S+:\n")
			for j := 1; j < len(repoSplit); j++ {
				curRepoContent := repoSplit[j]
				otherRepos := generalReposRe.FindStringSubmatchIndex(curRepoContent)

				endIndex := len(curRepoContent)
				if otherRepos != nil {
					endIndex = otherRepos[0]
				}
				restContent := ""
				if endIndex != len(curRepoContent) {
					restContent = curRepoContent[endIndex:]
				}

				validRepoContent := curRepoContent[:endIndex]
				validRepoContent = findMaster(validRepoContent, newBranch)
				curRepoContent = validRepoContent + restContent
				repoSplit[j] = curRepoContent
			}
			repoSection = strings.Join(repoSplit, repoLine)
		}
		eachRepos[i] = repoSection
	}

	resultString := strings.Join(eachRepos, "repos:")
	return resultString
}

// Convert read result into string and split it based on "branches:" to make it
// simpler to find "master:" and its contents. Get output from getContentOfNewBranch().
// Add new branch content before master section and rejoin the string.
// Only add new branches to repos proxy, istio, istio-releases
func findMaster(sourceString string, newBranch string) string {
	eachBranches := strings.Split(sourceString, "branches:")
	for i := 0; i < len(eachBranches); i++ {
		branch := eachBranches[i]
		index, branchContent := getContentOfNewBranch(branch, newBranch)
		if index == -1 {
			continue
		}
		branch = appendAtIndex(branch, branchContent, index)
		eachBranches[i] = branch
	}

	resultString := strings.Join(eachBranches, "branches:")
	return resultString
}

// Call function with go run rewriteConfig.go <new_branch_name>.
func main() {
	var fileName string
	flag.StringVar(&fileName, "InputFileName", "", "A file with original config information.")
	var branchName string
	flag.StringVar(&branchName, "NewBranchName", "", "A new branch to add to the original file.")
	var repoNames string
	flag.StringVar(&repoNames, "ReposeToAdd", "proxy, istio, istio-releases", "Repo names to add new branch to, separated by `,`.")

	flag.Parse()

	if strings.Compare(fileName, "") == 0 {
		log.Fatal("Please enter an input file name.")
	}

	if strings.Compare(branchName, "") == 0 {
		log.Fatal("Please enter a new branch name to add to original file.")
	}

	repos := strings.Split(repoNames, ",")

	source, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatalf("Unable to read input file: %v", err)
	}

	resultString := findRepo(source, repos, branchName)
	resultBytes := []byte(resultString)

	err = ioutil.WriteFile(fileName, resultBytes, 0600)
	if err != nil {
		log.Fatalf("Error when writing file: %v", err)
	}
}
