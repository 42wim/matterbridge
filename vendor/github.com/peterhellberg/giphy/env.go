package giphy

import (
	"os"
	"strconv"
)

// Env returns a string from the ENV, or fallback variable
func Env(key, fallback string) string {
	v := os.Getenv(key)
	if v != "" {
		return v
	}

	return fallback
}

// EnvBool returns a bool from the ENV, or fallback variable
func EnvBool(key string, fallback bool) bool {
	if b, err := strconv.ParseBool(os.Getenv(key)); err == nil {
		return b
	}

	return fallback
}

// EnvInt returns an int from the ENV, or fallback variable
func EnvInt(key string, fallback int) int {
	if i, err := strconv.Atoi(os.Getenv(key)); err == nil {
		return i
	}

	return fallback
}
