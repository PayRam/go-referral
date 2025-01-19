package utils

import (
	"crypto/rand"
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

const charset = "BF7CDXR0E3ZHPI1JK9L4N2OAQSFT5UVMW6Y8"

// CreateReferralCode generates a secure random referral code of the specified length.
func CreateReferralCode(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("length must be greater than zero")
	}

	code := make([]byte, length)
	for i := range code {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("failed to generate referral code: %w", err)
		}
		code[i] = charset[randomIndex.Int64()]
	}
	return string(code), nil
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
