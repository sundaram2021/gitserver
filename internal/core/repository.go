package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

func InitializeRepository() {
	os.Mkdir(".mygitserver", 0755)
	os.MkdirAll(".mygitserver/refs/heads", 0755)
	os.Mkdir(".mygitserver/objects", 0755)

	err := ioutil.WriteFile(filepath.Join(".mygitserver", "HEAD"), []byte("ref: refs/heads/main"), 0644)
	if err != nil {
		fmt.Println("Error initializing repository: ", err)
		return
	}

	_, err = os.Create(filepath.Join(".mygitserver", "refs", "heads", "main"))
	if err != nil {
		fmt.Println("Error creating main branch: ", err)
		return
	}

	fmt.Println("Initialized empty Git repository in .mygitserver/")
}
