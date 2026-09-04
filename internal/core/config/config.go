package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port       string
	EnvFile    string
	DBName     string
	SessionKey string
}

func Load() (Config, error) {
	envFile := os.Getenv("ENV_FILE")
	if envFile == "" {
		envFile = ".env"
	}
	if err := godotenv.Load(envFile); err != nil {
		slog.Warn("env file not found, using environment variables", "file", envFile)
	}

	sessionKey := os.Getenv("SESSION_KEY")
	if sessionKey == "" {
		return Config{}, fmt.Errorf("SESSION_KEY is not set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "./db/url-shortener.db"
	}

	return Config{
		Port:       port,
		EnvFile:    envFile,
		DBName:     dbName,
		SessionKey: sessionKey,
	}, nil
}
