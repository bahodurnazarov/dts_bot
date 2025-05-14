package config

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Constants
const (
	ContactSupportCallback = "contact_support"
	CloseCommand           = "/close"
	StartCommand           = "/start"
	ConnectCommand         = "/connect"
)

// AdminUserID should ideally be loaded from an environment variable or config.
const AdminUserID int64 = 8153143177 // ID администратора, измените на ID вашего администратора

// Active support sessions: map of user ID to admin's chat ID (or vice-versa)
var activeSessions = sync.Map{}

// User queue for support requests.  Using a mutex for thread safety.
var userQueue []int64
var queueMutex sync.Mutex

// Last queue notification sent to admin
var lastQueueNotification string
var lastQueueNotificationMutex sync.Mutex

// Function to add a user to the queue
func addUserToQueue(userID int64) {
	queueMutex.Lock()
	defer queueMutex.Unlock()
	userQueue = append(userQueue, userID)
	log.Printf("User %d added to queue. Current queue: %v", userID, userQueue)
}

// Function to remove a user from the queue
func removeUserFromQueue(userID int64) {
	queueMutex.Lock()
	defer queueMutex.Unlock()
	for i, id := range userQueue {
		if id == userID {
			userQueue = append(userQueue[:i], userQueue[i+1:]...)
			log.Printf("User %d removed from queue. Current queue: %v", userID, userQueue)
			return
		}
	}
	log.Printf("User %d not found in queue.", userID) // User was not in queue
}

// Function to get the next user from the queue
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

// Function to get the current queue
func getCurrentQueue() []int64 {
	queueMutex.Lock()
	defer queueMutex.Unlock()
	queueCopy := make([]int64, len(userQueue))
	copy(queueCopy, userQueue)
	return queueCopy
}

// Function to format the queue for admin notification
func formatQueueNotification() string {
	queueList := getCurrentQueue()
	if len(queueList) == 0 {
		return "Очередь пуста."
	}
	queueString := "Текущая очередь:\n"
	for i, userID := range queueList {
		queueString += fmt.Sprintf("%d: %d\n", i+1, userID)
	}
	return queueString
}

