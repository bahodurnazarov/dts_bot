package handler

import (
	"dts_bot/pkg/db"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"sync"
)

// Constants
const (
	ContactSupportCallback = "contact_support"
)

// AdminUserID should ideally be loaded from an environment variable or config.
const AdminUserID int64 = 6302455898 // ID администратора, измените на ID вашего администратора

// Active support sessions: map of user ID to admin's chat ID (or vice-versa)
var activeSessions = sync.Map{}

// User queue for support requests.  Using a mutex for thread safety.
var userQueue []int64
var queueMutex sync.Mutex

// Last queue notification sent to admin
var lastQueueNotification string
var lastQueueNotificationMutex sync.Mutex

func HandleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	chatID := message.Chat.ID

	// Fetch user language from DB
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
	// Check for the /close command first, regardless of the session status
	if message.Text == Translations[lang]["close_button"] {
		handleCloseCommand(bot, message, lang)
		return
	}

	if otherChatIDVal, ok := activeSessions.Load(chatID); ok {
		otherChatID := otherChatIDVal.(int64)
		var senderName string
		if chatID == AdminUserID {
			senderName = "Оператор"
		} else {
			senderName = "Клиент"
		}
		forwardedMessage := tgbotapi.NewMessage(otherChatID, fmt.Sprintf("👤 %s: %s", senderName, message.Text))
		if _, err := bot.Send(forwardedMessage); err != nil {
			log.Printf("Error sending message in session: %v", err)
		}
		return
	}

	// First check if this is a question selection
	if isQuestionSelection(message.Text) {
		handleSelectedQuestion(bot, message, lang)
		return
	}

	switch message.Text {
	case Translations[lang]["support_button"]:
		SendSupportMenu(bot, chatID, lang)
	case "/connect", "Подключиться", "Пайваст шудан": // Admin uses /connect without user ID
		handleConnectCommand(bot, message, lang)
	case Translations[lang]["view_queue"]:
		queueMsg := formatQueueNotification(lang)
		bot.Send(tgbotapi.NewMessage(chatID, queueMsg))

	case "/start":
		sendLanguageSelection(bot, chatID)
	case "Русский 🇷🇺":
		setUserLanguage(bot, message, "ru")
		sendMainMenu(bot, chatID, "ru")
	case Translations[lang]["back"]:
		sendMainMenu(bot, chatID, lang)
	case "Тоҷикӣ 🇹🇯":
		setUserLanguage(bot, message, "tg")
		sendMainMenu(bot, chatID, "tg")
		// case translations[lang]["support_button"]:
		// 	sendContactInfo(bot, chatID, lang)

	case "Меню":
		sendMainMenu(bot, chatID, lang)
	case Translations[lang]["permit_menu_button"]:
		ShowPermitMenu(bot, chatID, lang)
	case Translations[lang]["borhat_button"]:
		ShowBorhatMenu(bot, chatID, lang)
	case Translations[lang]["ijozatnoma_button"]:
		ShowIjozatnomaMenu(bot, chatID, lang)
	case Translations[lang]["roxkhat_button"]:
		ShowRoxkhatMenu(bot, chatID, lang)
	case Translations[lang]["certificate_button"]:
		ShowCertificateMenu(bot, chatID, lang)
	default:
		if !isQuestionSelection(message.Text) {
			msg := tgbotapi.NewMessage(chatID, Translations[lang]["welcome"])
			bot.Send(msg)
		}
	}

}

func SendSupportMenu(bot *tgbotapi.BotAPI, chatID int64, lang string) {
	contactButton := tgbotapi.NewInlineKeyboardButtonData(Translations[lang]["support_button"], ContactSupportCallback)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(contactButton))

	menuMsg := tgbotapi.NewMessage(chatID, Translations[lang]["support_text"])
	menuMsg.ReplyMarkup = keyboard
	if _, err := bot.Send(menuMsg); err != nil {
		log.Printf("Error sending support menu: %v", err)
	}
}

func handleViewQueueCommand(bot *tgbotapi.BotAPI, chatID int64, lang string) {
	if chatID != AdminUserID {
		msg := tgbotapi.NewMessage(chatID, "⛔ Только оператор может просматривать очередь.")
		bot.Send(msg)
		return
	}

	queue := getCurrentQueue()
	if len(queue) == 0 {
		msg := tgbotapi.NewMessage(chatID, Translations[lang]["queue_empty"])
		bot.Send(msg)
		return
	}

	// Create inline buttons for each user in queue
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, userID := range queue {
		btn := tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("👤 %d", userID), fmt.Sprintf("connect_user_%d", userID))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	msg := tgbotapi.NewMessage(chatID, Translations[lang]["current_queue"])
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	bot.Send(msg)
}

func handleConnectCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, lang string) {
	if msg.Chat.ID != AdminUserID {
		reply := tgbotapi.NewMessage(msg.Chat.ID, Translations[lang]["admin_only"])
		bot.Send(reply)
		return
	}

	userID, ok := getNextUserFromQueue()
	if !ok {
		reply := tgbotapi.NewMessage(msg.Chat.ID, Translations[lang]["queue_empty"])
		reply.ReplyMarkup = &tgbotapi.ReplyKeyboardMarkup{
			Keyboard: [][]tgbotapi.KeyboardButton{
				{tgbotapi.NewKeyboardButton(Translations[lang]["connect"])},
				{tgbotapi.NewKeyboardButton(Translations[lang]["close"])},
			},
			ResizeKeyboard: true,
		}
		bot.Send(reply)
		return
	}

	// Start the session
	activeSessions.Store(userID, msg.Chat.ID)
	activeSessions.Store(msg.Chat.ID, userID)

	adminNotification := tgbotapi.NewMessage(msg.Chat.ID,
		fmt.Sprintf("🤝 Вы подключились к пользователю %d. Теперь ваши сообщения будут отправляться напрямую.", userID))
	adminNotification.ReplyMarkup = &tgbotapi.ReplyKeyboardMarkup{
		Keyboard: [][]tgbotapi.KeyboardButton{
			{
				tgbotapi.NewKeyboardButton(Translations[lang]["close"]),
				//tgbotapi.NewKeyboardButton(Translations[lang]["connect"]),
			}, // Translated "Connect" button

		},
		ResizeKeyboard: true,
	}

	userNotification := tgbotapi.NewMessage(userID, Translations[lang]["operator_connected"])
	userNotification.ReplyMarkup = &tgbotapi.ReplyKeyboardMarkup{
		Keyboard: [][]tgbotapi.KeyboardButton{
			{
				tgbotapi.NewKeyboardButton(Translations[lang]["close"]),
				//tgbotapi.NewKeyboardButton(Translations[lang]["connect"]),
			}, // Translated "Connect" button

		},
		ResizeKeyboard: true,
	}

	bot.Send(adminNotification)
	bot.Send(userNotification)

	lastQueueNotificationMutex.Lock()
	lastQueueNotification = ""
	lastQueueNotificationMutex.Unlock()
}

func getNextUserFromQueue() (int64, bool) {
	queueMutex.Lock()
	defer queueMutex.Unlock()
	if len(userQueue) > 0 {
		userID := userQueue[0]
		userQueue = userQueue[1:] // Remove the first element
		log.Printf("User %d removed from queue and returned. Current queue: %v", userID, userQueue)
		return userID, true
	}
	return 0, false // Queue is empty
}

func sendMainMenu(bot *tgbotapi.BotAPI, chatID int64, lang string) {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(Translations[lang]["permit_menu_button"]),
			//tgbotapi.NewKeyboardButton(translations[lang]["contact_info_button"]),
			tgbotapi.NewKeyboardButton(Translations[lang]["support_button"]), // Add support button
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(Translations[lang]["borhat_button"]),
			tgbotapi.NewKeyboardButton(Translations[lang]["ijozatnoma_button"]),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(Translations[lang]["roxkhat_button"]),
			tgbotapi.NewKeyboardButton(Translations[lang]["certificate_button"]),
		),
	)

	msg := tgbotapi.NewMessage(chatID, Translations[lang]["main_menu"])
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}
func handleCloseCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, lang string) {
	chatID := msg.Chat.ID
	var otherChatID int64
	var isSessionActive bool

	if val, ok := activeSessions.Load(chatID); ok {
		otherChatID = val.(int64)
		isSessionActive = true
	}

	if !isSessionActive {
		reply := tgbotapi.NewMessage(chatID, Translations[lang]["no_active_session"])
		reply.ReplyMarkup = &tgbotapi.ReplyKeyboardMarkup{
			Keyboard: [][]tgbotapi.KeyboardButton{
				{
					tgbotapi.NewKeyboardButton(Translations[lang]["menu"]),
				},
			},
			ResizeKeyboard: true,
		}
		bot.Send(reply)
		return
	}

	// Remove the session from the active sessions
	activeSessions.Delete(chatID)
	activeSessions.Delete(otherChatID)
	removeUserFromQueue(otherChatID) // Remove the user from the queue

	// Notify both parties about the session closure
	notificationToUser := tgbotapi.NewMessage(otherChatID, Translations[lang]["session_ended"])
	notificationToUser.ReplyMarkup = &tgbotapi.ReplyKeyboardMarkup{
		Keyboard: [][]tgbotapi.KeyboardButton{
			{tgbotapi.NewKeyboardButton(Translations[lang]["menu"])},
		},
		ResizeKeyboard: true,
	}
	notificationToAdmin := tgbotapi.NewMessage(chatID, fmt.Sprintf(Translations[lang]["session_ended_admin"], otherChatID))
	notificationToAdmin.ReplyMarkup = &tgbotapi.ReplyKeyboardMarkup{
		Keyboard: [][]tgbotapi.KeyboardButton{
			{
				tgbotapi.NewKeyboardButton(Translations[lang]["menu"]),
				tgbotapi.NewKeyboardButton(Translations[lang]["view_queue"]), // <-- New "Queue" button
			},
			{
				tgbotapi.NewKeyboardButton(Translations[lang]["connect"]),
			},
		},
		ResizeKeyboard: true,
	}

	bot.Send(notificationToUser)
	bot.Send(notificationToAdmin)

	lastQueueNotificationMutex.Lock()
	lastQueueNotification = "" // Reset admin notification after close
	lastQueueNotificationMutex.Unlock()
}

func removeUserFromQueue(userID int64) {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	for i, id := range userQueue {
		if id == userID {
			// Remove the user from the queue
			userQueue = append(userQueue[:i], userQueue[i+1:]...)
			log.Printf("User %d removed from queue. Current queue: %v", userID, userQueue)
			break
		}
	}
}

// Sends language selection buttons
func sendLanguageSelection(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, Translations["ru"]["choose_lang"]) // Use Russian as base for initial message
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Русский 🇷🇺"),
			tgbotapi.NewKeyboardButton("Тоҷикӣ 🇹🇯"),
		),
	)

	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}
