package handlers

import (
	"net/http"

	"golive/auth/internal/application/auth"

	"github.com/gin-gonic/gin"
)

func (h *AuthHandler) logout(c *gin.Context) {
	var req logoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": err.Error()})
		return
	}
	if err := h.Logout.Execute(c.Request.Context(), auth.LogoutInput{RefreshToken: req.RefreshToken}); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "logout_failed", "message": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

