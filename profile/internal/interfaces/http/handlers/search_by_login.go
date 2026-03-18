package handlers

import (
	"net/http"
	"strconv"

	"golive/profile/internal/application/profile"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func (h *ProfileHandler) searchByLogin(c *gin.Context) {
	q := c.Query("login")
	limit := 10
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 50 {
			limit = n
		}
	}

	// exact match first
	var exact *profileResponse
	exactOut, err := h.GetByLogin.Execute(c.Request.Context(), profile.GetByLoginInput{Login: q})
	if err == nil {
		tmp := toProfileResponse(exactOut)
		exact = &tmp
		c.JSON(http.StatusOK, searchResponse{Exact: exact, Matches: []profileResponse{}})
		return
	}
	if err != nil && err != pgx.ErrNoRows {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": err.Error()})
		return
	}

	matches, err := h.SearchByLogin.Execute(c.Request.Context(), profile.SearchByLoginInput{
		Query: q,
		Limit: limit,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": err.Error()})
		return
	}
	resp := make([]profileResponse, 0, len(matches))
	for _, p := range matches {
		resp = append(resp, toProfileResponse(p))
	}
	c.JSON(http.StatusOK, searchResponse{Exact: nil, Matches: resp})
}

