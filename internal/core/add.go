package core

import (
	"fmt"
	"gitserver/internal/utils"
	"io"
	"os"
	"path/filepath"
)

func AddFile(paths []string) {
	for _, path := range paths {
		if _, err := os.Stat(path); err != nil {
			fmt.Printf("File %s does not exist.\n", path)
			continue
		}

		hash, err := generateFileHash(path)
		if err != nil {
			fmt.Println("Error hashing file:", err)
			continue
		}

		objectPath := filepath.Join(".git", "objects", hash)
		if _, err := os.Stat(objectPath); os.IsNotExist(err) {
			if err := copyFileToObject(path, objectPath); err != nil {
				fmt.Println("Error storing file:", err)
				continue
			}
			fmt.Printf("File %s added to staging (hash: %s).\n", path, hash)
		} else {
			fmt.Printf("File %s already staged (hash: %s).\n", path, hash)
		}
	}
}

func generateFileHash(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return utils.GenerateHash(string(content)), nil
}

func copyFileToObject(srcPath, dstPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
