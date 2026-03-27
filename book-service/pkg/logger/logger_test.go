package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("json_format", func(t *testing.T) {
		l, err := New("info", "json")
		assert.NoError(t, err)
		assert.NotNil(t, l)
		l.Sync()
	})

	t.Run("development_format", func(t *testing.T) {
		l, err := New("debug", "console")
		assert.NoError(t, err)
		assert.NotNil(t, l)
		l.Sync()
	})

	t.Run("invalid_level", func(t *testing.T) {
		l, err := New("invalid", "json")
		assert.NoError(t, err) // Should default to info
		assert.NotNil(t, l)
	})
}
