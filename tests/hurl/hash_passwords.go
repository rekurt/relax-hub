// Helper program to generate bcrypt password hashes for test data
// Run: go run hash_passwords.go
package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	passwords := map[string]string{
		"admin":      "admin-password",
		"owner":      "owner-password",
		"rep":        "rep-password",
		"client":     "client-password",
		"client2":    "client2-password",
		"blocked":    "blocked-password",
	}

	for user, password := range passwords {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			fmt.Printf("Error hashing password for %s: %v\n", user, err)
			continue
		}
		fmt.Printf("%s: %s\n", user, string(hash))
	}
}
