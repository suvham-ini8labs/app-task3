package logger

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestLoggerNew(t *testing.T) {
	t.Run("JSONFormat", func(t *testing.T) {
		l, err := New("info", "json")
		assert.NoError(t, err)
		assert.NotNil(t, l)
		l.Info("testing json logger")
	})

	t.Run("ConsoleFormat", func(t *testing.T) {
		l, err := New("debug", "console")
		assert.NoError(t, err)
		assert.NotNil(t, l)
		l.Debug("testing console logger")
	})

	t.Run("InvalidLevel", func(t *testing.T) {
		l, err := New("invalid", "json")
		assert.NoError(t, err) // Should fallback to InfoLevel
		assert.NotNil(t, l)
	})
}
