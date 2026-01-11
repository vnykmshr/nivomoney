package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// Cache key prefixes
const (
	PrefixSession = "session:"
	PrefixUser    = "user:"
	PrefixToken   = "token:"
)

// Default TTLs
const (
	SessionTTL     = 24 * time.Hour // Match JWT expiry
	UserProfileTTL = 15 * time.Minute
	TokenTTL       = 24 * time.Hour
)

// SessionKey generates a cache key for user sessions.
// Format: session:{user_id}:{token_hash_prefix}
func SessionKey(userID, tokenHash string) string {
	// Use first 16 chars of token hash for key brevity
	prefix := tokenHash
	if len(tokenHash) > 16 {
		prefix = tokenHash[:16]
	}
	return fmt.Sprintf("%s%s:%s", PrefixSession, userID, prefix)
}

// UserKey generates a cache key for user profiles.
// Format: user:{user_id}
func UserKey(userID string) string {
	return fmt.Sprintf("%s%s", PrefixUser, userID)
}

// TokenKey generates a cache key for token validation.
// Format: token:{token_hash}
func TokenKey(token string) string {
	// Hash the full token for the key
	hash := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%s%s", PrefixToken, hex.EncodeToString(hash[:]))
}

// HashToken creates a SHA-256 hash of a token string.
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
