package handlers

import (
	"net/http"

	"golive/profile/internal/application/profile"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func (h *ProfileHandler) getByLogin(c *gin.Context) {
	login := c.Param("login")
	out, err := h.GetByLogin.Execute(c.Request.Context(), profile.GetByLoginInput{Login: login})
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, toProfileResponse(out))
}

