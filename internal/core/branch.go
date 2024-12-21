package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func ListBranches() {
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

	branchesPath := filepath.Join(".mygitserver", "refs", "heads")
	files, err := ioutil.ReadDir(branchesPath)
	if err != nil {
		fmt.Println("Error reading branches:", err)
		return
	}

	fmt.Println("Branches:")
	for _, file := range files {
		branchName := file.Name()
		if branchName == currentBranch {
			fmt.Printf("* %s\n", branchName)
		} else {
			fmt.Printf("  %s\n", branchName)
		}
	}
}

func SwitchBranch(branchName string) {
	branchPath := filepath.Join(".mygitserver", "refs", "heads", branchName)
	if _, err := os.Stat(branchPath); os.IsNotExist(err) {
		fmt.Printf("Branch '%s' does not exist.\n", branchName)
		return
	}

	headContent := fmt.Sprintf("ref: refs/heads/%s", branchName)
	err := ioutil.WriteFile(".mygitserver/HEAD", []byte(headContent), 0644)
	if err != nil {
		fmt.Println("Error updating HEAD:", err)
		return
	}

	fmt.Printf("Switched to branch '%s'\n", branchName)
}

func CreateBranch(branchName string) {
	if branchName == "" {
		fmt.Println("Branch name cannot be empty")
		return
	}

	branchPath := filepath.Join(".mygitserver", "refs", "heads", branchName)
	if _, err := os.Stat(branchPath); err == nil {
		fmt.Printf("Branch '%s' already exists.\n", branchName)
		return
	}

	headContent, err := ioutil.ReadFile(".mygitserver/HEAD")
	if err != nil {
		fmt.Println("Error reading HEAD:", err)
		return
	}

	headRef := strings.TrimSpace(string(headContent))
	if strings.HasPrefix(headRef, "ref: ") {
		currentBranch := strings.TrimPrefix(headRef, "ref: ")
		currentBranchCommit, err := ioutil.ReadFile(filepath.Join(".mygitserver", currentBranch))
		if err != nil {
			fmt.Printf("Error reading current branch commit for %s: %v\n", currentBranch, err)
			return
		}
		headRef = string(currentBranchCommit)
	}

	err = ioutil.WriteFile(branchPath, []byte(headRef), 0644)
	if err != nil {
		fmt.Println("Error creating branch:", err)
		return
	}

	fmt.Printf("Branch '%s' created, pointing to commit %s\n", branchName, strings.TrimSpace(headRef))
}
