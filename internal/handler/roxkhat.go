package handler

import (
	"dts_bot/pkg/db"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func ShowRoxkhatMenu(bot *tgbotapi.BotAPI, chatID int64, lang string) {
	SetCurrentMenu(chatID, "roxkhat")
	questions, err := db.GetRoxkhatQuestions(lang)
	if err != nil || len(questions) == 0 {
		msg := tgbotapi.NewMessage(chatID, Translations[lang]["error_fetching_questions"])
		bot.Send(msg)
		return
	}

	// Create a custom reply keyboard with one question per line
	var keyboardRows [][]tgbotapi.KeyboardButton

	for i, q := range questions {
		// Use the full question text for each button
		btnText := fmt.Sprintf("%d. %s", i+1, q.Question)

		// Each question gets its own row
		keyboardRows = append(keyboardRows, []tgbotapi.KeyboardButton{
			tgbotapi.NewKeyboardButton(btnText),
		})
	}

	// Add back button on its own row
	keyboardRows = append(keyboardRows, []tgbotapi.KeyboardButton{
		tgbotapi.NewKeyboardButton(Translations[lang]["back"]),
	})

	// Create reply markup
	replyMarkup := tgbotapi.NewReplyKeyboard(keyboardRows...)
	replyMarkup.OneTimeKeyboard = true // Keyboard will hide after selection
	replyMarkup.ResizeKeyboard = true  // Auto-resize to fit screen

	// Send message with question list
	msg := tgbotapi.NewMessage(chatID, Translations[lang]["permit_menu"])
	msg.ReplyMarkup = replyMarkup
	bot.Send(msg)
}
