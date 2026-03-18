package handlers

import "github.com/gin-gonic/gin"

func (h *AuthHandler) RegisterRoutes(rg *gin.RouterGroup, logoutMiddleware ...gin.HandlerFunc) {
	rg.POST("/register", h.register)
	rg.POST("/login", h.login)
	rg.POST("/refresh", h.refresh)
	if len(logoutMiddleware) > 0 {
		rg.POST("/logout", append(logoutMiddleware, h.logout)...)
		return
	}
	rg.POST("/logout", h.logout)
}

