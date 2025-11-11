package main

import (
	"fmt"
	"os"

	apputils "github.com/detect-price-by-photo/backend/internal/utils"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run generate_password_hash.go <password>")
		fmt.Println("Example: go run generate_password_hash.go admin123")
		os.Exit(1)
	}

	password := os.Args[1]
	hasher := apputils.NewPasswordHasher()

	hash, err := hasher.Hash(password)
	if err != nil {
		fmt.Printf("Error hashing password: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Password: %s\n", password)
	fmt.Printf("Hash: %s\n", hash)
}
