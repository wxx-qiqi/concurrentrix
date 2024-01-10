package tools

import (
	"crypto/sha256"
	"encoding/hex"
)

// GetPhoneSha256 Coded as Base64
func GetPhoneSha256(strToHash string) string {
	// First SHA-256 hash
	sha256Hash := sha256.New()
	sha256Hash.Write([]byte(strToHash))
	hashedStr := sha256Hash.Sum(nil)
	hashedStrHex := hex.EncodeToString(hashedStr)

	// Second SHA-256 hash
	sha256Hash = sha256.New()
	sha256Hash.Write([]byte(hashedStrHex))
	digest := sha256Hash.Sum(nil)
	hexDigest := hex.EncodeToString(digest)

	return hexDigest
}

// GetPhoneSha256s Gain Authorization
func GetPhoneSha256s(strToHash string) string {
	sha256Hash := sha256.New()
	sha256Hash.Write([]byte(strToHash))
	hashedStr := sha256Hash.Sum(nil)
	hashedStrHex := hex.EncodeToString(hashedStr)
	return hashedStrHex
}
