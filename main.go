package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter the filepath: ")
	_filepath, _ := reader.ReadString('\n') //Note: This line read until the user hits enter
	_filepath = strings.TrimSpace(_filepath)

	src, err := os.Open(_filepath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to open file:", err)
		return
	}
	defer src.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, src); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to has file:", err)
	}

	sum := hasher.Sum(nil)
	hashHex := hex.EncodeToString(sum)

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to find home directory:", err)
		return
	}

	secretDir := filepath.Join(home, ".secret_hashes")    // hidden folder
	if err := os.MkdirAll(secretDir, 0o700); err != nil { // rwx------ for owner
		fmt.Fprintln(os.Stderr, "failed to create secret folder:", err)
		return
	}

	base := filepath.Base(_filepath)
	hashFileName := base + ".sha256"
	hashFilePath := filepath.Join(secretDir, hashFileName)

	tmpPath := hashFilePath + ".tmp"
	hashContents := fmt.Sprintf("%s  %s\n", hashHex, base) // common format: "<hash>  <filename>"

	if err := os.WriteFile(tmpPath, []byte(hashContents), 0o600); err != nil { // rw-------
		fmt.Fprintln(os.Stderr, "failed to write temp hash file:", err)
		return
	}
	if err := os.Rename(tmpPath, hashFilePath); err != nil {
		fmt.Fprintln(os.Stderr, "failed to move hash file into place:", err)
		// try to cleanup tmp
		_ = os.Remove(tmpPath)
		return
	}

	fmt.Println("SHA-256 hash computed and stored at:", hashFilePath)
}
