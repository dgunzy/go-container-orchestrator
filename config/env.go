package config

import (
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() (func(string) string, error) {
	err := godotenv.Load("../../.env")
	if err != nil {
		return nil, err
	}

	return os.Getenv, nil
}

// GetEnvOrDefault retrieves an environment variable or returns a default value if not set
func GetEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
