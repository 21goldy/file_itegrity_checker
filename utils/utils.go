package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// HashEntry stores a file hash + timestamp
type HashEntry struct {
	Hash      string
	Timestamp string
}

var (
	hashMap      = make(map[string][]HashEntry)
	watching     = false
	stopWatchSig = make(chan bool)
)

/*

type hash-history struct{
  UpdatedTimeStamp time.Time `json:"timestamp"`
  hash string `json:"hash"` }

setTimeStamp(filePath string, hash string, timestamp time.Timestamp){
	filehash = sha-256.hashstring(strings.TrimSpace(filepath),)
	this function should append `type hash-history list` to a file with the name filehash.json
}

getHistory(filepath string){
	filehash = sha-256.hashstring(strings.TrimSpace(filepath),)
	fetch data from filehash.json, and display it
}

*/

// --- Watches over the file ---
func WatchFile(filePath string) {
	if watching {
		fmt.Println("⚠️  Already watching another file. Stop it first.")
		return
	}
	watching = true
	fmt.Printf("👁️  Now watching: %s\n", filePath)

	for {
		hashHex, timestamp, err := ComputeFileHash(filePath)
		if err != nil {
			fmt.Printf("❌ Error reading file: %v\n", err)
			time.Sleep(3 * time.Second)
			continue
		}

		lastHash := GetLastHash(filePath)
		if lastHash == "" {
			AddNewEntry(filePath, hashHex, timestamp)
			fmt.Printf("[%s] ✅ Initial hash stored.\n", timestamp)
		} else if hashHex != lastHash {
			fmt.Println("🔄 File changed!")
			fmt.Printf("🕒 [%s] New hash recorded.\n", timestamp)
			AddNewEntry(filePath, hashHex, timestamp)
		} else {
			UpdateTimestamp(filePath, timestamp)
			fmt.Printf("[%s] No change detected.\n", timestamp)
		}

		StoreLatestHash(filePath, hashHex)
		time.Sleep(5 * time.Second)
	}
}

// --- Manual single-file processing ---
func ProcessFile(filePath string) {
	hashHex, timestamp, err := ComputeFileHash(filePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error processing file:", err)
		return
	}

	lastHash := GetLastHash(filePath)
	switch {
	case lastHash == "":
		AddNewEntry(filePath, hashHex, timestamp)
		fmt.Println("✅ New file hash added.")
	case hashHex != lastHash:
		AddNewEntry(filePath, hashHex, timestamp)
		fmt.Println("⚠️  Hash changed! Added new entry.")
	default:
		UpdateTimestamp(filePath, timestamp)
		fmt.Println("⏱️  Hash unchanged — timestamp updated.")
	}

	StoreLatestHash(filePath, hashHex)
}

// --- Computes file hash ---
func ComputeFileHash(filePath string) (string, string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", "", fmt.Errorf("failed to hash file: %w", err)
	}

	hashHex := hex.EncodeToString(hasher.Sum(nil))
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	return hashHex, timestamp, nil
}

// --- Helper functions ---
func AddNewEntry(filePath, hashHex, timestamp string) {
	entry := HashEntry{Hash: hashHex, Timestamp: timestamp}
	hashMap[filePath] = append(hashMap[filePath], entry)
}

// --- Updates time stamp ---
func UpdateTimestamp(filePath, timestamp string) {
	if entries, ok := hashMap[filePath]; ok && len(entries) > 0 {
		hashMap[filePath][len(entries)-1].Timestamp = timestamp
	}
}

// --- This fetches last hash (CHECK) ---
func GetLastHash(filePath string) string {
	if entries, ok := hashMap[filePath]; ok && len(entries) > 0 {
		return entries[len(entries)-1].Hash
	}
	return ""
}

// --- This prints the hash history for the file ---
func PrintHashHistory(filePath string) {
	if entries, ok := hashMap[filePath]; ok {
		fmt.Println("\n📜 Hash History:")
		for i, e := range entries {
			fmt.Printf("%d. %s  (%s)\n", i+1, e.Hash, e.Timestamp)
		}
	} else {
		fmt.Println("No hash history found for this file.")
	}
}

// --- File I/O for persistent hashes ---
func ReadLatestHash(filePath string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	hashFile := filepath.Join(home, ".secret_hashes", filepath.Base(filePath)+".sha256")
	data, err := os.ReadFile(hashFile)
	if err != nil {
		return ""
	}
	parts := strings.Fields(string(data))
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// --- Store latest hashes ---
func StoreLatestHash(filePath, hashHex string) {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to find home directory:", err)
		return
	}
	secretDir := filepath.Join(home, ".secret_hashes")
	os.MkdirAll(secretDir, 0700)
	hashFile := filepath.Join(secretDir, filepath.Base(filePath)+".sha256")
	content := fmt.Sprintf("%s  %s\n", hashHex, time.Now().Format("2006-01-02 15:04:05"))
	os.WriteFile(hashFile, []byte(content), 0600)
}
