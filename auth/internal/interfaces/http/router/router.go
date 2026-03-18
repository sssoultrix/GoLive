package router

import (
	"net/http"

	"golive/auth/internal/interfaces/http/handlers"
	"golive/auth/internal/interfaces/http/middleware"

	"github.com/gin-gonic/gin"
)

type Deps struct {
	Auth      *handlers.AuthHandler
	AuthGuard gin.HandlerFunc
}

func New(d Deps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/healthz", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	api := r.Group("/api/v1")
	authGroup := api.Group("/auth")

	if d.Auth != nil {
		if d.AuthGuard != nil {
			d.Auth.RegisterRoutes(authGroup, d.AuthGuard)
		} else {
			d.Auth.RegisterRoutes(authGroup)
		}
	}

	// Keep middleware package referenced even if Auth is nil in tests.
	_ = middleware.ContextUserID

	return r
}

