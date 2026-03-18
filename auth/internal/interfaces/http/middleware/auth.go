package middleware

import (
	"net/http"
	"strings"

	"golive/auth/internal/infrastructure/jwt_agent"

	"github.com/gin-gonic/gin"
)

const ContextUserID = "user_id"
const ContextLogin = "login"

func RequireBearerJWT(agent *jwt_agent.Agent) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if h == "" || !strings.HasPrefix(strings.ToLower(h), "bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		raw := strings.TrimSpace(h[len("Bearer "):])
		claims, err := agent.ParseAndVerifyAccessToken(raw)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Set(ContextUserID, claims.Subject)
		c.Set(ContextLogin, claims.Login)
		c.Next()
	}
}

