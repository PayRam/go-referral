package utils

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"testing"
)

func GenerateReferralCode() string {
	b := make([]byte, 8) // 8 bytes = 16 characters
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// AssertEqualNilable asserts equality for nilable values of any type.
func AssertEqualNilable[T comparable](t *testing.T, expected *T, actual *T, message string) {
	if expected == nil {
		assert.Nil(t, actual, message)
	} else {
		assert.NotNil(t, actual, message)
		assert.Equal(t, *expected, *actual, message)
	}
}

// AssertEqualIfExpectedNotNil asserts equality only if the expected value is not nil.
func AssertEqualIfExpectedNotNil[T comparable](t *testing.T, expected *T, actual T, message string) {
	if expected != nil {
		assert.Equal(t, *expected, actual, message)
	}
}

func StringPtr(s string) *string {
	return &s
}
