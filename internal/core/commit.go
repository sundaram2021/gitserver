package core

import (
	"fmt"
	"gitserver/internal/utils"
	"io/ioutil"
	"time"
)

func CommitChanges(message []string) {
	commitHash := utils.GenerateHash(time.Now().String())
	commitMessage := utils.JoinMessage(message)

	ioutil.WriteFile(fmt.Sprintf(".git/objects/%s", commitHash), []byte(commitMessage), 0644)

	err := ioutil.WriteFile(".git/HEAD", []byte(commitHash), 0644)
	if err != nil {
		fmt.Println("Error updating HEAD:", err)
	}
	fmt.Println("Commit successful:", commitHash)
}
