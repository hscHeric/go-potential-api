package hash

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	t.Run("should hash password successfully", func(t *testing.T) {
		password := "senhaSegura123"

		hashed, err := HashPassword(password)

		require.NoError(t, err)
		assert.NotEmpty(t, hashed)
		assert.NotEqual(t, password, hashed)
	})

	t.Run("should accept empty password", func(t *testing.T) {
		hashed, err := HashPassword("")

		require.NoError(t, err)
		assert.NotEmpty(t, hashed)
	})

	t.Run("should generate different hashes for same password", func(t *testing.T) {
		password := "samePassword"

		hash1, err1 := HashPassword(password)
		hash2, err2 := HashPassword(password)

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.NotEqual(t, hash1, hash2, "bcrypt should use random salt")
	})
}

func TestCheckPassword(t *testing.T) {
	validPassword := "secret123"
	validHash, _ := HashPassword(validPassword)

	t.Run("should return true for correct password", func(t *testing.T) {
		result := CheckPassword(validPassword, validHash)

		assert.True(t, result)
	})

	t.Run("should return false for incorrect password", func(t *testing.T) {
		result := CheckPassword("wrongPassword", validHash)

		assert.False(t, result)
	})

	t.Run("should return false for invalid hash", func(t *testing.T) {
		result := CheckPassword("anything", "invalid_hash")

		assert.False(t, result)
	})

	t.Run("should return false for empty hash", func(t *testing.T) {
		result := CheckPassword(validPassword, "")

		assert.False(t, result)
	})
}
