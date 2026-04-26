package config

import (
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL     string
	GeminiAPIKey    string
	Port            string
	AppEnv          string
	ScrapeIntervalH int
}

func Load() Config {
	scrapeInterval, _ := strconv.Atoi(getEnvOrDefault("SCRAPE_INTERVAL_HOURS", "24"))
	return Config{
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		GeminiAPIKey:    os.Getenv("GEMINI_API_KEY"),
		Port:            getEnvOrDefault("PORT", "8080"),
		AppEnv:          getEnvOrDefault("APP_ENV", "development"),
		ScrapeIntervalH: scrapeInterval,
	}
}

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
