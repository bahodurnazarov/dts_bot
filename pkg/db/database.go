package db

import (
	"context"
	"github.com/jackc/pgx/v5"
	"log"
)

// DB is the database connection
var DB *pgx.Conn

// ConnectDatabase initializes the database connection
func ConnectDatabase(dbURL string) {
	var err error
	DB, err = pgx.Connect(context.Background(), dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Connected to database!")
}

// CloseDatabase closes the database connection
func CloseDatabase() {
	if DB != nil {
		DB.Close(context.Background())
	}
}
