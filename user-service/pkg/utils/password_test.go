package utils

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestPasswordUtils(t *testing.T) {
	password := "my-secret-password"

	t.Run("HashAndCheck", func(t *testing.T) {
		hash, err := HashPassword(password)
		assert.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.NotEqual(t, password, hash)

		assert.True(t, CheckPasswordHash(password, hash))
		assert.False(t, CheckPasswordHash("wrong-password", hash))
	})

	t.Run("EmptyPassword", func(t *testing.T) {
		hash, err := HashPassword("")
		assert.NoError(t, err)
		assert.True(t, CheckPasswordHash("", hash))
	})
}
