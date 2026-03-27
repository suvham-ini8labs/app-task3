package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	// Save current env to restore later
	envs := []string{"SERVER_PORT", "DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "READ_TIMEOUT", "WRITE_TIMEOUT", "IDLE_TIMEOUT", "LOG_LEVEL"}
	oldEnvs := make(map[string]string)
	for _, env := range envs {
		oldEnvs[env] = os.Getenv(env)
		_ = os.Unsetenv(env)
	}
	defer func() {
		for k, v := range oldEnvs {
			_ = os.Setenv(k, v)
		}
	}()

	t.Run("default_values", func(t *testing.T) {
		cfg, err := Load()
		assert.NoError(t, err)
		assert.Equal(t, "8080", cfg.ServerPort)
		assert.Equal(t, "localhost", cfg.DBHost)
		assert.Equal(t, "5432", cfg.DBPort)
		assert.Equal(t, "postgres", cfg.DBUser)
		assert.Equal(t, "postgres", cfg.DBPassword)
		assert.Equal(t, "postgres", cfg.DBName)
		assert.Equal(t, 15*time.Second, cfg.ReadTimeout)
		assert.Equal(t, 15*time.Second, cfg.WriteTimeout)
		assert.Equal(t, 60*time.Second, cfg.IdleTimeout)
		assert.Equal(t, "info", cfg.LogLevel)
	})

	t.Run("custom_values", func(t *testing.T) {
		_ = os.Setenv("SERVER_PORT", "9090")
		_ = os.Setenv("DB_HOST", "db")
		_ = os.Setenv("READ_TIMEOUT", "30s")
		
		cfg, err := Load()
		assert.NoError(t, err)
		assert.Equal(t, "9090", cfg.ServerPort)
		assert.Equal(t, "db", cfg.DBHost)
		assert.Equal(t, 30*time.Second, cfg.ReadTimeout)
	})

	t.Run("invalid_duration", func(t *testing.T) {
		_ = os.Setenv("READ_TIMEOUT", "invalid")
		cfg, err := Load()
		assert.NoError(t, err)
		assert.Equal(t, 15*time.Second, cfg.ReadTimeout) // Should fall back to default
	})
}

func TestGetDBConnectionString(t *testing.T) {
	cfg := &Config{
		DBHost:     "localhost",
		DBPort:     "5432",
		DBUser:     "user",
		DBPassword: "password",
		DBName:     "dbname",
	}
	expected := "host=localhost port=5432 user=user password=password dbname=dbname sslmode=disable"
	assert.Equal(t, expected, cfg.GetDBConnectionString())
}
