package main

import (
	"dts_bot/internal/bot"
	"dts_bot/pkg/config"
	"dts_bot/pkg/db"
	"log"
)

func main() {
	// Load config
	cfg := config.LoadConfig()

	// Connect to database
	db.ConnectDatabase(cfg.DbURL)
	defer db.CloseDatabase()

	// Start the bot
	log.Println("Starting bot...")
	bot.StartBot(cfg.BotToken)
}
