package handlers

import "github.com/gin-gonic/gin"

func (h *ProfileHandler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware ...gin.HandlerFunc) {
	// Public (static paths first)
	rg.GET("/search", h.searchByLogin)

	// Protected
	if len(authMiddleware) > 0 {
		rg.GET("/me", append(authMiddleware, h.getMe)...)
		rg.PUT("/me", append(authMiddleware, h.upsertMe)...)
		rg.GET("/:login", h.getByLogin)
		return
	}
	rg.GET("/me", h.getMe)
	rg.PUT("/me", h.upsertMe)
	rg.GET("/:login", h.getByLogin)
}

