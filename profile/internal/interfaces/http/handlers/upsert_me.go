package handlers

import (
	"net/http"
	"time"

	"golive/profile/internal/application/profile"
	"golive/profile/internal/interfaces/http/middleware"

	"github.com/gin-gonic/gin"
)

func (h *ProfileHandler) upsertMe(c *gin.Context) {
	userID, _ := c.Get(middleware.ContextUserID)
	uid, _ := userID.(string)
	loginV, _ := c.Get(middleware.ContextLogin)
	login, _ := loginV.(string)
	if uid == "" || login == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req upsertMeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": err.Error()})
		return
	}

	out, err := h.UpsertMe.Execute(c.Request.Context(), profile.UpsertMeInput{
		UserID:      uid,
		Login:       login,
		DisplayName: req.DisplayName,
		AvatarURL:   req.AvatarURL,
		Bio:         req.Bio,
		Now:         time.Now().UTC(),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, toProfileResponse(out))
}

