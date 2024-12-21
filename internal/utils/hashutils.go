package utils

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
	"strings"
)

func GenerateHash(data string) string {
	h := sha1.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func JoinMessage(args []string) string {
	return "Commit message: " + strings.Join(args, " ")
}

func GenerateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha1.New()
	_, err = io.Copy(hasher, file)
	if err != nil {
		return "", err
	}

	hash := hasher.Sum(nil)
	return hex.EncodeToString(hash), nil
}
