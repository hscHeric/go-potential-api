package middleware

import (
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hscHeric/go-potential-api/internal/config"
	"github.com/hscHeric/go-potential-api/pkg/response"
	"github.com/hscHeric/go-potential-api/pkg/utils"
)

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, 401, "Authorization header required")
			c.Abort()
			return
		}

		// Bearer token format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, 401, "Invalid authorization header format")
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := utils.ValidateToken(token, cfg.JWT.Secret)
		if err != nil {
			response.Error(c, 401, "Token invalido ou expirado")
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)

		c.Next()
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			response.Error(c, 401, "Unauthorized")
			c.Abort()
			return
		}

		roleStr := userRole.(string)
		hasRole := slices.Contains(roles, roleStr)

		if !hasRole {
			response.Error(c, 403, "Forbidden: permissão negada")
			c.Abort()
			return
		}

		c.Next()
	}
}
