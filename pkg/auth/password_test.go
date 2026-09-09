package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword_ReturnsDifferentHashForSamePassword(t *testing.T) {
	hash1, err := HashPassword("mySecret123")
	require.NoError(t, err)

	hash2, err := HashPassword("mySecret123")
	require.NoError(t, err)

	// bcrypt подмешивает случайную соль на каждый вызов, поэтому хеши
	// одного и того же пароля не должны совпадать.
	assert.NotEqual(t, hash1, hash2)
}

func TestVerifyPassword_CorrectPassword_ReturnsTrue(t *testing.T) {
	hash, err := HashPassword("mySecret123")
	require.NoError(t, err)

	assert.True(t, VerifyPassword(hash, "mySecret123"))
}

func TestVerifyPassword_WrongPassword_ReturnsFalse(t *testing.T) {
	hash, err := HashPassword("mySecret123")
	require.NoError(t, err)

	assert.False(t, VerifyPassword(hash, "wrongPassword"))
}

func TestVerifyPassword_InvalidHash_ReturnsFalse(t *testing.T) {
	assert.False(t, VerifyPassword("not-a-valid-bcrypt-hash", "mySecret123"))
}
