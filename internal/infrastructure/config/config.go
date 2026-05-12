package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort string
	AppEnv  string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	JWTSecret string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		AppPort:    getOrDefault("APP_PORT", "8080"),
		AppEnv:     getOrDefault("APP_ENV", "local"),
		DBHost:     getOrDefault("DB_HOST", "localhost"),
		DBPort:     getOrDefault("DB_PORT", "5432"),
		DBUser:     getOrDefault("DB_USER", "postgres"),
		DBPassword: getOrDefault("DB_PASSWORD", "postgres"),
		DBName:     getOrDefault("DB_NAME", "room_booking"),
		DBSSLMode:  getOrDefault("DB_SSLMODE", "disable"),
		JWTSecret:  getOrDefault("JWT_SECRET", "super-secret-key"),
	}

	if cfg.DBHost == "" ||
		cfg.DBPort == "" ||
		cfg.DBUser == "" ||
		cfg.DBPassword == "" ||
		cfg.DBName == "" {
		return nil, fmt.Errorf("database config is not fully set")
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("jwt secret is not set")
	}

	return cfg, nil
}

func (c *Config) DatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
		c.DBSSLMode,
	)
}

func getOrDefault(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}

	return val
}