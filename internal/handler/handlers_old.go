package handler

//type UserSession struct {
//	CurrentMenu string // "permit", "borhat", "ijozatnoma", etc.
//}
//
//var userSessions = make(map[int64]UserSession)
//
//func getCurrentMenu(chatID int64) string {
//	if session, exists := userSessions[chatID]; exists {
//		return session.CurrentMenu
//	}
//	return ""
//}
//
//func SetCurrentMenu(chatID int64, menu string) {
//	userSessions[chatID] = UserSession{CurrentMenu: menu}
//}

//func HandleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
//	chatID := message.Chat.ID
//
//	// Fetch user language from DB
//	userLang, err := db.GetUserLanguage(chatID)
//	if err != nil {
//		log.Printf("Error fetching user language for chat %d: %v", chatID, err)
//		userLang = "ru" // fallback to Russian if there's an error
//	}
//
//	lang := userLang
//	// Ensure the selected language exists in the translations map
//	if _, exists := Translations[lang]; !exists {
//		log.Printf("Unsupported language '%s' for chat %d, defaulting to 'ru'", lang, chatID)
//		lang = "ru" // Fallback to Russian if the language doesn't exist
//	}
//
//	// First check if this is a question selection (starts with number and dot)
//	if isQuestionSelection(message.Text) {
//		handleSelectedQuestion(bot, message, lang)
//		return
//	}
//
//	switch message.Text {
//	case "/start":
//		sendLanguageSelection(bot, chatID)
//	case "Русский 🇷🇺":
//		setUserLanguage(bot, message, "ru")
//		sendMainMenu(bot, chatID, "ru")
//	case Translations[lang]["back"]:
//		sendMainMenu(bot, chatID, lang)
//	case "Тоҷикӣ 🇹🇯":
//		setUserLanguage(bot, message, "tg")
//		sendMainMenu(bot, chatID, "tg")
//	//case translations[lang]["contact_info_button"]:
//	//	sendContactInfo(bot, chatID, lang)
//	case Translations[lang]["permit_menu_button"]:
//		ShowPermitMenu(bot, chatID, lang)
//	case Translations[lang]["borhat_button"]:
//		ShowBorhatMenu(bot, chatID, lang)
//	case Translations[lang]["ijozatnoma_button"]:
//		ShowIjozatnomaMenu(bot, chatID, lang)
//	case Translations[lang]["roxkhat_button"]:
//		ShowRoxkhatMenu(bot, chatID, lang)
//	case Translations[lang]["certificate_button"]:
//		ShowCertificateMenu(bot, chatID, lang)
//	default:
//		if !isQuestionSelection(message.Text) {
//			msg := tgbotapi.NewMessage(chatID, Translations[lang]["welcome"])
//			bot.Send(msg)
//		}
//	}
//}

//func HandleCallbackQuery(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery) {
//	chatID := callbackQuery.Message.Chat.ID
//	data := callbackQuery.Data
//	lang := "ru" // default to Russian
//
//	// Get user language from DB
//	userLang, err := db.GetUserLanguage(chatID)
//	if err != nil {
//		log.Printf("Error fetching user language for chat %d: %v", chatID, err)
//	} else {
//		lang = userLang
//	}
//
//	switch {
//	case data == "main_menu":
//		sendMainMenu(bot, chatID, lang)
//		bot.Send(tgbotapi.NewCallback(callbackQuery.ID, ""))
//
//	case strings.HasPrefix(data, "permit_answer_"):
//		// Extract question ID from callback data
//		parts := strings.Split(data, "_")
//		if len(parts) < 3 {
//			return
//		}
//		questionID, err := strconv.Atoi(parts[2])
//		if err != nil {
//			return
//		}
//
//		// Get the full answer from DB
//		answer, err := db.GetAnswerByID(questionID, lang)
//		if err != nil {
//			msg := tgbotapi.NewMessage(chatID, Translations[lang]["error_fetching_answer"])
//			bot.Send(msg)
//			return
//		}
//
//		// Send the answer to the user
//		msg := tgbotapi.NewMessage(chatID, answer)
//		bot.Send(msg)
//
//		// Acknowledge the callback
//		bot.Send(tgbotapi.NewCallback(callbackQuery.ID, ""))
//	}
//}

