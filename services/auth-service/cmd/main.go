package main

import (
	"fmt"
	"log"

	"github.com/crypto-platform/auth-service/config"
	"github.com/crypto-platform/auth-service/internal/handler"
	"github.com/crypto-platform/auth-service/internal/model"
	"github.com/crypto-platform/auth-service/internal/repository"
	"github.com/crypto-platform/auth-service/internal/router"
	"github.com/crypto-platform/auth-service/internal/service"
	"github.com/crypto-platform/auth-service/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Set Gin mode
	gin.SetMode(cfg.Server.GinMode)

	// Connect to database
	db, err := initDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize JWT manager
	jwtManager := utils.NewJWTManager(
		cfg.JWT.Secret,
		cfg.JWT.AccessExpiry,
		cfg.JWT.RefreshExpiry,
	)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, refreshTokenRepo, jwtManager)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)

	// Setup router
	r := router.SetupRouter(authHandler, jwtManager)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Starting auth-service on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func initDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.Database.DSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Database connection established")
	return db, nil
}

func runMigrations(db *gorm.DB) error {
	log.Println("Running database migrations...")

	if err := db.AutoMigrate(
		&model.User{},
		&model.RefreshToken{},
	); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Ensure at least one ADMIN exists: promote first user if no admin
	var adminCount int64
	if err := db.Model(&model.User{}).Where("role = ?", model.RoleAdmin).Count(&adminCount).Error; err != nil {
		return fmt.Errorf("failed to count admins: %w", err)
	}
	if adminCount == 0 {
		var first model.User
		if err := db.Order("id ASC").First(&first).Error; err == nil {
			first.Role = model.RoleAdmin
			if err := db.Save(&first).Error; err != nil {
				return fmt.Errorf("failed to seed admin: %w", err)
			}
			log.Printf("Seeded first user (id=%d, email=%s) as ADMIN", first.ID, first.Email)
		}
	}

	log.Println("Migrations completed successfully")
	return nil
}
