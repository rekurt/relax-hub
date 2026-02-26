package tests

import (
	"fmt"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// VerifyPasswordHash is a helper function for password verification tests
func VerifyPasswordHash(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// TestAuthIntegration_PasswordHashing validates bcrypt password hashing works correctly for test data
func TestAuthIntegration_PasswordHashing(t *testing.T) {
	testCases := []struct {
		email    string
		password string
		hash     string
	}{
		{"admin@test.com", "admin-password", "$2a$10$2R./p9GTXF/325PJnx/tvOXaxMjJZhC/mEaAZJ3FRGQiZ4U6f/JBy"},
		{"owner@test.com", "owner-password", "$2a$10$WiOza7gHYyBIT2dfFJsW9eo.7hS8.kuwt05VdT3tQI2Mz8BwCoBX6"},
		{"representative@test.com", "rep-password", "$2a$10$5etUHu7KX4g4nel2jxrXgO06z.QkkqAGC25dCGPDcTZplJz3CXFY6"},
		{"client@test.com", "client-password", "$2a$10$xh7pDp6sitliSK6TJLzXj.O7dwHd8FoA.aN2B3KX6UxThvmySPvNu"},
		{"client2@test.com", "client2-password", "$2a$10$zrLx8jTPUlADKRGqWjlRM.A95yOBZnoM6LdE/eW9tMhfiZKeNoGve"},
		{"blocked@test.com", "blocked-password", "$2a$10$m5SaOpY7.Xm0/kIYrxVPhu0CsgEq2yi9dMKUsco37j0rYgDNFs/Hi"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s_%s", tc.email, tc.password), func(t *testing.T) {
			if !VerifyPasswordHash(tc.hash, tc.password) {
				t.Errorf("password verification failed for %s", tc.email)
			}
		})
	}
}

// TestAuthIntegration_IncorrectPasswordRejection validates that wrong passwords are rejected
func TestAuthIntegration_IncorrectPasswordRejection(t *testing.T) {
	hash := "$2a$10$2R./p9GTXF/325PJnx/tvOXaxMjJZhC/mEaAZJ3FRGQiZ4U6f/JBy" // admin-password hash

	wrongPasswords := []string{
		"wrong-password",
		"admin-password-wrong",
		"password",
		"",
	}

	for _, wrongPass := range wrongPasswords {
		if VerifyPasswordHash(hash, wrongPass) {
			t.Errorf("password verification should fail for %s", wrongPass)
		}
	}
}
