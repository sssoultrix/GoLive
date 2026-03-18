package router

import (
	"net/http"

	"golive/profile/internal/interfaces/http/handlers"
	"golive/profile/internal/interfaces/http/middleware"

	"github.com/gin-gonic/gin"
)

type Deps struct {
	Profile      *handlers.ProfileHandler
	ProfileGuard gin.HandlerFunc
}

func New(d Deps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/healthz", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	api := r.Group("/api/v1")
	profileGroup := api.Group("/profile")

	if d.Profile != nil {
		if d.ProfileGuard != nil {
			d.Profile.RegisterRoutes(profileGroup, d.ProfileGuard)
		} else {
			d.Profile.RegisterRoutes(profileGroup)
		}
	}

	// Keep middleware package referenced even if Profile is nil in tests.
	_ = middleware.ContextUserID

	return r
}

