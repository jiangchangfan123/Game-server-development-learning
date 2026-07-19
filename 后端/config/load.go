package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

var GitHub *GitHubAuthConfig
var DB *DBConfig

func Load() error {
	if err := godotenv.Load("../.env"); err != nil {
		return fmt.Errorf("load .env: %w", err)
	}

	GitHub = &GitHubAuthConfig{
		ClientId:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		RedirectURL:  "http://localhost:8080/api/github/callback",
		AuthURL:      "https://github.com/login/oauth/authorize",
		TokenURL:     "https://github.com/login/oauth/access_token",
	}

	if GitHub.ClientId == "" || GitHub.ClientSecret == "" {
		return fmt.Errorf("CLIENT_ID and CLIENT_SECRET must be set in .env")
	}

	DB = &DBConfig{
		Host:     envOrDefault("DB_HOST", "127.0.0.1"),
		Port:     envOrDefault("DB_PORT", "3306"),
		Name:     envOrDefault("DB_NAME", "game"),
		User:     envOrDefault("DB_USER", "root"),
		Password: trimEnv("DB_PASSWORD"),
	}

	if DB.Name == "" {
		return fmt.Errorf("DB_NAME must be set in .env")
	}

	if proxy := trimEnv("HTTP_PROXY"); proxy != "" {
		os.Setenv("HTTP_PROXY", proxy)
		os.Setenv("HTTPS_PROXY", proxy)
	}

	return nil
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func trimEnv(key string) string {
	v := strings.TrimSpace(os.Getenv(key))
	return strings.Trim(v, `"`)
}