// sendAdminQueueNotification sends the queue status to the admin if it has changed.
func sendAdminQueueNotification(bot *tgbotapi.BotAPI, newUserAdded bool) {
	currentNotification := formatQueueNotification()
	lastQueueNotificationMutex.Lock()
	defer lastQueueNotificationMutex.Unlock()
	if currentNotification != lastQueueNotification && newUserAdded { //send message only when new user added to queue
		msg := tgbotapi.NewMessage(AdminUserID, fmt.Sprintf("🔔 Новый запрос в очереди.\n%sПодключитесь: /connect", currentNotification))
		msg.ReplyMarkup = &tgbotapi.ReplyKeyboardMarkup{
			Keyboard: [][]tgbotapi.KeyboardButton{
				{
					tgbotapi.NewKeyboardButton(ConnectCommand),
				},
				{
					tgbotapi.NewKeyboardButton(StartCommand),
					tgbotapi.NewKeyboardButton(CloseCommand),
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
			handleIncomingMessage(bot, update.Message)
		} else if update.CallbackQuery != nil {
			handleIncomingCallbackQuery(bot, update.CallbackQuery)
		}
	}
}

// handleIncomingCallbackQuery processes incoming callback queries.
func handleIncomingCallbackQuery(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery) {
	switch query.Data {
	case ContactSupportCallback:
		processContactSupportRequest(bot, query)
	}
}

// processContactSupportRequest handles the "contact_support" callback.
func processContactSupportRequest(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery) {
	userID := query.Message.Chat.ID

	// Acknowledge the button press
	callbackResponse := tgbotapi.NewCallback(query.ID, "Запрос отправлен оператору")
	if _, err := bot.Send(callbackResponse); err != nil {
		log.Printf("Error sending callback response: %v", err)
	}

	// Notify the user
	notificationMsg := tgbotapi.NewMessage(userID, "⏳ Ожидайте подключения оператора...")
	notificationMsg.ReplyMarkup = &tgbotapi.ReplyKeyboardMarkup{
		Keyboard: [][]tgbotapi.KeyboardButton{
			{
				tgbotapi.NewKeyboardButton(CloseCommand),
			},
		},
		ResizeKeyboard: true,
	}

	if _, err := bot.Send(notificationMsg); err != nil {
		log.Printf("Error sending user notification: %v", err)
	}

	// Add the user to the queue
	addUserToQueue(userID)
	sendAdminQueueNotification(bot, true) // Notify admin about new user
}

// handleIncomingMessage processes incoming text messages.
func handleIncomingMessage(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID

	// Check for the /close command first, regardless of the session status
	if msg.Text == CloseCommand {
		handleCloseCommand(bot, msg)
		return
	}

	// Check if this chat is part of an active support session
	if otherChatIDVal, ok := activeSessions.Load(chatID); ok {
		otherChatID := otherChatIDVal.(int64)
		var senderName string
		if chatID == AdminUserID {
			senderName = "Оператор"
		} else {
			senderName = "Клиент"
		}
		forwardedMessage := tgbotapi.NewMessage(otherChatID, fmt.Sprintf("👤 %s: %s", senderName, msg.Text))
		if _, err := bot.Send(forwardedMessage); err != nil {
			log.Printf("Error sending message in session: %v", err)
		}
		return
	}

	// Handle commands that are not part of an active session
	switch {
	case msg.Text == StartCommand:
		SendSupportMenu(bot, chatID)

	case msg.Text == ConnectCommand: // Admin uses /connect without user ID
		handleConnectCommand(bot, msg)

	case strings.HasPrefix(msg.Text, "/close_"): // Keep the old admin-specific close
		if msg.Chat.ID == AdminUserID {
			handleAdminCloseCommand(bot, msg)
		}
		return

	default:
		// If not a recognized command and not in a session, provide a default message
		defaultMsg := tgbotapi.NewMessage(chatID, "Введите /start, чтобы начать")
		defaultMsg.ReplyMarkup = &tgbotapi.ReplyKeyboardMarkup{
			Keyboard: [][]tgbotapi.KeyboardButton{
				{
					tgbotapi.NewKeyboardButton(StartCommand),
					tgbotapi.NewKeyboardButton(CloseCommand),
				},
			},
			ResizeKeyboard: true,
		}
		if _, err := bot.Send(defaultMsg); err != nil {
			log.Printf("Error sending default message: %v", err)
		}
	}
}

// handleConnectCommand allows the admin to connect to the *next* user in the queue.
func handleConnectCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	if msg.Chat.ID != AdminUserID {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "⚠️ Эта команда доступна только администраторам.")
		bot.Send(reply)
		return
	}

	userID, ok := getNextUserFromQueue() // Get the next user from the queue
	if !ok {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "❌ Очередь пуста.")
		reply.ReplyMarkup = &tgbotapi.ReplyKeyboardMarkup{
			Keyboard: [][]tgbotapi.KeyboardButton{
				{
					tgbotapi.NewKeyboardButton(ConnectCommand),
				},
				{
					tgbotapi.NewKeyboardButton(StartCommand),
					tgbotapi.NewKeyboardButton(CloseCommand),
				},
			},
			ResizeKeyboard: true,
		}
		bot.Send(reply)
		return
	}

	// Start the support session
	activeSessions.Store(userID, msg.Chat.ID)
	activeSessions.Store(msg.Chat.ID, userID)

	// Notify both the admin and the user about the connection
	adminNotification := tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("🤝 Вы подключились к пользователю %d. Теперь ваши сообщения будут отправляться напрямую.", userID))
	adminNotification.ReplyMarkup = &tgbotapi.ReplyKeyboardMarkup{
		Keyboard: [][]tgbotapi.KeyboardButton{
			{
				tgbotapi.NewKeyboardButton(CloseCommand),
			},
		},
		ResizeKeyboard: true,
	}
	userNotification := tgbotapi.NewMessage(userID, "💬 Оператор подключился. Теперь вы можете отправлять сообщения напрямую.")
	userNotification.ReplyMarkup = &tgbotapi.ReplyKeyboardMarkup{
		Keyboard: [][]tgbotapi.KeyboardButton{
			{
				tgbotapi.NewKeyboardButton(CloseCommand),
			},
		},
		ResizeKeyboard: true,
	}

	bot.Send(adminNotification)
	bot.Send(userNotification)

	lastQueueNotificationMutex.Lock()
	lastQueueNotification = "" // Reset admin notification after connect.
	lastQueueNotificationMutex.Unlock()

}

