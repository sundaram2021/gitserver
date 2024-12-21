package core

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func Log() {
	headContent, err := ioutil.ReadFile(filepath.Join(".mygitserver", "HEAD"))
	if err != nil {
		fmt.Println("Error reading HEAD:", err)
		return
	}

	currentBranch := strings.TrimSpace(strings.TrimPrefix(string(headContent), "ref: refs/heads/"))
	branchPath := filepath.Join(".mygitserver", "refs", "heads", currentBranch)

	commitHash, err := ioutil.ReadFile(branchPath)
	if err != nil {
		fmt.Printf("Error reading branch '%s': %v\n", currentBranch, err)
		return
	}

	fmt.Printf("Commit history for branch '%s':\n", currentBranch)
	for len(commitHash) > 0 {
		commitFilePath := filepath.Join(".mygitserver", "objects", strings.TrimSpace(string(commitHash)))
		commitContent, err := ioutil.ReadFile(commitFilePath)
		if err != nil {
			fmt.Printf("Error reading commit object: %v\n", err)
			return
		}

		fmt.Printf("Commit: %s\n", strings.TrimSpace(string(commitHash)))
		fmt.Println(commitContent)

		scanner := bufio.NewScanner(strings.NewReader(string(commitContent)))
		parentHash := ""
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "parent:") {
				parentHash = strings.TrimSpace(strings.TrimPrefix(line, "parent:"))
				break
			}
		}

		commitHash = []byte(parentHash)
		if len(parentHash) == 0 {
			break
		}
	}
}

func printCommit(commitHash string) {
	commitPath := filepath.Join(".mygitserver", "objects", commitHash)
	file, err := os.Open(commitPath)
	if err != nil {
		fmt.Println("Error reading commit:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	fmt.Printf("commit %s\n", commitHash)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "message:") {
			fmt.Printf("    %s\n", strings.TrimPrefix(line, "message: "))
		} else if strings.HasPrefix(line, "timestamp:") {
			fmt.Printf("    Date: %s\n", strings.TrimPrefix(line, "timestamp: "))
		}
	}
	fmt.Println()
}

func getParentCommit(commitHash string) string {
	commitPath := filepath.Join(".mygitserver", "objects", commitHash)
	file, err := os.Open(commitPath)
	if err != nil {
		fmt.Println("Error reading commit:", err)
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "parent:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "parent: "))
		}
	}
	return ""
}
