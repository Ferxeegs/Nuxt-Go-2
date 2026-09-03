// Package hash wraps bcrypt for password hashing/verification, plus a fast
// SHA-256 digest for opaque tokens (refresh tokens) that need exact-match
// lookup rather than bcrypt's slow, salted comparison.
package hash

import (
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

// HashToken returns a hex-encoded SHA-256 digest, used to store refresh
// tokens at rest without keeping the raw secret in the database.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func Hash(password string, cost int) (string, error) {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}
	b, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func Verify(password, hashed string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password)) == nil
}
