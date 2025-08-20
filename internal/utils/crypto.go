package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// --- Password Hashing ---

// HashPassword generates a bcrypt hash of the password.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

// CheckPasswordHash compares a password with its bcrypt hash.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// --- Secure Token Generation ---

// GenerateSecureToken creates a cryptographically secure, URL-safe, random string.
// It's suitable for generating authorization codes, refresh tokens, client secrets, etc.
func GenerateSecureToken(length int) (string, error) {
	// The number of random bytes to generate.
	// Each byte becomes about 1.33 characters in base64, so we adjust.
	numBytes := length * 3 / 4
	if length%3 != 0 {
		numBytes++
	}

	b := make([]byte, numBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// --- PKCE (Proof Key for Code Exchange) ---

// GeneratePKCEVerifier creates a high-entropy cryptographic random string.
// Per RFC 7636, the verifier is a string of 43-128 characters. We'll default to a good length.
func GeneratePKCEVerifier() (string, error) {
	// 32 bytes of entropy is a common standard, resulting in a 43-character verifier.
	return GenerateSecureToken(43)
}

// GeneratePKCEChallengeS256 creates a code challenge from a verifier using SHA-256.
func GeneratePKCEChallengeS256(verifier string) string {
	h := sha256.New()
	h.Write([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
