package config

import (
	"os"
)

type Config struct {
	Database DatabaseConfig
	Kafka    KafkaConfig
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
	Broker  string
	Topic   string
	GroupID string
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
		Kafka: KafkaConfig{
			Broker:  getEnv("KAFKA_BROKER", "kafka:9092"),
			Topic:   getEnv("KAFKA_TOPIC", "news_new_article"),
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