//func sendMainMenu(bot *tgbotapi.BotAPI, chatID int64, lang string) {
//	keyboard := tgbotapi.NewReplyKeyboard(
//		tgbotapi.NewKeyboardButtonRow(
//			tgbotapi.NewKeyboardButton(Translations[lang]["permit_menu_button"]),
//			//tgbotapi.NewKeyboardButton(translations[lang]["contact_info_button"]),
//		),
//		tgbotapi.NewKeyboardButtonRow(
//			tgbotapi.NewKeyboardButton(Translations[lang]["borhat_button"]),
//			tgbotapi.NewKeyboardButton(Translations[lang]["ijozatnoma_button"]),
//		),
//		tgbotapi.NewKeyboardButtonRow(
//			tgbotapi.NewKeyboardButton(Translations[lang]["roxkhat_button"]),
//			tgbotapi.NewKeyboardButton(Translations[lang]["certificate_button"]),
//		),
//	)
//
//	msg := tgbotapi.NewMessage(chatID, Translations[lang]["main_menu"])
//	msg.ReplyMarkup = keyboard
//	bot.Send(msg)
//}
//
//// Sends language selection buttons
//func sendLanguageSelection(bot *tgbotapi.BotAPI, chatID int64) {
//	msg := tgbotapi.NewMessage(chatID, Translations["ru"]["choose_lang"]) // Use Russian as base for initial message
//	keyboard := tgbotapi.NewReplyKeyboard(
//		tgbotapi.NewKeyboardButtonRow(
//			tgbotapi.NewKeyboardButton("Русский 🇷🇺"),
//			tgbotapi.NewKeyboardButton("Тоҷикӣ 🇹🇯"),
//		),
//	)
//
//	msg.ReplyMarkup = keyboard
//	bot.Send(msg)
//}

