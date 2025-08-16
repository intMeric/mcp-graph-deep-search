package env

import (
	"os"
)

func GetOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func GetHostName() string {
	h, err := os.Hostname()
	if err != nil {
		return ""
	}
	return h
}
