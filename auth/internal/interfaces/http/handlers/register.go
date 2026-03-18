package handlers

import (
	"net/http"
	"time"

	"golive/auth/internal/application/auth"

	"github.com/gin-gonic/gin"
)

func (h *AuthHandler) register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": err.Error()})
		return
	}

	now := time.Now().UTC()
	ua := c.GetHeader("User-Agent")
	var uaPtr *string
	if ua != "" {
		uaPtr = &ua
	}
	ip := c.ClientIP()
	var ipPtr *string
	if ip != "" {
		ipPtr = &ip
	}

	out, err := h.Register.Execute(c.Request.Context(), auth.RegisterInput{
		Login:     req.Login,
		Password:  req.Password,
		UserAgent: uaPtr,
		IP:        ipPtr,
		Now:       now,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "register_failed", "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, authResponse{
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
		ExpiresIn:    int64(time.Until(out.ExpiresAt).Seconds()),
		TokenType:    "Bearer",
	})
}

