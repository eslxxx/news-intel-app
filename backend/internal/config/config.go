package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	DBPath      string
	DataDir     string
	OpenAIKey   string
	OpenAIBase  string
	OpenAIModel string
}

var AppConfig *Config

func Load() *Config {
	godotenv.Load()

	AppConfig = &Config{
		Port:        getEnv("PORT", "5555"),
		DBPath:      getEnv("DB_PATH", "./data/news.db"),
		DataDir:     getEnv("DATA_DIR", "./data"),
		OpenAIKey:   getEnv("OPENAI_API_KEY", ""),
		OpenAIBase:  getEnv("OPENAI_BASE_URL", "https://api.openai.com/v1"),
		OpenAIModel: getEnv("OPENAI_MODEL", "gpt-4o-mini"),
	}

	return AppConfig
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
