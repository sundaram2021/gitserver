package core

import (
	"fmt"
	"gitserver/internal/utils"
	"io/ioutil"
	"path/filepath"
	"strings"
)

func Diff() {
	headContent, err := ioutil.ReadFile(".mygitserver/HEAD")
	if err != nil {
		fmt.Println("Error reading HEAD:", err)
		return
	}

	headRef := strings.TrimSpace(string(headContent))
	var currentBranch string
	if strings.HasPrefix(headRef, "ref: ") {
		currentBranch = strings.TrimPrefix(headRef, "ref: refs/heads/")
	} else {
		currentBranch = "detached"
	}

	commitHash, err := ioutil.ReadFile(filepath.Join(".mygitserver", "refs", "heads", currentBranch))
	if err != nil {
		fmt.Println("Error reading commit hash:", err)
		return
	}

	stagedFiles := getStagedFiles()

	workingFiles, err := listWorkingDirectoryFiles(".")
	if err != nil {
		fmt.Println("Error reading working directory:", err)
		return
	}

	var unstagedChanges []string
	for _, file := range workingFiles {
		currentHash, err := utils.GenerateFileHash(file)
		if err != nil {
			fmt.Println("Error generating file hash for", file, ":", err)
			continue
		}

		if stagedHash, isStaged := stagedFiles[file]; isStaged {
			if currentHash != stagedHash {
				unstagedChanges = append(unstagedChanges, file)
			}
		}
	}

	var stagedChanges []string
	for file, stagedHash := range stagedFiles {
		committedFilePath := filepath.Join(".mygitserver", "objects", strings.TrimSpace(string(commitHash)))
		committedHash, err := utils.GenerateFileHash(committedFilePath)
		if err != nil || stagedHash != committedHash {
			stagedChanges = append(stagedChanges, file)
		}
	}

	if len(unstagedChanges) > 0 {
		fmt.Println("Unstaged changes (working directory vs staging area):")
		for _, file := range unstagedChanges {
			fmt.Println("\t", file)
		}
	}

	if len(stagedChanges) > 0 {
		fmt.Println("Staged changes (staging area vs last commit):")
		for _, file := range stagedChanges {
			fmt.Println("\t", file)
		}
	}

	if len(unstagedChanges) == 0 && len(stagedChanges) == 0 {
		fmt.Println("No differences found.")
	}
}
