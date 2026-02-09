package config

import (
	"os"
	"strings"
)

type Config struct {
	Database DatabaseConfig
	Kafka    KafkaConfig
	Binance  BinanceConfig
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

type BinanceConfig struct {
	APIURL  string
	WSURL   string
	Symbols []string
}

type ServerConfig struct {
	Port    string
	GinMode string
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
			Topic:  getEnv("KAFKA_TOPIC", "market_price_updates"),
		},
		Binance: BinanceConfig{
			APIURL:  getEnv("BINANCE_API_URL", "https://api.binance.com"),
			WSURL:   getEnv("BINANCE_WS_URL", "wss://stream.binance.com:9443"),
			Symbols: parseSymbols(getEnv("DEFAULT_SYMBOLS", "BTCUSDT,ETHUSDT")),
		},
		Server: ServerConfig{
			Port:    getEnv("SERVER_PORT", "8082"),
			GinMode: getEnv("GIN_MODE", "debug"),
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

func parseSymbols(symbolsStr string) []string {
	if symbolsStr == "" {
		return []string{}
	}
	symbols := strings.Split(symbolsStr, ",")
	for i := range symbols {
		symbols[i] = strings.TrimSpace(symbols[i])
	}
	return symbols
}
