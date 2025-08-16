package config

import (
	"os"
	"strconv"
)

// Settings holds all environment variables used throughout the application
type Settings struct {
	// Database Configuration
	DatabaseURL string

	// Server Configuration
	Port    string
	BaseURL string

	// JWT Configuration
	JWTSecret    string
	JWTExpiryHours int

	// Signature Configuration
	SignatureSecret string

	// Storage Configuration
	StoragePath string
	MaxStorage  int64

	// System Configuration
	SystemName string
	Debug      bool
}

// NewSettings loads configuration from environment variables
func NewSettings() *Settings {
	settings := &Settings{
		// Database
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/shbucket?sslmode=disable"),

		// Server
		Port:    getEnv("PORT", "8080"),
		BaseURL: getEnv("BASE_URL", ""),

		// JWT
		JWTSecret:      getEnv("JWT_SECRET", "your-jwt-secret-change-in-production"),
		JWTExpiryHours: getEnvAsInt("JWT_EXPIRY_HOURS", 24),

		// Signature
		SignatureSecret: getEnv("SIGNATURE_SECRET", "your-signature-secret-change-in-production"),

		// Storage
		StoragePath: getEnv("STORAGE_PATH", "./storage"),
		MaxStorage:  getEnvAsInt64("MAX_STORAGE", 10*1024*1024*1024), // 10GB default

		// System
		SystemName: getEnv("SYSTEM_NAME", "SHBucket"),
		Debug:      getEnvAsBool("DEBUG", false),
	}

	// Set default BaseURL if not provided
	if settings.BaseURL == "" {
		settings.BaseURL = "http://localhost:" + settings.Port
	}

	return settings
}

// getEnv gets environment variable with fallback
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets environment variable as integer with fallback
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsInt64 gets environment variable as int64 with fallback
func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsBool gets environment variable as boolean with fallback
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// GetSettings returns a singleton instance of settings
var globalSettings *Settings

func GetSettings() *Settings {
	if globalSettings == nil {
		globalSettings = NewSettings()
	}
	return globalSettings
}