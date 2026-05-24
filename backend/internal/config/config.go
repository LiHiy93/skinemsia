package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL    string
	BotToken       string
	AppSecret      string
	Port           string
	AllowedOrigins string
	WebAppURL      string
	AppEnv         string
}

func Load() *Config {
	_ = godotenv.Load()

	cfg := &Config{
		DatabaseURL:    getEnv("DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/skinemsia?sslmode=disable"),
		BotToken:       getEnv("BOT_TOKEN", ""),
		AppSecret:      getEnv("APP_SECRET", "dev_secret"),
		Port:           getEnv("PORT", "8080"),
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "http://localhost:5173"),
		WebAppURL:      getEnv("WEB_APP_URL", "http://localhost:5173"),
		AppEnv:         getEnv("APP_ENV", "development"),
	}

	if cfg.BotToken == "" {
		log.Println("WARNING: BOT_TOKEN is not set — bot will not start")
	}

	return cfg
}

func (c *Config) IsDev() bool {
	return c.AppEnv == "development"
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
