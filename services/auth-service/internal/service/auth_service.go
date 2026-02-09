package service

import (
	"errors"
	"time"

	"github.com/crypto-platform/auth-service/internal/model"
	"github.com/crypto-platform/auth-service/internal/repository"
	"github.com/crypto-platform/auth-service/pkg/utils"
	"gorm.io/gorm"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
)

type AuthService interface {
	Register(req *model.RegisterRequest) (*model.User, error)
	Login(req *model.LoginRequest) (*model.AuthResponse, error)
	RefreshToken(refreshToken string) (*model.AuthResponse, error)
	ValidateToken(token string) (*model.User, error)
	Logout(refreshToken string) error
}

type authService struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	jwtManager       *utils.JWTManager
}

func NewAuthService(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	jwtManager *utils.JWTManager,
) AuthService {
	return &authService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtManager:       jwtManager,
	}
}

func (s *authService) Register(req *model.RegisterRequest) (*model.User, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.FindByEmail(req.Email)
	if err == nil && existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &model.User{
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		IsActive:     true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *authService) Login(req *model.LoginRequest) (*model.AuthResponse, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Check password
	if err := utils.CheckPassword(req.Password, user.PasswordHash); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("user account is deactivated")
	}

	// Generate tokens
	return s.generateTokens(user)
}

func (s *authService) RefreshToken(refreshToken string) (*model.AuthResponse, error) {
	// Validate refresh token
	claims, err := s.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Check if refresh token exists in database
	tokenRecord, err := s.refreshTokenRepo.FindByToken(refreshToken)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid refresh token")
		}
		return nil, err
	}

	// Check if token is expired
	if tokenRecord.ExpiresAt.Before(time.Now()) {
		s.refreshTokenRepo.DeleteByToken(refreshToken)
		return nil, errors.New("refresh token expired")
	}

	// Get user
	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Delete old refresh token
	s.refreshTokenRepo.DeleteByToken(refreshToken)

	// Generate new tokens
	return s.generateTokens(user)
}

func (s *authService) ValidateToken(token string) (*model.User, error) {
	// Validate token
	claims, err := s.jwtManager.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	// Get user
	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

func (s *authService) Logout(refreshToken string) error {
	return s.refreshTokenRepo.DeleteByToken(refreshToken)
}

func (s *authService) generateTokens(user *model.User) (*model.AuthResponse, error) {
	// Generate access token
	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	// Save refresh token to database
	tokenRecord := &model.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(s.jwtManager.GetRefreshExpiry()),
	}

	if err := s.refreshTokenRepo.Create(tokenRecord); err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.jwtManager.GetAccessExpiry().Seconds()),
		User:         user,
	}, nil
}
