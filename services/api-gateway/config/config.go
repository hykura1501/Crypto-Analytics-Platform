package config

import (
	"log"
	"os"
	"strings"
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
		AllowedOrigins: func() []string {
			origins := getEnv("ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173,http://127.0.0.1:3000,http://127.0.0.1:5173")
			if origins == "" {
				return []string{"http://localhost:3000"}
			}
			// Split by comma and trim spaces
			result := []string{}
			for _, origin := range strings.Split(origins, ",") {
				trimmed := strings.TrimSpace(origin)
				if trimmed != "" {
					result = append(result, trimmed)
				}
			}
			if len(result) == 0 {
				return []string{"http://localhost:3000"}
			}
			return result
		}(),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	log.Printf("Environment variable %s not set, using default: %s", key, fallback)
	return fallback
}
