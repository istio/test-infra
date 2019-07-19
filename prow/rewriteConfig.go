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

	// Find branch line with the branch name of "master" as the start of the master content.
	spacesForBranch := strings.Repeat(" ", spacesForBranchNum)
	masterLine := spacesForBranch + "master:\n"
	masterStart := strings.Index(branchContent, masterLine)

	nonMasterContent := branchContent[masterStart+1:]

	// The end of master content is the start of other branch with the same number of spaces before branch name
	// and names not being the master name.
	generalBranchRe := regexp.MustCompile("\n" + spacesForBranch + "\\S+:\n")
	otherBranches := generalBranchRe.FindStringSubmatchIndex(nonMasterContent)

	// Get master content between start of master branch line and other branch line.
	// If no other branch lines are found then the end of master content is the end of the
	// branch section.
	var masterContent string
	if otherBranches == nil {
		masterContent = branchContent[masterStart:]
	} else {
		masterContent = branchContent[masterStart:otherBranches[0]]
	}

	// Eliminate the first line (the master brach name line) in master content.
	masterContent = strings.Join(strings.Split(masterContent, "\n")[1:], "\n")

	// Check if master content contains the line "contexts:", if not, add lines of
	// "required_status_checks:" and "contexts:".
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
			masterContent += "\n"
		}

		masterContent = masterContent + statusCheckString + contextsString
	}

	// Find new context index after adding context lines to original master content.
	contextInd = contextRe.FindStringSubmatchIndex(masterContent)
	contextStart := contextInd[0] + 1
	contextContent := masterContent[contextStart:]

	// Check if master content contains "merges-blocked-needs-admin", if not, add line to master content.
	if !strings.Contains(contextContent, "- \"merges-blocked-needs-admin\"") {
		contextSpaces := countLeadingSpace(masterContent[contextStart:contextInd[1]])
		adminLine := strings.Repeat(" ", contextSpaces) + "- \"merges-blocked-needs-admin\"\n"

		if strings.Compare(string(masterContent[len(masterContent)-1]), "\n") != 0 {
			masterContent += "\n"
		}

		masterContent += adminLine
	}

	// Rebuild content with new branch name line and new content for the branch.
	withinMaster := spacesForBranch + newBranch + ":\n" + masterContent
	return masterStart, withinMaster
}

// Process source byte slice to extract only the repos specified by users to add new branches to.
func findRepo(source []byte, reposToAddNewBranch []string, newBranch string) string {
	sourceString := string(source)
	// Split the source string by "repos:\n" to get the content between each "repos:\n" line.
	contentBetweenReposLines := strings.Split(sourceString, "repos:\n")
	if len(contentBetweenReposLines) < 1 {
		// If source string does not contain "repos:\n" line, then there is nowhere in source string
		// to add new branch content. Return source string without changing it.
		return sourceString
	}

	// The first element in `contentBetweenReposLines` does not contain content between repos, ignore it.
	for i := 1; i < len(contentBetweenReposLines); i++ {
		repoSection := contentBetweenReposLines[i]

		// Trim leading "\n"s from repo section.
		repoSectionWithLeadingLineFeedsTrimed := strings.Trim(repoSection, "\n")
		repoLines := strings.Split(repoSectionWithLeadingLineFeedsTrimed, "\n")

		// Count the number of spaces before lines for each specific repo name under the "repos:\n" line.
		numSpaceForEachRepoName := countLeadingSpace(repoLines[0])
		leadingSpaceForEachRepoName := strings.Repeat(" ", numSpaceForEachRepoName)

		// Loop through each repo from user input slice as interesting repo to add new branch content to.
		for _, repoToAddNewBranch := range reposToAddNewBranch {
			// The leading repo name line for interesting repo `repoToAddNewBranch` has the same number of
			// spaces as all other repo name lines.
			interestingRepoNameLine := leadingSpaceForEachRepoName + repoToAddNewBranch + ":\n"
			// Split the repo section by the interesting repo name line.
			splitRepoByInterestingRepoNameLine := strings.Split(repoSection, interestingRepoNameLine)
			if len(splitRepoByInterestingRepoNameLine) == 1 {
				// Current repoSection does not contain the current interesting repo line.
				// Skip the rest of the loop to find current interesting repo line in the next repo section.
				continue
			}

			// After splitted by the current interesting repo line, none of the sections inside `repoSplit`
			// contains the current interesting repo line. The next line containing the same number of spaces
			// as the current interesting repo line, repo name and ":\n" would be for a different non-interesting
			// repo.
			generalReposRe := regexp.MustCompile("\n" + leadingSpaceForEachRepoName + "\\S+:\n")
			// Skip the first element in `splitRepoByInterestingRepoNameLine` because it does not contain
			// the interesting repo name line.
			for j := 1; j < len(splitRepoByInterestingRepoNameLine); j++ {
				// Get each section that has the interesting repo name line before its first line.
				curRepoContentWithLeadingInterestingRepoContent := splitRepoByInterestingRepoNameLine[j]
				// Find the first occurrence of a non-interesting repo name line.
				nonInterestingRepoNameLineSlice := generalReposRe.FindStringSubmatchIndex(curRepoContentWithLeadingInterestingRepoContent)
				// If there is no non-interesting repo name line, the whole `curRepoContentWithLeadingInterestingRepoContent` is interesting repo.
				endIndex := len(curRepoContentWithLeadingInterestingRepoContent)

				// If there is non-interesting repo name line in `curRepoContentWithLeadingInterestingRepoContent`,
				// the end of the interesting repo section is before the first occurrence of non-interesting repo name line.
				if nonInterestingRepoNameLineSlice != nil {
					endIndex = nonInterestingRepoNameLineSlice[0]
				}

				// Get the non-interesting repo content in `curRepoContentWithLeadingInterestingRepoContent`
				// that begins with the `endIndex` of the interesting repo content.
				// If the `endIndex` for current interesting repo is the length of `curRepoContentWithLeadingInterestingRepoContent`
				// the non-interesting repo content is empty string.
				nonInterestingContentInCurRepoSection := ""
				if endIndex != len(curRepoContentWithLeadingInterestingRepoContent) {
					nonInterestingContentInCurRepoSection = curRepoContentWithLeadingInterestingRepoContent[endIndex:]
				}

				// Interesting repo content comes between the start of the `curRepoContentWithLeadingInterestingRepoContent`
				// and the `endIndex`.
				interestingRepoContent := curRepoContentWithLeadingInterestingRepoContent[:endIndex]
				// Get new repo content with new branch content coming from master branch in the interesting repo content.
				interestingRepoContent = findMaster(interestingRepoContent, newBranch)
				// Rebuild `curRepoContentWithLeadingInterestingRepoContent` with new interesting repo content.
				curRepoContentWithLeadingInterestingRepoContent = interestingRepoContent + nonInterestingContentInCurRepoSection
				splitRepoByInterestingRepoNameLine[j] = curRepoContentWithLeadingInterestingRepoContent
			}
			repoSection = strings.Join(splitRepoByInterestingRepoNameLine, interestingRepoNameLine)
		}
		contentBetweenReposLines[i] = repoSection
	}

	resultString := strings.Join(contentBetweenReposLines, "repos:\n")
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
