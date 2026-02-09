package config

import (
	"os"
)

type Config struct {
	Database DatabaseConfig
	Kafka    KafkaConfig
	API      APIConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

const (
	KafkaTopicNewsNewArticle          = "news_new_article"
	KafkaTopicNewsAnalyzeRssStructure = "news_analyze_rss_structure"
	KafkaTopicNewsAnalyzeCssSelector  = "news_analyze_css_selector"
)

var topics = []string{
	KafkaTopicNewsNewArticle,
	KafkaTopicNewsAnalyzeRssStructure,
	KafkaTopicNewsAnalyzeCssSelector,
}

type KafkaConfig struct {
	Broker  string
	Topics  []string
	GroupID string
}

type APIConfig struct {
	Port string
}

func Load() *Config {
	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "postgres"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "crypto_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		API: APIConfig{
			Port: getEnv("API_PORT", "8080"),
		},
		Kafka: KafkaConfig{
			Broker:  getEnv("KAFKA_BROKER", "kafka:9092"),
			Topics:  topics,
			GroupID: getEnv("KAFKA_GROUP_ID", "ai-service-group-v2"),
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
