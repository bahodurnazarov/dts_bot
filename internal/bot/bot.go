package bot

import (
	handler "dts_bot/internal/handler"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"time"
)

// StartBot initializes and runs the Telegram bot.
func StartBot(botToken string) {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatalf("Error creating bot: %v", err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	updates := bot.GetUpdatesChan(updateConfig)

	// Start a goroutine to check the queue and notify the admin
	go func() {
		for {
			time.Sleep(2 * time.Second) // Check every 2 seconds
			//sendAdminQueueNotification(bot, false) // Removed to send only on new user.
		}
	}()

	for update := range updates {
		if update.Message != nil {
			go handler.HandleMessage(bot, update.Message)
		} else if update.CallbackQuery != nil {
			go handler.HandleCallbackQuery(bot, update.CallbackQuery)
		}
	}
}

//// StartBot initializes the bot
//func StartBot(botToken string) {
//	bot, err := tgbotapi.NewBotAPI(botToken)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	bot.Debug = true
//	log.Printf("Authorized on account %s", bot.Self.UserName)
//
//	// Set up update configuration
//	updateConfig := tgbotapi.NewUpdate(0)
//	updateConfig.Timeout = 60
//	updates := bot.GetUpdatesChan(updateConfig)
//
//	//go CheckServers(bot)
//
//	for update := range updates {
//		if update.Message != nil {
//			go handler.HandleMessage(bot, update.Message)
//		} else if update.CallbackQuery != nil {
//			go handler.HandleCallbackQuery(bot, update.CallbackQuery)
//		}
//	}
//}
