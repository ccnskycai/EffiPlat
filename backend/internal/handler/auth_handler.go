package handler

import (
	"EffiPlat/backend/internal/models"
	"EffiPlat/backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}
	resp, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		RespondWithError(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}
	RespondWithSuccess(c, http.StatusOK, "Login successful", resp)
}

// GetMe retrieves the current user's information based on the JWT claims.
func (h *AuthHandler) GetMe(c *gin.Context) {
	// Retrieve claims set by the JWT middleware
	claimsValue, exists := c.Get("user")
	if !exists {
		RespondWithError(c, http.StatusInternalServerError, "User claims not found in context")
		return
	}

	claims, ok := claimsValue.(*models.Claims)
	if !ok {
		RespondWithError(c, http.StatusInternalServerError, "Invalid claims format in context")
		return
	}

	// Respond with user information from claims
	RespondWithSuccess(c, http.StatusOK, "User details retrieved successfully", gin.H{
		"id":    claims.UserID,
		"name":  claims.Name,
		"email": claims.Email,
	})
}

// Logout handles the user logout request.
// In a stateless JWT setup, this often just means the client discards the token.
// Server-side action (like blacklisting) is optional and more complex.
func (h *AuthHandler) Logout(c *gin.Context) {
	// Optionally: Add token to a blacklist here if implementing server-side invalidation.
	RespondWithSuccess(c, http.StatusOK, "Logout successful", nil)
}
