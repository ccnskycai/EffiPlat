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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	resp, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetMe retrieves the current user's information based on the JWT claims.
func (h *AuthHandler) GetMe(c *gin.Context) {
	// Retrieve claims set by the JWT middleware
	claimsValue, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user claims not found in context"})
		return
	}

	claims, ok := claimsValue.(*models.Claims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid claims format in context"})
		return
	}

	// Respond with user information from claims
	// Consider fetching full user details from service/repo if needed,
	// but for basic /me, claims might be sufficient.
	c.JSON(http.StatusOK, gin.H{
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
	c.JSON(http.StatusOK, gin.H{"message": "logout successful"})
}
