package handler

import (
	"dts_bot/pkg/db"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
	"strings"
)

func HandleCallbackQuery(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery) {
	chatID := callbackQuery.Message.Chat.ID
	data := callbackQuery.Data

	userLang, err := db.GetUserLanguage(chatID)
	if err != nil {
		log.Printf("Error fetching user language for chat %d: %v", chatID, err)
		userLang = "ru" // fallback to Russian if there's an error
	}

	lang := userLang
	// Ensure the selected language exists in the translations map
	if _, exists := Translations[lang]; !exists {
		log.Printf("Unsupported language '%s' for chat %d, defaulting to 'ru'", lang, chatID)
		lang = "ru" // Fallback to Russian if the language doesn't exist
	}

	switch {
	case data == "main_menu":
		sendMainMenu(bot, chatID, lang)
		bot.Send(tgbotapi.NewCallback(callbackQuery.ID, ""))

	case strings.HasPrefix(data, "connect_user_"):
		parts := strings.Split(data, "_")
		if len(parts) != 3 {
			return
		}
		userID, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			log.Printf("Invalid user ID in callback: %v", err)
			return
		}

		activeSessions.Store(chatID, userID)
		activeSessions.Store(userID, chatID)

		// Notify admin
		adminMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf(Translations[lang]["admin_connected"], userID))
		bot.Send(adminMsg)

		// Notify user
		userLang, _ := db.GetUserLanguage(userID)
		userMsg := tgbotapi.NewMessage(userID, Translations[userLang]["operator_connected"])
		bot.Send(userMsg)

		// Remove user from queue
		removeUserFromQueue(userID)

		bot.Send(tgbotapi.NewCallback(callbackQuery.ID, "✅ Подключено"))

	case strings.HasPrefix(data, "permit_answer_"):
		// Extract question ID from callback data
		parts := strings.Split(data, "_")
		if len(parts) < 3 {
			return
		}
		questionID, err := strconv.Atoi(parts[2])
		if err != nil {
			return
		}

		// Get the full answer from DB
		answer, err := db.GetAnswerByID(questionID, lang)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, Translations[lang]["error_fetching_answer"])
			bot.Send(msg)
			return
		}

		// Send the answer to the user
		msg := tgbotapi.NewMessage(chatID, answer)
		bot.Send(msg)

		// Acknowledge the callback
		bot.Send(tgbotapi.NewCallback(callbackQuery.ID, ""))

	case data == ContactSupportCallback:
		processContactSupportRequest(bot, callbackQuery, lang)
	default:
		// Unknown callback - send default response
		msg := tgbotapi.NewMessage(chatID, Translations[lang]["unknown_command"])
		bot.Send(msg)
		bot.Send(tgbotapi.NewCallback(callbackQuery.ID, ""))
	}
}

// processContactSupportRequest handles the "contact_support" callback.
func processContactSupportRequest(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, lang string) {
	userID := query.Message.Chat.ID

	// Acknowledge the button press
	callbackResponse := tgbotapi.NewCallback(query.ID, Translations[lang]["request_sent_to_operator"])
	if _, err := bot.Send(callbackResponse); err != nil {
		log.Printf("Error sending callback response: %v", err)
	}

	// Notify the user
	notificationMsg := tgbotapi.NewMessage(userID, Translations[lang]["waiting_for_operator"])
	notificationMsg.ReplyMarkup = &tgbotapi.ReplyKeyboardMarkup{
		Keyboard: [][]tgbotapi.KeyboardButton{
			{
				tgbotapi.NewKeyboardButton(Translations[lang]["close"]), // Translated "Close" button
				tgbotapi.NewKeyboardButton(Translations[lang]["menu"]),
			},
		},
		ResizeKeyboard: true,
	}

	if _, err := bot.Send(notificationMsg); err != nil {
		log.Printf("Error sending user notification: %v", err)
	}

	// Add the user to the queue
	addUserToQueue(userID)
	sendAdminQueueNotification(bot, true, lang) // Notify admin about new user
}

func sendAdminQueueNotification(bot *tgbotapi.BotAPI, newUserAdded bool, lang string) {
	currentNotification := formatQueueNotification(lang)
	lastQueueNotificationMutex.Lock()
	defer lastQueueNotificationMutex.Unlock()

	if currentNotification != lastQueueNotification && newUserAdded { // send message only when new user added to queue
		msg := tgbotapi.NewMessage(AdminUserID, fmt.Sprintf("🔔 %s\n%s", Translations[lang]["new_request_in_queue"], currentNotification))
		msg.ReplyMarkup = &tgbotapi.ReplyKeyboardMarkup{
			Keyboard: [][]tgbotapi.KeyboardButton{
				{
					tgbotapi.NewKeyboardButton(Translations[lang]["connect"]), // Translated "Connect" button
					//tgbotapi.NewKeyboardButton(Translations[lang]["close"]),   // Translated "Close" button

				},
				{
					tgbotapi.NewKeyboardButton(Translations[lang]["close"]), // Translated "Close" button
				},
			},
			ResizeKeyboard: true,
		}
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Error sending queue notification: %v", err)
		}
		lastQueueNotification = currentNotification
	}
}

func formatQueueNotification(lang string) string {
	queueList := getCurrentQueue()
	if len(queueList) == 0 {
		return Translations[lang]["queue_empty"] // Translate "Queue is empty"
	}
	queueString := Translations[lang]["current_queue"] + ":\n" // Translate "Current queue"
	for i, userID := range queueList {
		queueString += fmt.Sprintf("%d: %d\n", i+1, userID)
	}
	return queueString
}

// Function to get the current queue
func getCurrentQueue() []int64 {
	queueMutex.Lock()
	defer queueMutex.Unlock()
	queueCopy := make([]int64, len(userQueue))
	copy(queueCopy, userQueue)
	return queueCopy
}

func addUserToQueue(userID int64) {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	// Check if user already in queue
	for _, id := range userQueue {
		if id == userID {
			log.Printf("User %d is already in the queue.", userID)
			return
		}
	}

	// Add user if not already in queue
	userQueue = append(userQueue, userID)
	log.Printf("User %d added to queue. Current queue: %v", userID, userQueue)
}
