package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
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
			AddEntry(filePath, hashHex, timestamp) // here
			fmt.Printf("[%s] ✅ Initial hash stored.\n", timestamp)
		} else if hashHex != lastHash {
			fmt.Println("🔄 File changed!")
			fmt.Printf("🕒 [%s] New hash recorded.\n", timestamp)
			AddEntry(filePath, hashHex, timestamp) // here
		} else {
			fmt.Printf("[%s] No change detected.\n", timestamp)
		}

		time.Sleep(5 * time.Second)
	}
}

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

func AddEntry(filePath, hashHex string, timestamp time.Time) error {
	hashfilename := createhashFileName(filePath) //secret-bajshqhhjgdhgduyqgdqshgdhuqgdwyggsbqhgd-256.json
	fmt.Println(hashfilename)
	// check if this file exists in .secrets-hashes
	// if yes
	// open the file
	// read the content
	// extract it to []HashEntry
	// append to the  with the hashhex adt timestamp
	// and write back to the file (basically append back to the file)
	// else
	// create a new slice of []HashEntry
	// append to the  with the hashhex adt timestamp
	// create the file and write the new json dara

	return nil
}

func GetLatestHash(filePath string) string {
	hashfilename := createhashFileName(filePath) //secret-bajshqhhjgdhgduyqgdqshgdhuqgdwyggsbqhgd-256.json
	fmt.Println(hashfilename)

	// check if the file extist
	// if
	// 	opent the file
	// extract the []HashEntry
	// returh []HashEntry[-1].Hash
	// else
	// 	return ""
	return ""
}

func PrintHashHistory(filePath string) {
	hashfilename := createhashFileName(filePath) //secret-bajshqhhjgdhgduyqgdqshgdhuqgdwyggsbqhgd-256.json
	fmt.Println(hashfilename)
	// check if the file extist
	// if
	// 	opent the file
	// extract the []HashEntry
	// for each entry :
	//    print (timestamp : hash )
}
