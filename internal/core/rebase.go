package core

import (
	"bufio"
	"fmt"
	"gitserver/internal/utils"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func AbortRebase() {
	err := os.Remove(filepath.Join(".mygitserver", "rebase"))
	if err != nil {
		fmt.Println("Error aborting rebase:", err)
		return
	}

	fmt.Println("Rebase aborted.")
}

func getLatestCommitHash(branch string) (string, error) {
	commitHashPath := filepath.Join(".mygitserver", "refs", "heads", branch)
	commitHashBytes, err := ioutil.ReadFile(commitHashPath)
	if err != nil {
		return "", fmt.Errorf("could not read commit hash for branch '%s'", branch)
	}
	return strings.TrimSpace(string(commitHashBytes)), nil
}

func getCommitsAfter(baseCommit string, branch string) ([]string, error) {
	commits, err := getCommitsFromBranch(branch)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, commit := range commits {
		if commit == baseCommit {
			break
		}
		result = append(result, commit)
	}
	return result, nil
}

func findBaseCommit(sourceBranch, targetBranch string) (string, error) {
	sourceCommits, err := getCommitsFromBranch(sourceBranch)
	if err != nil {
		return "", err
	}

	targetCommits, err := getCommitsFromBranch(targetBranch)
	if err != nil {
		return "", err
	}

	for _, sourceCommit := range sourceCommits {
		for _, targetCommit := range targetCommits {
			if sourceCommit == targetCommit {
				return sourceCommit, nil // Found common commit
			}
		}
	}

	return "", fmt.Errorf("no common commit found between '%s' and '%s'", sourceBranch, targetBranch)
}

func getCommitsFromBranch(branch string) ([]string, error) {
	var commits []string
	commitHash, err := getLatestCommitHash(branch)
	if err != nil {
		return nil, err
	}

	for commitHash != "" {
		commits = append(commits, commitHash)
		commitHash = getParentCommit(commitHash)
	}
	return commits, nil
}

func reapplyCommit(commitHash, baseCommitHash string) {
	commitPath := filepath.Join(".mygitserver", "objects", commitHash)
	commitContent, err := ioutil.ReadFile(commitPath)
	if err != nil {
		fmt.Printf("Error reading commit '%s': %v\n", commitHash, err)
		return
	}

	newCommitContent := strings.Replace(string(commitContent), fmt.Sprintf("parent: %s", getParentCommit(commitHash)), fmt.Sprintf("parent: %s", baseCommitHash), 1)

	newCommitHash := utils.GenerateHash(newCommitContent)

	err = ioutil.WriteFile(filepath.Join(".mygitserver", "objects", newCommitHash), []byte(newCommitContent), 0644)
	if err != nil {
		fmt.Printf("Error writing new commit '%s': %v\n", newCommitHash, err)
		return
	}

	fmt.Printf("Reapplied commit '%s' as '%s'\n", commitHash, newCommitHash)
}

func squashCommit(previousCommitHash, commitHash string) {
	previousCommitPath := filepath.Join(".mygitserver", "objects", previousCommitHash)
	commitPath := filepath.Join(".mygitserver", "objects", commitHash)

	previousContent, err := ioutil.ReadFile(previousCommitPath)
	if err != nil {
		fmt.Printf("Error reading commit '%s': %v\n", previousCommitHash, err)
		return
	}

	commitContent, err := ioutil.ReadFile(commitPath)
	if err != nil {
		fmt.Printf("Error reading commit '%s': %v\n", commitHash, err)
		return
	}

	squashedContent := string(previousContent) + "\n" + string(commitContent)

	newCommitHash := utils.GenerateHash(squashedContent)

	err = ioutil.WriteFile(filepath.Join(".mygitserver", "objects", newCommitHash), []byte(squashedContent), 0644)
	if err != nil {
		fmt.Printf("Error writing squashed commit: %v\n", err)
		return
	}

	fmt.Printf("Squashed commit '%s' with '%s' into '%s'\n", previousCommitHash, commitHash, newCommitHash)
}

func editCommit(commitHash string) {
	commitPath := filepath.Join(".mygitserver", "objects", commitHash)
	content, err := ioutil.ReadFile(commitPath)
	if err != nil {
		fmt.Printf("Error reading commit '%s': %v\n", commitHash, err)
		return
	}

	fmt.Println("Current commit content:")
	fmt.Println(string(content))

	fmt.Println("Enter new commit content (leave empty to keep unchanged):")
	reader := bufio.NewReader(os.Stdin)
	newContent, _ := reader.ReadString('\n')
	newContent = strings.TrimSpace(newContent)

	if newContent != "" {
		err = ioutil.WriteFile(commitPath, []byte(newContent), 0644)
		if err != nil {
			fmt.Printf("Error saving edited commit '%s': %v\n", commitHash, err)
			return
		}
		fmt.Printf("Commit '%s' edited successfully.\n", commitHash)
	} else {
		fmt.Println("No changes made to the commit.")
	}
}

func InteractiveRebase(sourceBranch, targetBranch string) {
	targetCommitHash, err := getLatestCommitHash(targetBranch)
	if err != nil {
		fmt.Printf("Error getting latest commit for target branch '%s': %v\n", targetBranch, err)
		return
	}

	baseCommitHash, err := findBaseCommit(sourceBranch, targetBranch)
	if err != nil {
		fmt.Printf("Error finding base commit between '%s' and '%s': %v\n", sourceBranch, targetBranch, err)
		return
	}

	sourceCommits, err := getCommitsAfter(baseCommitHash, sourceBranch)
	if err != nil {
		fmt.Printf("Error getting commits after base commit from source branch '%s': %v\n", sourceBranch, err)
		return
	}

	actions := promptUserForActions(sourceCommits)

	for _, action := range actions {
		switch action.Action {
		case "pick":
			conflictResolved := resolveConflict(action.CommitHash, targetCommitHash)
			if conflictResolved {
				fmt.Printf("Conflict automatically resolved between '%s' and '%s'.\n", action.CommitHash, targetBranch)
			} else {
				fmt.Printf("Conflict detected while rebasing commit '%s' onto '%s'.\n", action.CommitHash, targetBranch)
				fmt.Println("Resolve the conflicts and run 'gitserver rebase --continue' to resume, or 'gitserver rebase --abort' to abort.")
				pauseRebase(action.CommitHash, targetCommitHash)
				return
			}
			reapplyCommit(action.CommitHash, targetCommitHash)
			targetCommitHash = action.CommitHash

		case "squash":
			if action.PreviousCommitHash == "" {
				fmt.Println("Squash requires a previous commit to combine with.")
				return
			}
			conflictResolved := resolveConflict(action.PreviousCommitHash, action.CommitHash)
			if conflictResolved {
				squashCommit(action.PreviousCommitHash, action.CommitHash)
			} else {
				fmt.Printf("Conflict detected during squash of '%s' and '%s'. Resolve the conflict and continue.\n", action.PreviousCommitHash, action.CommitHash)
				pauseRebase(action.PreviousCommitHash, action.CommitHash)
				return
			}

		case "edit":
			editCommit(action.CommitHash)
		case "drop":
			fmt.Printf("Dropping commit '%s'\n", action.CommitHash)
		}
	}

	err = ioutil.WriteFile(filepath.Join(".mygitserver", "refs", "heads", sourceBranch), []byte(targetCommitHash), 0644)
	if err != nil {
		fmt.Printf("Error updating source branch '%s' to new commit: %v\n", sourceBranch, err)
		return
	}

	fmt.Printf("Successfully rebased branch '%s' onto '%s' interactively.\n", sourceBranch, targetBranch)
}

func resolveConflict(commitHash1, commitHash2 string) bool {
	commit1Path := filepath.Join(".mygitserver", "objects", commitHash1)
	commit2Path := filepath.Join(".mygitserver", "objects", commitHash2)

	content1, err := ioutil.ReadFile(commit1Path)
	if err != nil {
		fmt.Printf("Error reading commit '%s': %v\n", commitHash1, err)
		return false
	}

	content2, err := ioutil.ReadFile(commit2Path)
	if err != nil {
		fmt.Printf("Error reading commit '%s': %v\n", commitHash2, err)
		return false
	}

	if string(content1) == string(content2) {
		return true
	}

	markConflict(commitHash1, commitHash2, string(content1), string(content2))
	return false
}

func markConflict(commitHash1, commitHash2, content1, content2 string) {
	conflictContent := fmt.Sprintf("<<<<<<< %s\n%s\n=======\n%s\n>>>>>> %s\n",
		commitHash1, content1, content2, commitHash2)

	conflictFilePath := filepath.Join(".", "conflicted_file.txt")
	err := ioutil.WriteFile(conflictFilePath, []byte(conflictContent), 0644)
	if err != nil {
		fmt.Println("Error writing conflict file:", err)
		return
	}

	fmt.Println("Conflict detected and markers have been added to 'conflicted_file.txt'.")
	fmt.Println("Please resolve the conflicts and run 'gitserver rebase --continue'.")
}

func promptUserForActions(commits []string) []Action {
	var actions []Action
	fmt.Println("Interactive Rebase - Choose actions for each commit:")
	fmt.Println("Available actions: 'pick', 'squash', 'edit', 'drop'")

	previousCommit := ""
	for _, commit := range commits {
		fmt.Printf("Commit: %s - Choose action: ", commit)
		reader := bufio.NewReader(os.Stdin)
		action, _ := reader.ReadString('\n')
		action = strings.TrimSpace(action)

		if action == "squash" && previousCommit == "" {
			fmt.Println("Error: Squash requires a previous commit to combine with.")
			return nil
		}

		actions = append(actions, Action{
			Action:             action,
			CommitHash:         commit,
			PreviousCommitHash: previousCommit,
		})

		previousCommit = commit
	}
	return actions
}

func pauseRebase(currentCommit, targetCommit string) {
	err := ioutil.WriteFile(filepath.Join(".mygitserver", "rebase"), []byte(fmt.Sprintf("%s %s", currentCommit, targetCommit)), 0644)
	if err != nil {
		fmt.Println("Error saving rebase state:", err)
	}
	fmt.Println("Rebase paused. Resolve conflicts and resume with 'gitserver rebase --continue'.")
}

func ResumeRebase() {
	rebaseState, err := ioutil.ReadFile(filepath.Join(".mygitserver", "rebase"))
	if err != nil {
		fmt.Println("No rebase in progress.")
		return
	}

	parts := strings.Split(string(rebaseState), " ")
	if len(parts) != 2 {
		fmt.Println("Invalid rebase state.")
		return
	}

	currentCommit := parts[0]
	targetCommit := parts[1]

	reapplyCommit(currentCommit, targetCommit)

	os.Remove(filepath.Join(".mygitserver", "rebase"))

	fmt.Println("Rebase continued successfully.")
}

type Action struct {
	Action             string
	CommitHash         string
	PreviousCommitHash string
}
