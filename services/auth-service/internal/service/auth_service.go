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
	ListUsers() ([]*model.User, error)
	UpdateUserRole(userID uint, role string) (*model.User, error)
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
		Role:         model.RoleNormal,
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
	// Check if refresh token exists in database first
	tokenRecord, err := s.refreshTokenRepo.FindByToken(refreshToken)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid refresh token")
		}
		return nil, err
	}

	// Check if token is expired in database
	if tokenRecord.ExpiresAt.Before(time.Now()) {
		s.refreshTokenRepo.DeleteByToken(refreshToken)
		return nil, errors.New("refresh token expired")
	}

	// Validate refresh token JWT structure
	claims, err := s.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		// Token is invalid or malformed, remove from database
		s.refreshTokenRepo.DeleteByToken(refreshToken)
		return nil, errors.New("invalid refresh token")
	}

	// Get user
	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Check if user is active
	if !user.IsActive {
		s.refreshTokenRepo.DeleteByToken(refreshToken)
		return nil, errors.New("user account is deactivated")
	}

	// Change: Do NOT delete old refresh token (disable Token Rotation) to prevent race conditions
	// s.refreshTokenRepo.DeleteByToken(refreshToken)

	// Generate only new Access Token
	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	// Return new Access Token but keep existing Refresh Token
	return &model.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken, // Reuse existing token
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.jwtManager.GetAccessExpiry().Seconds()),
		User:         user,
	}, nil
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

func (s *authService) ListUsers() ([]*model.User, error) {
	return s.userRepo.FindAll()
}

func (s *authService) UpdateUserRole(userID uint, role string) (*model.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	if user.Role == model.RoleAdmin {
		return nil, errors.New("cannot change role of admin user")
	}
	if role != model.RoleNormal && role != model.RoleVIP {
		return nil, errors.New("role must be NORMAL or VIP")
	}
	user.Role = role
	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *authService) generateTokens(user *model.User) (*model.AuthResponse, error) {
	role := user.Role
	if role == "" {
		role = model.RoleNormal
	}
	// Generate access token
	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email, role)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID, user.Email, role)
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
