package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
)

func main() {
	// Generate a 32-byte secure key
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		log.Fatal("Error generating random key:", err)
	}

	// Print the key as a hexadecimal string
	secureKey := hex.EncodeToString(key)
	fmt.Printf("Your secure SESSION_KEY: %s\n", secureKey)
}
