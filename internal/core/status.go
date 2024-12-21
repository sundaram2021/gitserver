package core

import (
	"fmt"
	"gitserver/internal/utils"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func Status() {
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

	fmt.Printf("On branch %s\n\n", currentBranch)

	workingFiles, err := listWorkingDirectoryFiles(".")
	if err != nil {
		fmt.Println("Error reading working directory:", err)
		return
	}

	stagedFiles := getStagedFiles()

	var modifiedFiles []string
	var untrackedFiles []string

	for _, file := range workingFiles {
		if hash, isStaged := stagedFiles[file]; isStaged {
			currentHash, err := utils.GenerateFileHash(file)
			if err != nil {
				fmt.Println("Error reading file:", err)
				continue
			}
			if currentHash != hash {
				modifiedFiles = append(modifiedFiles, file) // The file has been modified since staging
			}
		} else {
			untrackedFiles = append(untrackedFiles, file) // File is not tracked
		}
	}

	if len(stagedFiles) > 0 {
		fmt.Println("Staged changes:")
		for file := range stagedFiles {
			fmt.Println("\t", file)
		}
	}
	if len(modifiedFiles) > 0 {
		fmt.Println("Modified (unstaged) changes:")
		for _, file := range modifiedFiles {
			fmt.Println("\t", file)
		}
	}
	if len(untrackedFiles) > 0 {
		fmt.Println("Untracked files:")
		for _, file := range untrackedFiles {
			fmt.Println("\t", file)
		}
	}
	if len(stagedFiles) == 0 && len(modifiedFiles) == 0 && len(untrackedFiles) == 0 {
		fmt.Println("No changes in the working directory.")
	}
}

func listWorkingDirectoryFiles(path string) ([]string, error) {
	var files []string
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || strings.HasPrefix(path, ".mygitserver") {
			return nil
		}
		files = append(files, path)
		return nil
	})
	return files, err
}

func getStagedFiles() map[string]string {
	stagedFiles := make(map[string]string)
	objectDir := filepath.Join(".mygitserver", "objects")

	files, err := ioutil.ReadDir(objectDir)
	if err != nil {
		fmt.Println("Error reading objects directory:", err)
		return stagedFiles
	}

	for _, file := range files {
		stagedFiles[file.Name()] = file.Name() // Store file hash for comparison
	}
	return stagedFiles
}
