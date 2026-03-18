package handlers

import (
	"net/http"

	"golive/profile/internal/application/profile"
	"golive/profile/internal/interfaces/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func (h *ProfileHandler) getMe(c *gin.Context) {
	userID, _ := c.Get(middleware.ContextUserID)
	uid, _ := userID.(string)
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	out, err := h.GetMe.Execute(c.Request.Context(), profile.GetMeInput{UserID: uid})
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusOK, toProfileResponse(out))
}

