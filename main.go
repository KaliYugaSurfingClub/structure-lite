package main

import (
	"log"
	"log/slog"
	"os"
	"structure-lite/internal/tables"
	"time"
)

type User struct {
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Address   string    `json:"address"`
	Tags      []string  `json:"tags"`
	Age       int       `json:"age"`
	CreatedAt time.Time `json:"created_at"`
}

func main() {
	opts := slog.HandlerOptions{Level: slog.LevelInfo}
	handler := slog.NewTextHandler(os.Stdin, &opts)
	logger := slog.New(handler)

	location := "./temp_test_table"

	itemsOnPage := 2

	table, err := tables.New[User](itemsOnPage, location, logger)
	if err != nil {
		log.Fatalf("Failed to initialize UserManager: %v", err)
	}

	manager := NewUserManager(table)

	app := NewUserTableApp(manager)

	app.Run()
}
