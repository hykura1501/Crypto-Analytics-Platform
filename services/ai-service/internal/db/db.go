package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/crypto-platform/ai-service/config"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func Init(cfg *config.Config) error {
	dsn := cfg.Database.DSN()

	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err := DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connection established")
	return nil
}

func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
