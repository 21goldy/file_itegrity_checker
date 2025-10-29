package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// HashEntry stores a file hash + timestamp
type HashEntry struct {
	Hash      string    `json:"hash"`
	Timestamp time.Time `json:"timestamp"`
}

var (
	hashMap      = make(map[string][]HashEntry)
	watching     = false
	stopWatchSig = make(chan bool)
)

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

		lastHash := GetLatestHash(filePath)
		if lastHash == "" {
			AddEntry(filePath, hashHex, timestamp)
			fmt.Printf("[%s] ✅ Initial hash stored.\n", timestamp.Format(time.RFC3339))
		} else if hashHex != lastHash {
			fmt.Println("🔄 File changed!")
			fmt.Printf("🕒 [%s] New hash recorded.\n", timestamp.Format(time.RFC3339))
			AddEntry(filePath, hashHex, timestamp)
		} else {
			fmt.Printf("[%s] No change detected.\n", timestamp.Format(time.RFC3339))
		}

		time.Sleep(5 * time.Second)
	}
}

// --- Creates unique hash file name ---
func createhashFileName(filepath string) string {
	hashfile := sha256.Sum256([]byte(filepath))
	return fmt.Sprintf("secret-%s-256.json", hex.EncodeToString(hashfile[:]))
}

// --- Computes file hash ---
func ComputeFileHash(filePath string) (string, time.Time, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", time.Time{}, fmt.Errorf("failed to hash file: %w", err)
	}

	hashHex := hex.EncodeToString(hasher.Sum(nil))
	timestamp := time.Now()
	return hashHex, timestamp, nil
}

// --- Add a new hash entry ---
func AddEntry(filePath, hashHex string, timestamp time.Time) error {
	hashfilename := createhashFileName(filePath)
	secretsDir := ".secret-hashes"

	// ensure directory exists
	if _, err := os.Stat(secretsDir); os.IsNotExist(err) {
		if err := os.Mkdir(secretsDir, 0755); err != nil {
			return fmt.Errorf("failed to create secrets dir: %w", err)
		}
	}

	filePathFull := filepath.Join(secretsDir, hashfilename)
	var entries []HashEntry

	// if file exists, load existing entries
	if _, err := os.Stat(filePathFull); err == nil {
		data, err := os.ReadFile(filePathFull)
		if err == nil && len(data) > 0 {
			_ = json.Unmarshal(data, &entries)
		}
	}

	// avoid duplicate hash (only update time if same hash)
	if len(entries) > 0 && entries[len(entries)-1].Hash == hashHex {
		entries[len(entries)-1].Timestamp = timestamp
	} else {
		entries = append(entries, HashEntry{Hash: hashHex, Timestamp: timestamp})
	}

	// write back
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal json: %w", err)
	}

	if err := os.WriteFile(filePathFull, data, 0644); err != nil {
		return fmt.Errorf("failed to write hash file: %w", err)
	}

	fmt.Printf("💾 Hash entry saved to %s\n", filePathFull)
	return nil
}

// --- Retrieve latest hash ---
func GetLatestHash(filePath string) string {
	hashfilename := createhashFileName(filePath)
	secretsDir := ".secret-hashes"
	filePathFull := filepath.Join(secretsDir, hashfilename)

	if _, err := os.Stat(filePathFull); os.IsNotExist(err) {
		return ""
	}

	data, err := os.ReadFile(filePathFull)
	if err != nil {
		fmt.Printf("⚠️  Failed to read hash file: %v\n", err)
		return ""
	}

	var entries []HashEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		fmt.Printf("⚠️  Failed to parse hash file: %v\n", err)
		return ""
	}

	if len(entries) == 0 {
		return ""
	}
	return entries[len(entries)-1].Hash
}

// --- Print all hash history ---
func PrintHashHistory(filePath string) {
	hashfilename := createhashFileName(filePath)
	secretsDir := ".secret-hashes"
	filePathFull := filepath.Join(secretsDir, hashfilename)

	if _, err := os.Stat(filePathFull); os.IsNotExist(err) {
		fmt.Println("ℹ️  No hash history found.")
		return
	}

	data, err := os.ReadFile(filePathFull)
	if err != nil {
		fmt.Printf("❌ Failed to read hash history: %v\n", err)
		return
	}

	var entries []HashEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		fmt.Printf("❌ Failed to parse hash history: %v\n", err)
		return
	}

	fmt.Println("📜 Hash History:")
	for _, e := range entries {
		fmt.Printf("🕒 %s → %s\n", e.Timestamp.Format(time.RFC3339), e.Hash)
	}
}
