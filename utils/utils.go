package utils

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateReferralCode() string {
	b := make([]byte, 8) // 8 bytes = 16 characters
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func StringPtr(s string) *string {
	return &s
}
