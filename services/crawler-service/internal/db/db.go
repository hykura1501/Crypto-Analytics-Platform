package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"

	"github.com/crypto-platform/crawler-service/config"
)

func Connect(cfg *config.DatabaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Run migrations
	if err := RunMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("crawler-service: database connection established")
	return db, nil
}

func RunMigrations(db *sql.DB) error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS articles (
		id SERIAL PRIMARY KEY,
		source_id VARCHAR(100) NOT NULL,
		url VARCHAR(2000) UNIQUE NOT NULL,
		url_normalized VARCHAR(2000),
		content_hash VARCHAR(64),
		title VARCHAR(1000) NOT NULL,
		author VARCHAR(200),
		published_at TIMESTAMP,
		crawled_at TIMESTAMP NOT NULL DEFAULT NOW(),
		html_raw_path VARCHAR(500),
		content_text TEXT NOT NULL,
		language VARCHAR(5) DEFAULT 'en',
		tags TEXT[],
		summary TEXT,
		entities JSONB,
		sentiment_score FLOAT,
		embedding_vector FLOAT[],
		confidence FLOAT DEFAULT 1.0,
		event_time TIMESTAMP,
		event_type VARCHAR(50),
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_articles_source_id ON articles(source_id);
	CREATE INDEX IF NOT EXISTS idx_articles_url ON articles(url);
	CREATE INDEX IF NOT EXISTS idx_articles_url_normalized ON articles(url_normalized);
	CREATE INDEX IF NOT EXISTS idx_articles_published_at ON articles(published_at);
	CREATE INDEX IF NOT EXISTS idx_articles_event_type ON articles(event_type);
	CREATE INDEX IF NOT EXISTS idx_articles_content_hash ON articles(content_hash);
	`

	if _, err := db.Exec(createTableSQL); err != nil {
		return fmt.Errorf("failed to create articles table: %w", err)
	}

	log.Println("✅ Articles table migration completed")
	return nil
}
