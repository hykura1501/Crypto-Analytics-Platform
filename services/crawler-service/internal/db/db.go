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
	CREATE TABLE IF NOT EXISTS sources (
		source_id VARCHAR(100) PRIMARY KEY,
		rss_url VARCHAR(2000) NOT NULL,
		title_tag VARCHAR(100),
		link_tag VARCHAR(100),
		pub_date_tag VARCHAR(100),
		summary_selector TEXT,
		content_selector TEXT,
		author_selector TEXT,
		tags_selector TEXT,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS articles (
		id SERIAL PRIMARY KEY,
		source_id VARCHAR(100) NOT NULL,
		url VARCHAR(2000) UNIQUE NOT NULL,
		title VARCHAR(1000) NOT NULL,
		author VARCHAR(1000),
		published_at VARCHAR(150),
		crawled_at TIMESTAMP NOT NULL DEFAULT NOW(),
		content_text TEXT NOT NULL,
		language VARCHAR(5) DEFAULT 'en',
		tags VARCHAR(1000),
		summary TEXT,
		sentiment_score FLOAT,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_articles_source_id ON articles(source_id);
	CREATE INDEX IF NOT EXISTS idx_articles_url ON articles(url);
	CREATE INDEX IF NOT EXISTS idx_articles_published_at ON articles(published_at);
	`

	if _, err := db.Exec(createTableSQL); err != nil {
		return fmt.Errorf("failed to create articles table: %w", err)
	}

	log.Println("✅ Articles table migration completed")
	return nil
}
