package config

import (
	"fmt"
	"os"
)

type Config struct {
	Database DatabaseConfig
	Kafka    KafkaConfig
	RSS      RSSConfig
	Crawler  CrawlerConfig
	Server   ServerConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type KafkaConfig struct {
	Broker string
	Topic  string
}

type RSSConfig struct {
	CoinDeskURL       string
	CoinTelegraphURL  string
	VNExpressURL      string
	VnEconomyURL      string
	MaxArticlesPerRun int
}

type CrawlerConfig struct {
	IntervalMinutes int
}

type ServerConfig struct {
	Port string
}

func Load() *Config {
	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "crypto_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Kafka: KafkaConfig{
			Broker: getEnv("KAFKA_BROKER", "localhost:29092"),
			Topic:  getEnv("KAFKA_TOPIC", "news_new_article"),
		},
		RSS: RSSConfig{
			CoinDeskURL:       getEnv("COINDESK_RSS_URL", "https://www.coindesk.com/arc/outboundfeeds/rss/"),
			CoinTelegraphURL:  getEnv("COINTELEGRAPH_RSS_URL", "https://cointelegraph.com/rss"),
			VNExpressURL:      getEnv("VNEXPRESS_RSS_URL", "https://vnexpress.net/rss/kinh-doanh.rss"),
			VnEconomyURL:      getEnv("VNECONOMY_RSS_URL", "https://vneconomy.vn/thi-truong-chung-khoan.rss"),
			MaxArticlesPerRun: getEnvInt("MAX_ARTICLES_PER_RUN", 20),
		},
		Crawler: CrawlerConfig{
			IntervalMinutes: getEnvInt("CRAWL_INTERVAL_MINUTES", 10),
		},
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8083"),
		},
	}
}

func (c *DatabaseConfig) DSN() string {
	return "host=" + c.Host +
		" port=" + c.Port +
		" user=" + c.User +
		" password=" + c.Password +
		" dbname=" + c.DBName +
		" sslmode=" + c.SSLMode
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var v int
		_, err := fmt.Sscanf(value, "%d", &v)
		if err == nil {
			return v
		}
	}
	return defaultValue
}
