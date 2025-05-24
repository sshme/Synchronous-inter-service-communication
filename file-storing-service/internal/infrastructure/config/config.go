package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the service
type Config struct {
	// Server config
	ServerPort string

	// Database config
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// S3 config
	S3Bucket         string
	S3Region         string
	S3Endpoint       string
	S3AccessKey      string
	S3SecretKey      string
	S3ForcePathStyle bool
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	err := godotenv.Load("./config/.env")
	if err != nil {
		return nil, err
	}

	config := &Config{
		// Server config
		ServerPort: getEnv("SERVER_PORT", "8000"),

		// Database config
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "files_db"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		// S3 config
		S3Bucket:         getEnv("S3_BUCKET", "files"),
		S3Region:         getEnv("S3_REGION", "us-east-1"),
		S3Endpoint:       getEnv("S3_ENDPOINT", ""),
		S3AccessKey:      getEnv("S3_ACCESS_KEY", ""),
		S3SecretKey:      getEnv("S3_SECRET_KEY", ""),
		S3ForcePathStyle: getBoolEnv("S3_FORCE_PATH_STYLE", true),
	}

	return config, nil
}

// Helper function to get environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// Helper function to get boolean environment variable with a fallback value
func getBoolEnv(key string, fallback bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		boolValue, err := strconv.ParseBool(value)
		if err == nil {
			return boolValue
		}
	}
	return fallback
}
