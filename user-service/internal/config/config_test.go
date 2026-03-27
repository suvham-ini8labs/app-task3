package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfigLoad(t *testing.T) {
	t.Run("DefaultValues", func(t *testing.T) {
		// Clean env before testing
		os.Clearenv()
		
		cfg, err := Load()
		assert.NoError(t, err)
		assert.Equal(t, "8081", cfg.ServerPort)
		assert.Equal(t, "localhost", cfg.DBHost)
	})

	t.Run("EnvironmentValues", func(t *testing.T) {
		os.Setenv("SERVER_PORT", "9090")
		os.Setenv("DB_HOST", "db-service")
		os.Setenv("JWT_EXPIRATION", "2h")
		defer os.Clearenv()

		cfg, err := Load()
		assert.NoError(t, err)
		assert.Equal(t, "9090", cfg.ServerPort)
		assert.Equal(t, "db-service", cfg.DBHost)
		assert.Equal(t, 2*time.Hour, cfg.JWTExpiration)
	})

	t.Run("InvalidDuration", func(t *testing.T) {
		os.Setenv("JWT_EXPIRATION", "invalid")
		defer os.Unsetenv("JWT_EXPIRATION")

		cfg, err := Load()
		assert.NoError(t, err)
		// Should fall back to default (24h)
		assert.Equal(t, 24*time.Hour, cfg.JWTExpiration)
	})

	t.Run("GetDBConnectionString", func(t *testing.T) {
		cfg := &Config{
			DBHost: "h", DBPort: "p", DBUser: "u", DBPassword: "pw", DBName: "d",
		}
		expected := "host=h port=p user=u password=pw dbname=d sslmode=disable"
		assert.Equal(t, expected, cfg.GetDBConnectionString())
	})
}
