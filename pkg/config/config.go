package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config structure
type Config struct {
	BotToken     string
	OpenAIAPIKey string
	DbURL        string
}

// LoadConfig loads environment variables
func LoadConfig() Config {
	// Load .env file if available
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	return Config{
		BotToken:     os.Getenv("TELEGRAM_BOT_TOKEN"),
		OpenAIAPIKey: os.Getenv("OPENAI_API_KEY"),
		DbURL:        os.Getenv("DATABASE_URL"),
	}
}