//
//// Translations for bot messages
//var Translations = map[string]map[string]string{
//	"ru": {
//		"welcome":                  "Добро пожаловать в цифровую транспортную систему!",
//		"lang_set":                 "Язык установлен: Русский 🇷🇺",
//		"main_menu":                "Главное меню:",
//		"choose_lang":              "Пожалуйста, выберите язык:",
//		"permit_menu":              "Выберите вопрос о разрешениях:",
//		"back":                     "⬅️ Назад в меню",
//		"permit_menu_button":       "Международное разрешение",
//		"contact_info_button":      "ℹ️ Контакты и помощь",
//		"error_fetching_questions": "Ошибка при получении вопросов",
//		"invalid_question":         "Неверный номер вопроса",
//		"error_fetching_answer":    "Ошибка при получении ответа",
//		"borhat_button":            "Товарно-транспортная накладная",
//		"ijozatnoma_button":        "Лицензия",
//		"roxkhat_button":           "Путевой лист",
//		"certificate_button":       "Сертификат",
//	},
//	"tg": {
//		"welcome":                  "Хуш омадед ба низоми рақамии нақлиёт!",
//		"lang_set":                 "Забон интихоб шуд: Тоҷикӣ 🇹🇯",
//		"main_menu":                "Менюи асосӣ:",
//		"choose_lang":              "Лутфан забонро интихоб кунед:",
//		"permit_menu":              "Дар бораи рухсатномаҳо савол интихоб кунед:",
//		"back":                     "⬅️ Бозгашт ба меню",
//		"permit_menu_button":       "Рухсатнома",
//		"contact_info_button":      "ℹ️ Тамос ва кумак",
//		"error_fetching_questions": "Хатоги дар гирифтани саволҳо",
//		"error_fetching_answer":    "Хатоги дар гирифтани ҷавоб",
//		"invalid_question":         "Рақами савол нодуруст",
//		"borhat_button":            "Борхат",
//		"ijozatnoma_button":        "Иҷозатнома",
//		"roxkhat_button":           "Роххат",
//		"certificate_button":       "Сертификат",
//	},
//}
//
//func handleSelectedQuestion(bot *tgbotapi.BotAPI, message *tgbotapi.Message, lang string) {
//	chatID := message.Chat.ID
//
//	// Extract question number
//	dotIndex := strings.Index(message.Text, ".")
//	if dotIndex == -1 {
//		return
//	}
//	questionNum, err := strconv.Atoi(strings.TrimSpace(message.Text[:dotIndex]))
//	if err != nil {
//		return
//	}
//
//	// Get current menu type from session
//	currentMenu := getCurrentMenu(chatID)
//	if currentMenu == "" {
//		msg := tgbotapi.NewMessage(chatID, Translations[lang]["menu_session_expired"])
//		bot.Send(msg)
//		sendMainMenu(bot, chatID, lang)
//		return
//	}
//
//	// Get questions based on current menu
//	var questions []db.Question
//	switch currentMenu {
//	case "permit":
//		questions, err = db.GetPermitMenuQuestions(lang)
//	case "borhat":
//		questions, err = db.GetBorhatQuestions(lang)
//	case "ijozatnoma":
//		questions, err = db.GetIjozatnomaQuestions(lang)
//	case "roxkhat":
//		questions, err = db.GetRoxkhatQuestions(lang)
//	case "certificate":
//		questions, err = db.GetCertificateQuestions(lang)
//	default:
//		msg := tgbotapi.NewMessage(chatID, Translations[lang]["invalid_menu"])
//		bot.Send(msg)
//		return
//	}
//	log.Println("QUESTION :", questions)
//	if err != nil || len(questions) == 0 {
//		msg := tgbotapi.NewMessage(chatID, Translations[lang]["error_fetching_questions"])
//		bot.Send(msg)
//		return
//	}
//
//	// Validate question number
//	if questionNum < 1 || questionNum > len(questions) {
//		msg := tgbotapi.NewMessage(chatID, Translations[lang]["invalid_question"])
//		bot.Send(msg)
//		return
//	}
//
//	// Get and send the answer
//	answer, err := db.GetAnswerByID(questions[questionNum-1].ID, lang)
//	if err != nil {
//		msg := tgbotapi.NewMessage(chatID, Translations[lang]["error_fetching_answer"])
//		bot.Send(msg)
//		return
//	}
//
//	msg := tgbotapi.NewMessage(chatID, answer)
//	bot.Send(msg)
//
//	// Show the same menu again
//	switch currentMenu {
//	case "permit":
//		ShowPermitMenu(bot, chatID, lang)
//	case "borhat":
//		ShowBorhatMenu(bot, chatID, lang)
//	case "ijozatnoma":
//		ShowIjozatnomaMenu(bot, chatID, lang)
//	case "roxkhat":
//		ShowRoxkhatMenu(bot, chatID, lang)
//	case "certificate":
//		ShowCertificateMenu(bot, chatID, lang)
//	}
//}
//
//func isQuestionSelection(text string) bool {
//	// Find the first dot in the text
//	dotIndex := strings.Index(text, ".")
//	if dotIndex == -1 {
//		return false
//	}
//
//	// Extract the number part before the dot
//	numberPart := strings.TrimSpace(text[:dotIndex])
//
//	// Try to convert to integer
//	_, err := strconv.Atoi(numberPart)
//	return err == nil
//}
//
//// Sets the user's language and updates the database
//func setUserLanguage(bot *tgbotapi.BotAPI, message *tgbotapi.Message, lang string) {
//	chatID := message.Chat.ID
//	firstName := message.Chat.FirstName
//	lastName := message.Chat.LastName
//	userName := message.Chat.UserName
//	if err := db.SetUserLanguageAndInfo(chatID, lang, firstName, lastName, userName); err != nil {
//		log.Printf("Error setting language for chat %d: %v", chatID, err)
//	}
//	msg := tgbotapi.NewMessage(chatID, Translations[lang]["lang_set"])
//	bot.Send(msg)
//}
