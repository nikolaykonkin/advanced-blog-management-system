package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateToken_ReturnsValidTokenAndExpiry(t *testing.T) {
	before := time.Now()

	token, expiresAt, err := GenerateToken(1, "user@test.com", "user1", "test-secret")

	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.WithinDuration(t, before.Add(24*time.Hour), expiresAt, time.Second)
}

func TestValidateToken_ValidToken_ReturnsClaims(t *testing.T) {
	token, _, err := GenerateToken(42, "user@test.com", "user1", "test-secret")
	require.NoError(t, err)

	claims, err := ValidateToken(token, "test-secret")

	require.NoError(t, err)
	assert.Equal(t, 42, claims.UserID)
	assert.Equal(t, "user@test.com", claims.Email)
	assert.Equal(t, "user1", claims.Username)
}

func TestValidateToken_ExpiredToken_ReturnsError(t *testing.T) {
	// GenerateToken всегда ставит срок действия +24ч от текущего момента,
	// поэтому просроченный токен собираем вручную, минуя GenerateToken.
	claims := Claims{
		UserID:   1,
		Email:    "user@test.com",
		Username: "user1",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("test-secret"))
	require.NoError(t, err)

	_, err = ValidateToken(signed, "test-secret")

	assert.Error(t, err)
}

func TestValidateToken_WrongSecret_ReturnsError(t *testing.T) {
	token, _, err := GenerateToken(1, "user@test.com", "user1", "correct-secret")
	require.NoError(t, err)

	_, err = ValidateToken(token, "wrong-secret")

	assert.Error(t, err)
}

func TestValidateToken_MalformedToken_ReturnsError(t *testing.T) {
	_, err := ValidateToken("this-is-not-a-jwt", "test-secret")

	assert.Error(t, err)
}
