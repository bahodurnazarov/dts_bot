package handler

import (
	"dts_bot/pkg/db"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
	"strings"
)

func handleSelectedQuestion(bot *tgbotapi.BotAPI, message *tgbotapi.Message, lang string) {
	chatID := message.Chat.ID

	// Extract question number
	dotIndex := strings.Index(message.Text, ".")
	if dotIndex == -1 {
		return
	}
	questionNum, err := strconv.Atoi(strings.TrimSpace(message.Text[:dotIndex]))
	if err != nil {
		return
	}

	// Get current menu type from session
	currentMenu := getCurrentMenu(chatID)
	if currentMenu == "" {
		msg := tgbotapi.NewMessage(chatID, Translations[lang]["menu_session_expired"])
		bot.Send(msg)
		sendMainMenu(bot, chatID, lang)
		return
	}

	// Get questions based on current menu
	var questions []db.Question
	switch currentMenu {
	case "permit":
		questions, err = db.GetPermitMenuQuestions(lang)
	case "borhat":
		questions, err = db.GetBorhatQuestions(lang)
	case "ijozatnoma":
		questions, err = db.GetIjozatnomaQuestions(lang)
	case "roxkhat":
		questions, err = db.GetRoxkhatQuestions(lang)
	case "certificate":
		questions, err = db.GetCertificateQuestions(lang)
	default:
		msg := tgbotapi.NewMessage(chatID, Translations[lang]["invalid_menu"])
		bot.Send(msg)
		return
	}
	log.Println("QUESTION :", questions)
	if err != nil || len(questions) == 0 {
		msg := tgbotapi.NewMessage(chatID, Translations[lang]["error_fetching_questions"])
		bot.Send(msg)
		return
	}

	// Validate question number
	if questionNum < 1 || questionNum > len(questions) {
		msg := tgbotapi.NewMessage(chatID, Translations[lang]["invalid_question"])
		bot.Send(msg)
		return
	}

	// Get and send the answer
	answer, err := db.GetAnswerByID(questions[questionNum-1].ID, lang)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, Translations[lang]["error_fetching_answer"])
		bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(chatID, answer)
	bot.Send(msg)

	// Show the same menu again
	switch currentMenu {
	case "permit":
		ShowPermitMenu(bot, chatID, lang)
	case "borhat":
		ShowBorhatMenu(bot, chatID, lang)
	case "ijozatnoma":
		ShowIjozatnomaMenu(bot, chatID, lang)
	case "roxkhat":
		ShowRoxkhatMenu(bot, chatID, lang)
	case "certificate":
		ShowCertificateMenu(bot, chatID, lang)
	}
}

func isQuestionSelection(text string) bool {
	// Find the first dot in the text
	dotIndex := strings.Index(text, ".")
	if dotIndex == -1 {
		return false
	}

	// Extract the number part before the dot
	numberPart := strings.TrimSpace(text[:dotIndex])

	// Try to convert to integer
	_, err := strconv.Atoi(numberPart)
	return err == nil
}

// Sets the user's language and updates the database
func setUserLanguage(bot *tgbotapi.BotAPI, message *tgbotapi.Message, lang string) {
	chatID := message.Chat.ID
	firstName := message.Chat.FirstName
	lastName := message.Chat.LastName
	userName := message.Chat.UserName
	if err := db.SetUserLanguageAndInfo(chatID, lang, firstName, lastName, userName); err != nil {
		log.Printf("Error setting language for chat %d: %v", chatID, err)
	}
	msg := tgbotapi.NewMessage(chatID, Translations[lang]["lang_set"])
	bot.Send(msg)
}
