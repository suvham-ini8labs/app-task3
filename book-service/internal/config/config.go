package config

import (
	"os"
	"time"
	"fmt"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort     string
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
	LogLevel       string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		DBHost:        getEnv("DB_HOST", "localhost"),
		DBPort:        getEnv("DB_PORT", "5432"),
		DBUser:        getEnv("DB_USER", "postgres"),
		DBPassword:    getEnv("DB_PASSWORD", "postgres"),
		DBName:        getEnv("DB_NAME", "postgres"),
		ReadTimeout:    getDurationEnv("READ_TIMEOUT", 15*time.Second),
		WriteTimeout:   getDurationEnv("WRITE_TIMEOUT", 15*time.Second),
		IdleTimeout:    getDurationEnv("IDLE_TIMEOUT", 60*time.Second),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
	}

	return cfg, nil
}

func (c *Config) GetDBConnectionString() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
