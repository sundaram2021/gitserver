package core

import (
	"fmt"
	"gitserver/internal/utils"
	"io/ioutil"
	"path/filepath"
	"strings"
)

func MergeBranch(sourceBranch string) {
	headContent, err := ioutil.ReadFile(".mygitserver/HEAD")
	if err != nil {
		fmt.Println("Error reading HEAD:", err)
		return
	}

	headRef := strings.TrimSpace(string(headContent))
	if !strings.HasPrefix(headRef, "ref: ") {
		fmt.Println("HEAD is not pointing to a branch")
		return
	}
	currentBranch := strings.TrimPrefix(headRef, "ref: refs/heads/")

	sourceBranchRef := filepath.Join(".mygitserver", "refs", "heads", sourceBranch)
	sourceCommitHash, err := ioutil.ReadFile(sourceBranchRef)
	if err != nil {
		fmt.Printf("Error reading commit for branch %s: %v\n", sourceBranch, err)
		return
	}

	currentBranchRef := filepath.Join(".mygitserver", "refs", "heads", currentBranch)
	currentCommitHash, err := ioutil.ReadFile(currentBranchRef)
	if err != nil {
		fmt.Printf("Error reading commit for branch %s: %v\n", currentBranch, err)
		return
	}

	mergeCommitMessage := fmt.Sprintf("Merge branch '%s' into '%s'", sourceBranch, currentBranch)
	newCommitHash := createMergeCommit(string(currentCommitHash), string(sourceCommitHash), mergeCommitMessage)

	err = ioutil.WriteFile(currentBranchRef, []byte(newCommitHash), 0644)
	if err != nil {
		fmt.Println("Error updating current branch:", err)
		return
	}

	fmt.Printf("Successfully merged branch '%s' into '%s'. New commit: %s\n", sourceBranch, currentBranch, newCommitHash)
}

func createMergeCommit(parent1, parent2, message string) string {
	commitContent := fmt.Sprintf("Parent 1: %s\nParent 2: %s\nMessage: %s", parent1, parent2, message)
	newCommitHash := utils.GenerateHash(commitContent)

	err := ioutil.WriteFile(filepath.Join(".mygitserver", "objects", newCommitHash), []byte(commitContent), 0644)
	if err != nil {
		fmt.Println("Error writing new merge commit:", err)
	}

	return newCommitHash
}
