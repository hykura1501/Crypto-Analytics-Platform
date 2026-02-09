package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/crypto-platform/auth-service/internal/middleware"
	"github.com/crypto-platform/auth-service/internal/model"
	"github.com/crypto-platform/auth-service/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body model.RegisterRequest true "Register request"
// @Success 201 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 409 {object} model.ErrorResponse
// @Router /register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	user, err := h.authService.Register(&req)
	if err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			c.JSON(http.StatusConflict, model.ErrorResponse{
				Error:   "User already exists",
				Message: "A user with this email already exists",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "Registration failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, model.SuccessResponse{
		Success: true,
		Message: "User registered successfully",
		Data:    user,
	})
}

// Login godoc
// @Summary Login user
// @Description Authenticate user and return JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body model.LoginRequest true "Login request"
// @Success 200 {object} model.AuthResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Router /login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	authResponse, err := h.authService.Login(&req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, model.ErrorResponse{
				Error:   "Invalid credentials",
				Message: "Email or password is incorrect",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "Login failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, authResponse)
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Get new access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body model.RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} model.AuthResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Router /refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req model.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	authResponse, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error:   "Token refresh failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, authResponse)
}

// Validate godoc
// @Summary Validate access token
// @Description Validate the current access token and return user info
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} model.SuccessResponse
// @Failure 401 {object} model.ErrorResponse
// @Router /validate [get]
func (h *AuthHandler) Validate(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error:   "Unauthorized",
			Message: "User ID not found in token",
		})
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse{
		Success: true,
		Message: "Token is valid",
		Data: gin.H{
			"user_id": userID,
			"email":   c.GetString("email"),
		},
	})
}

// Logout godoc
// @Summary Logout user
// @Description Invalidate refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body model.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Router /logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req model.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	if err := h.authService.Logout(req.RefreshToken); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, model.SuccessResponse{
				Success: true,
				Message: "Logged out successfully",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "Logout failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse{
		Success: true,
		Message: "Logged out successfully",
	})
}

// Me godoc
// @Summary Get current user
// @Description Get information about the authenticated user
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} model.User
// @Failure 401 {object} model.ErrorResponse
// @Router /me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if len(token) > 7 {
		token = token[7:] // Remove "Bearer " prefix
	}

	user, err := h.authService.ValidateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error:   "Unauthorized",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// ListUsers returns all users (ADMIN only)
func (h *AuthHandler) ListUsers(c *gin.Context) {
	users, err := h.authService.ListUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "Failed to list users",
			Message: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, model.SuccessResponse{
		Success: true,
		Data:    users,
	})
}

// UpdateUserRole updates a user's role to NORMAL or VIP (ADMIN only)
func (h *AuthHandler) UpdateUserRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "Invalid user ID",
			Message: "user id must be a positive integer",
		})
		return
	}

	var req model.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	user, err := h.authService.UpdateUserRole(uint(id), req.Role)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "User not found",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "Update failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse{
		Success: true,
		Message: "Role updated successfully",
		Data:    user,
	})
}
