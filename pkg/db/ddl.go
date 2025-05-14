package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"log"
)

// SetUserLanguageAndInfo stores the user's language preference and additional info in the database
func SetUserLanguageAndInfo(chatID int64, lang, firstName, lastName, username string) error {
	// Insert or update user information, including language and other fields
	_, err := DB.Exec(context.Background(), `
		INSERT INTO users (chat_id, language, first_name, last_name, username) 
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (chat_id) 
		DO UPDATE 
		SET language = $2, first_name = $3, last_name = $4, username = $5`,
		chatID, lang, firstName, lastName, username)
	if err != nil {
		log.Println("Error saving user information:", err)
		return err
	}
	return nil
}

// GetUserState retrieves the user's state from the database.
func GetUserState(chatID int64) (string, error) {
	query := "SELECT state FROM users WHERE chat_id = $1" // Adjust table and column names as necessary.  Use 'users'
	row := DB.QueryRow(context.Background(), query, chatID)

	var state string
	err := row.Scan(&state)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("user state not found for chat ID %d", chatID)
		}
		return "", fmt.Errorf("error querying user state: %w", err)
	}
	return state, nil
}

// SetUserState sets the user's state in the database.
func SetUserState(chatID int64, state string) error {
	query := "UPDATE users SET state = $1 WHERE chat_id = $2" //  Use 'users' table
	_, err := DB.Exec(context.Background(), query, state, chatID)
	if err != nil {
		return fmt.Errorf("error updating user state: %w", err)
	}
	return nil
}

// GetUserLanguage retrieves the user's language from the database.
func GetUserLanguage(chatID int64) (string, error) {
	var lang string
	query := "SELECT language FROM users WHERE chat_id = $1" // Use 'users' table
	row := DB.QueryRow(context.Background(), query, chatID)
	err := row.Scan(&lang)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Println("User not found, defaulting to Russian") // Changed default to Russian
			return "ru", nil                                     // Return "ru" and nil error for not found
		}
		log.Println("Error fetching language:", err, chatID)           // Keep the log
		return "", fmt.Errorf("error querying user language: %w", err) // wrap
	}
	log.Println("Lang :", lang, chatID)
	return lang, nil
}

// SetUserLanguage sets the user's language in the database.
func SetUserLanguage(chatID int64, lang string) error {
	query := "UPDATE users SET language = $1 WHERE chat_id = $2"
	_, err := DB.Exec(context.Background(), query, lang, chatID)
	if err != nil {
		return fmt.Errorf("error setting user language: %w", err)
	}
	return nil
}

// Function to get all chat IDs from the database
func GetAllChatIDs() ([]int64, error) {
	var chatIDs []int64
	rows, err := DB.Query(context.Background(), "SELECT chat_id FROM users")
	if err != nil {
		log.Printf("Error fetching chat IDs: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var chatID int64
		if err := rows.Scan(&chatID); err != nil {
			log.Printf("Error scanning chat ID: %v", err)
			continue
		}
		chatIDs = append(chatIDs, chatID)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
		return nil, err
	}

	return chatIDs, nil
}

// In your db package
type Question struct {
	ID       int
	Question string
	Category string
	Language string
}

func GetPermitMenuQuestions(lang string) ([]Question, error) {
	return getQuestionsByCategory("permit", lang)
}
func GetBorhatQuestions(lang string) ([]Question, error) {
	return getQuestionsByCategory("borhat", lang)
}

func GetIjozatnomaQuestions(lang string) ([]Question, error) {
	return getQuestionsByCategory("ijozatnoma", lang)
}

func GetRoxkhatQuestions(lang string) ([]Question, error) {
	return getQuestionsByCategory("roxkhat", lang)
}

func GetCertificateQuestions(lang string) ([]Question, error) {
	return getQuestionsByCategory("certificate", lang)
}

func getQuestionsByCategory(category string, lang string) ([]Question, error) {
	var questions []Question

	log.Println("CAtegoryyy : ", category)
	query := `
        SELECT id, command 
        FROM directory_menu 
        WHERE category = $1 AND language = $2
        ORDER BY id`

	rows, err := DB.Query(context.Background(), query, category, lang)
	if err != nil {
		log.Printf("Error fetching %s questions: %v", category, err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var q Question
		if err := rows.Scan(&q.ID, &q.Question); err != nil {
			log.Printf("Error scanning question row: %v", err)
			return nil, err
		}
		q.Category = category
		q.Language = lang
		questions = append(questions, q)
		log.Printf("Loaded %s question [%d]: %s", category, q.ID, q.Question)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error in question rows: %v", err)
		return nil, err
	}

	return questions, nil
}

func GetAnswerByID(id int, lang string) (string, error) {
	var answer string
	query := `
        SELECT answer 
        FROM directory_menu 
        WHERE id = $1 AND language = $2`

	err := DB.QueryRow(context.Background(), query, id, lang).Scan(&answer)
	if err != nil {
		log.Printf("Error fetching answer for ID %d (%s): %v", id, lang, err)
		return "", err
	}
	return answer, nil
}
