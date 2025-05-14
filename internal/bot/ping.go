package bot

import (
	"dts_bot/pkg/db"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"net/http"
	"time"
)

type Server struct {
	URL  string
	Name string
}

func CheckServers(bot *tgbotapi.BotAPI) {
	servers := []Server{
		{"http://94.199.18.153", "Сервер Сохторҳои зертобеъ"},
		{"http://185.177.0.112", "Сервер Сохторҳои зертобеъ"},
		{"http://10.10.25.25", "Сервер Сохторҳои зертобеъ"},
		{"http://10.10.25.239:8089", "Сервер НЗН"},
		{"http://gpsonline.tj", "Сервер НЗН"},
		{"http://10.230.71.156:3335", "Сервер НЗН"},
		{"http://taj.solidstreet.eu", "Сервер НИДР"},
		{"http://192.168.5.250:8088", "Сервер Общий"},
	}

	// Define the number of iterations you want (2 times a day)
	for i := 0; i < 2; i++ {
		for _, server := range servers {
			pingServer(server, bot)
		}
		// Sleep for 12 hours (12 hours = 12 * 60 * 60 seconds)
		time.Sleep(12 * time.Hour)
	}
}

func pingServer(server Server, bot *tgbotapi.BotAPI) {
	resp, err := http.Get(server.URL)
	if err != nil {
		log.Printf("Ошибка при проверке сервера %s (%s): %v", server.Name, server.URL, err)
		SendMessageToAllUsers(bot, fmt.Sprintf("❌ %s (%s) не работает. Пожалуйста, проверьте: %v", server.Name, server.URL, err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		log.Printf("✅ Сервер %s (%s) работает корректно.", server.Name, server.URL)
	} else {
		log.Printf("⚠️ %s (%s) is down. Status Code: %d", server.Name, server.URL, resp.StatusCode)
		SendMessageToAllUsers(bot, fmt.Sprintf("⚠️ %s (%s) не отвечает. Код состояния: %d", server.Name, server.URL, resp.StatusCode))
	}
}

// Function to send a message to all users
func SendMessageToAllUsers(bot *tgbotapi.BotAPI, message string) {
	// Fetch all chat IDs from the database
	chatIDs, err := db.GetAllChatIDs()
	if err != nil {
		log.Println("Error fetching chat IDs:", err)
		return
	}

	// Iterate over each chat ID and send the message
	for _, chatID := range chatIDs {
		msg := tgbotapi.NewMessage(chatID, message)
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Error sending message to chat %d: %v", chatID, err)
		} else {
			log.Printf("Message sent to chat %d", chatID)
		}
	}
}
