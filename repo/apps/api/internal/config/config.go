package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	Port                string
	DatabaseURL         string
	JWTSecret           string
	MasterEncryptionKey string
	FileVaultPath       string
	DownloadTokenTTL    time.Duration
	LogLevel            string
}

func Load() (*Config, error) {
	port := getEnv("PORT", "8080")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbHost := getEnv("DB_HOST", "localhost")
		dbPort := getEnv("DB_PORT", "5432")
		dbUser := getEnv("DB_USER", "travel")
		dbPass := os.Getenv("DB_PASSWORD")
		dbName := getEnv("DB_NAME", "travel_platform")
		dbSSL := getEnv("DB_SSLMODE", "disable")
		if dbPass == "" {
			return nil, fmt.Errorf("DB_PASSWORD or DATABASE_URL is required")
		}
		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			dbUser, dbPass, dbHost, dbPort, dbName, dbSSL)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	masterKey := os.Getenv("MASTER_ENCRYPTION_KEY")
	if masterKey == "" {
		return nil, fmt.Errorf("MASTER_ENCRYPTION_KEY is required")
	}

	ttlStr := getEnv("DOWNLOAD_TOKEN_TTL", "15m")
	ttl, err := time.ParseDuration(ttlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid DOWNLOAD_TOKEN_TTL: %w", err)
	}

	return &Config{
		Port:                port,
		DatabaseURL:         dbURL,
		JWTSecret:           jwtSecret,
		MasterEncryptionKey: masterKey,
		FileVaultPath:       getEnv("FILE_VAULT_PATH", "./vault"),
		DownloadTokenTTL:    ttl,
		LogLevel:            getEnv("LOG_LEVEL", "info"),
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
