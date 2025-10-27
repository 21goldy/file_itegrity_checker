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

func main() {
	printIntro()
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("\n> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch {
		case input == "exit":
			fmt.Println("👋 Exiting program...")
			return

		case input == "help":
			printHelp()

		case strings.HasPrefix(input, "watch "):
			filePath := strings.TrimSpace(strings.TrimPrefix(input, "watch "))
			go watchFile(filePath)

		case input == "stopwatch":
			if watching {
				stopWatchSig <- true
			} else {
				fmt.Println("⚠️  No file is currently being watched.")
			}

		case input == "":
			continue

		default:
			processFile(input)
			printHashHistory(input)
		}
	}
}

// --- Watch file changes in real-time ---
func watchFile(filePath string) {
	if watching {
		fmt.Println("⚠️  Already watching another file. Stop it first.")
		return
	}
	watching = true
	fmt.Printf("👁️  Now watching: %s\n", filePath)

	for {
		select {
		case <-stopWatchSig:
			fmt.Println("🛑 Stopped watching the file.")
			watching = false
			return

		default:
			hashHex, timestamp, err := computeFileHash(filePath)
			if err != nil {
				fmt.Printf("❌ Error reading file: %v\n", err)
				time.Sleep(3 * time.Second)
				continue
			}

			lastHash := getLastHash(filePath)
			if lastHash == "" {
				addNewEntry(filePath, hashHex, timestamp)
				fmt.Printf("[%s] ✅ Initial hash stored.\n", timestamp)
			} else if hashHex != lastHash {
				fmt.Println("🔄 File changed!")
				fmt.Printf("🕒 [%s] New hash recorded.\n", timestamp)
				addNewEntry(filePath, hashHex, timestamp)
			} else {
				updateTimestamp(filePath, timestamp)
				fmt.Printf("[%s] No change detected.\n", timestamp)
			}

			storeLatestHash(filePath, hashHex)
			time.Sleep(5 * time.Second)
		}
	}
}

// --- Manual single-file processing ---
func processFile(filePath string) {
	hashHex, timestamp, err := computeFileHash(filePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error processing file:", err)
		return
	}

	lastHash := getLastHash(filePath)
	switch {
	case lastHash == "":
		addNewEntry(filePath, hashHex, timestamp)
		fmt.Println("✅ New file hash added.")
	case hashHex != lastHash:
		addNewEntry(filePath, hashHex, timestamp)
		fmt.Println("⚠️  Hash changed! Added new entry.")
	default:
		updateTimestamp(filePath, timestamp)
		fmt.Println("⏱️  Hash unchanged — timestamp updated.")
	}

	storeLatestHash(filePath, hashHex)
}

func computeFileHash(filePath string) (string, string, error) {
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

func addNewEntry(filePath, hashHex, timestamp string) {
	entry := HashEntry{Hash: hashHex, Timestamp: timestamp}
	hashMap[filePath] = append(hashMap[filePath], entry)
}

func updateTimestamp(filePath, timestamp string) {
	if entries, ok := hashMap[filePath]; ok && len(entries) > 0 {
		hashMap[filePath][len(entries)-1].Timestamp = timestamp
	}
}

func getLastHash(filePath string) string {
	if entries, ok := hashMap[filePath]; ok && len(entries) > 0 {
		return entries[len(entries)-1].Hash
	}
	return ""
}

func printHashHistory(filePath string) {
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
func readLatestHash(filePath string) string {
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

func storeLatestHash(filePath, hashHex string) {
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

// --- CLI intro and help section ---
func printIntro() {
	fmt.Println("🔒 File Hash Watcher CLI")
	fmt.Println("─────────────────────────────")
	fmt.Println("Keep track of file integrity using SHA-256 hashing.")
	fmt.Println("Type 'help' anytime to see all commands.")
	fmt.Println()
	printHelp()
}

func printHelp() {
	fmt.Println("📘 Commands Guide:")
	fmt.Println("──────────────────")
	fmt.Println(" <file-path>         → Compute and record hash of a file once")
	fmt.Println(" watch <file-path>   → Continuously monitor file for changes")
	fmt.Println(" stopwatch           → Stop watching the currently watched file")
	fmt.Println(" help                → Show this help menu again")
	fmt.Println(" exit                → Quit the program")
	fmt.Println()
	fmt.Println("🪶 Notes:")
	fmt.Println(" - When a file's hash stays the same, only its timestamp updates.")
	fmt.Println(" - All hashes are stored in ~/.secret_hashes for reference.")
	fmt.Println(" - The tool avoids duplicate entries for unchanged files.")
	fmt.Println("──────────────────")
}