// handleCloseCommand allows either the admin or the user to close the support session.
func handleCloseCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	var otherChatID int64
	var isSessionActive bool

	if val, ok := activeSessions.Load(chatID); ok {
		otherChatID = val.(int64)
		isSessionActive = true
	}

	if !isSessionActive {
		reply := tgbotapi.NewMessage(chatID, "❌ Нет активного сеанса поддержки для закрытия.")
		reply.ReplyMarkup = &tgbotapi.ReplyKeyboardMarkup{
			Keyboard: [][]tgbotapi.KeyboardButton{
				{
					tgbotapi.NewKeyboardButton(StartCommand),
					tgbotapi.NewKeyboardButton(ConnectCommand),
					tgbotapi.NewKeyboardButton(CloseCommand),
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
	notificationToUser := tgbotapi.NewMessage(otherChatID, "✅ Сеанс поддержки завершен.")
	notificationToUser.ReplyMarkup = &tgbotapi.ReplyKeyboardMarkup{
		Keyboard: [][]tgbotapi.KeyboardButton{
			{
				tgbotapi.NewKeyboardButton(StartCommand),
			},
		},
		ResizeKeyboard: true,
	}
	notificationToAdmin := tgbotapi.NewMessage(chatID, fmt.Sprintf("✅ Сеанс поддержки с пользователем %d завершен.", otherChatID))
	notificationToAdmin.ReplyMarkup = &tgbotapi.ReplyKeyboardMarkup{
		Keyboard: [][]tgbotapi.KeyboardButton{
			{
				tgbotapi.NewKeyboardButton(ConnectCommand),
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

// handleAdminCloseCommand allows the admin to close a specific user's session.
func handleAdminCloseCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	if msg.Chat.ID != AdminUserID {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "⚠️ Эта команда доступна только администраторам.")
		bot.Send(reply)
		return
	}

	parts := strings.Split(msg.Text, "_")
	if len(parts) != 2 {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "❌ Неверный формат команды. Используйте: /close_<user_id>")
		bot.Send(reply)
		return
	}

	userIDStr := parts[1]
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "❌ Неверный user ID.")
		bot.Send(reply)
		return
	}

	adminChatID := msg.Chat.ID

	// Check if a session exists with this user
	if userAdminIDVal, userSessionActive := activeSessions.Load(userID); userSessionActive && userAdminIDVal.(int64) == adminChatID {
		activeSessions.Delete(userID)
		activeSessions.Delete(adminChatID)
		removeUserFromQueue(userID) // Remove user from queue
		// Notify both parties
		notificationToUser := tgbotapi.NewMessage(userID, "✅ Сеанс поддержки завершен оператором.")
		notificationToAdmin := tgbotapi.NewMessage(adminChatID, fmt.Sprintf("✅ Сеанс поддержки с пользователем %d завершен.", userID))
		notificationToAdmin.ReplyMarkup = &tgbotapi.ReplyKeyboardMarkup{
			Keyboard: [][]tgbotapi.KeyboardButton{
				{
					tgbotapi.NewKeyboardButton(ConnectCommand),
				},
			},
			ResizeKeyboard: true,
		}
		bot.Send(notificationToUser)
		bot.Send(notificationToAdmin)
	} else if adminUserIDVal, adminSessionActive := activeSessions.Load(adminChatID); adminSessionActive && adminUserIDVal.(int64) == userID {
		activeSessions.Delete(userID)
		activeSessions.Delete(adminChatID)
		removeUserFromQueue(userID)
		// Notify both parties
		notificationToUser := tgbotapi.NewMessage(userID, "✅ Сеанс поддержки завершен оператором.")
		notificationToAdmin := tgbotapi.NewMessage(adminChatID, fmt.Sprintf("✅ Сеанс поддержки с пользователем %d завершен.", userID))
		notificationToAdmin.ReplyMarkup = &tgbotapi.ReplyKeyboardMarkup{
			Keyboard: [][]tgbotapi.KeyboardButton{
				{
					tgbotapi.NewKeyboardButton(ConnectCommand),
				},
			},
			ResizeKeyboard: true,
		}
		bot.Send(notificationToUser)
		bot.Send(notificationToAdmin)
	} else {
		reply := tgbotapi.NewMessage(adminChatID, fmt.Sprintf("❌ Нет активного сеанса с пользователем %d.", userID))
		reply.ReplyMarkup = &tgbotapi.ReplyKeyboardMarkup{
			Keyboard: [][]tgbotapi.KeyboardButton{
				{
					tgbotapi.NewKeyboardButton(ConnectCommand),
				},
				{
					tgbotapi.NewKeyboardButton(StartCommand),
					tgbotapi.NewKeyboardButton(CloseCommand),
				},
			},
			ResizeKeyboard: true,
		}
		bot.Send(reply)
	}
}

// sendSupportMenu sends the initial support menu to the user.
func SendSupportMenu(bot *tgbotapi.BotAPI, chatID int64) {
	contactButton := tgbotapi.NewInlineKeyboardButtonData("📞 Связаться с поддержкой", ContactSupportCallback)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(contactButton))

	menuMsg := tgbotapi.NewMessage(chatID, "Добро пожаловать! Если у Вас есть вопрос, нажмите кнопку ниже:")
	menuMsg.ReplyMarkup = keyboard
	if _, err := bot.Send(menuMsg); err != nil {
		log.Printf("Error sending support menu: %v", err)
	}
}
