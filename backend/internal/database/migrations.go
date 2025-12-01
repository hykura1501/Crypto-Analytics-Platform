package database

import (
	"fmt"
)

func RunMigrations() error {
	migrations := []string{
		`
		CREATE TABLE IF NOT EXISTS trading_pairs (
			id SERIAL PRIMARY KEY,
			symbol VARCHAR(20) UNIQUE NOT NULL,
			base_asset VARCHAR(10) NOT NULL,
			quote_asset VARCHAR(10) NOT NULL,
			status VARCHAR(20) DEFAULT 'active',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		`,
		`
		CREATE TABLE IF NOT EXISTS price_history (
			id SERIAL PRIMARY KEY,
			pair_id INTEGER REFERENCES trading_pairs(id),
			price DECIMAL(20, 8) NOT NULL,
			volume DECIMAL(20, 8) NOT NULL,
			high DECIMAL(20, 8) NOT NULL,
			low DECIMAL(20, 8) NOT NULL,
			open_price DECIMAL(20, 8) NOT NULL,
			close_price DECIMAL(20, 8) NOT NULL,
			timestamp TIMESTAMP NOT NULL,
			interval VARCHAR(10) NOT NULL DEFAULT '1m',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(pair_id, timestamp, interval)
		);
		CREATE INDEX IF NOT EXISTS idx_price_history_pair_timestamp ON price_history(pair_id, timestamp DESC);
		CREATE INDEX IF NOT EXISTS idx_price_history_interval ON price_history(interval);
		`,
		`
		CREATE TABLE IF NOT EXISTS news_sources (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) UNIQUE NOT NULL,
			url VARCHAR(255) NOT NULL,
			selector_type VARCHAR(50) DEFAULT 'css',
			title_selector TEXT,
			content_selector TEXT,
			date_selector TEXT,
			link_selector TEXT,
			status VARCHAR(20) DEFAULT 'active',
			last_crawled_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		`,
		`
		CREATE TABLE IF NOT EXISTS news (
			id SERIAL PRIMARY KEY,
			source_id INTEGER REFERENCES news_sources(id),
			title TEXT NOT NULL,
			content TEXT,
			url VARCHAR(500) UNIQUE NOT NULL,
			author VARCHAR(100),
			published_at TIMESTAMP NOT NULL,
			crawled_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			sentiment_score DECIMAL(5, 4),
			sentiment_label VARCHAR(20),
			keywords TEXT[],
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_news_published_at ON news(published_at DESC);
		CREATE INDEX IF NOT EXISTS idx_news_source_id ON news(source_id);
		CREATE INDEX IF NOT EXISTS idx_news_sentiment ON news(sentiment_label);
		`,
		`
		CREATE TABLE IF NOT EXISTS news_price_alignment (
			id SERIAL PRIMARY KEY,
			news_id INTEGER REFERENCES news(id) ON DELETE CASCADE,
			pair_id INTEGER REFERENCES trading_pairs(id),
			price_before DECIMAL(20, 8),
			price_after DECIMAL(20, 8),
			price_change_percent DECIMAL(10, 4),
			time_window_hours INTEGER DEFAULT 24,
			correlation_score DECIMAL(5, 4),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_alignment_news_pair ON news_price_alignment(news_id, pair_id);
		`,
		`
		CREATE TABLE IF NOT EXISTS ai_analysis (
			id SERIAL PRIMARY KEY,
			pair_id INTEGER REFERENCES trading_pairs(id),
			analysis_type VARCHAR(50) NOT NULL,
			prediction VARCHAR(20),
			confidence_score DECIMAL(5, 4),
			reasoning TEXT,
			time_horizon VARCHAR(20),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			valid_until TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_ai_analysis_pair ON ai_analysis(pair_id, created_at DESC);
		`,
		`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50) UNIQUE NOT NULL,
			email VARCHAR(100) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		`,
		`
		CREATE TABLE IF NOT EXISTS watchlists (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			pair_id INTEGER REFERENCES trading_pairs(id) ON DELETE CASCADE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, pair_id)
		);
		CREATE INDEX IF NOT EXISTS idx_watchlist_user ON watchlists(user_id);
		`,
	}

	for i, migration := range migrations {
		if _, err := DB.Exec(migration); err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}
	}

	// Insert default trading pairs
	defaultPairs := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT", "SOLUSDT", "ADAUSDT", "XRPUSDT"}
	for _, symbol := range defaultPairs {
		_, err := DB.Exec(`
			INSERT INTO trading_pairs (symbol, base_asset, quote_asset) 
			VALUES ($1, $2, $3)
			ON CONFLICT (symbol) DO NOTHING
		`, symbol, symbol[:len(symbol)-4], "USDT")
		if err != nil {
			return fmt.Errorf("failed to insert pair %s: %w", symbol, err)
		}
	}

	// Insert default news sources
	defaultSources := []struct {
		name, url, titleSel, contentSel, linkSel string
	}{
		{"CoinTelegraph", "https://cointelegraph.com", "h2.post-card-inline__title a", "div.post-card-inline__text", "h2.post-card-inline__title a"},
		{"CoinDesk", "https://www.coindesk.com", "h3.card-title a", "div.card-text", "h3.card-title a"},
	}

	for _, source := range defaultSources {
		_, err := DB.Exec(`
			INSERT INTO news_sources (name, url, title_selector, content_selector, link_selector) 
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (name) DO NOTHING
		`, source.name, source.url, source.titleSel, source.contentSel, source.linkSel)
		if err != nil {
			return fmt.Errorf("failed to insert source %s: %w", source.name, err)
		}
	}

	fmt.Println("Migrations completed successfully")
	return nil
}
