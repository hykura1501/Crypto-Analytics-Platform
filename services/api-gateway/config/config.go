package config

import (
	"log"
	"os"
)

type Config struct {
	Port              string
	JWTSecret         string
	AuthServiceURL    string
	MarketServiceURL  string
	CrawlerServiceURL string
	AllowedOrigins    []string
}

func Load() *Config {
	return &Config{
		Port:              getEnv("PORT", "8080"),
		JWTSecret:         getEnv("JWT_SECRET", "your-secret-key"),
		AuthServiceURL:    getEnv("AUTH_SERVICE_URL", "http://auth-service:8081"),
		MarketServiceURL:  getEnv("MARKET_SERVICE_URL", "http://market-service:8082"),
		CrawlerServiceURL: getEnv("CRAWLER_SERVICE_URL", "http://crawler-service:8083"),
		AllowedOrigins: []string{
			getEnv("ALLOWED_ORIGIN", "http://localhost:3000"),
		},
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	log.Printf("Environment variable %s not set, using default: %s", key, fallback)
	return fallback
}
