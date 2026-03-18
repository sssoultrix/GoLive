package handlers

import (
	"net/http"
	"time"

	"golive/auth/internal/application/auth"

	"github.com/gin-gonic/gin"
)

func (h *AuthHandler) refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": err.Error()})
		return
	}

	now := time.Now().UTC()
	out, err := h.Refresh.Execute(c.Request.Context(), auth.RefreshInput{
		RefreshToken: req.RefreshToken,
		Now:          now,
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_refresh_token"})
		return
	}

	c.JSON(http.StatusOK, authResponse{
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
		ExpiresIn:    int64(time.Until(out.ExpiresAt).Seconds()),
		TokenType:    "Bearer",
	})
}

