package handler

type UserSession struct {
	CurrentMenu string
}

var userSessions = make(map[int64]UserSession)

func getCurrentMenu(chatID int64) string {
	if session, exists := userSessions[chatID]; exists {
		return session.CurrentMenu
	}
	return ""
}

func SetCurrentMenu(chatID int64, menu string) {
	userSessions[chatID] = UserSession{CurrentMenu: menu}
}
