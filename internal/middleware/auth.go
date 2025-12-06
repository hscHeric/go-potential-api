// Package middleware adicionar definição
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hscHeric/go-potential-api/pkg/jwt"
)

const (
	AuthorizationHeader = "Authorization"
	AuthPayloadKey      = "auth_payload"
)

// AuthMiddleware verifica se o usuário está autenticado via JWT
func AuthMiddleware(jwtService *jwt.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Extrair token do header "Bearer <token>"
		fields := strings.Fields(authHeader)
		if len(fields) < 2 || strings.ToLower(fields[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		token := fields[1]

		// Validar token
		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			if err == jwt.ErrExpiredToken {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Token has expired",
				})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Invalid token",
				})
			}
			c.Abort()
			return
		}

		// Adicionar informações do usuário no contexto
		c.Set(AuthPayloadKey, claims)
		c.Set("auth_id", claims.AuthID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// GetAuthID extrai o auth_id do contexto
func GetAuthID(c *gin.Context) (uuid.UUID, error) {
	authID, exists := c.Get("auth_id")
	if !exists {
		return uuid.Nil, jwt.ErrInvalidToken
	}

	id, ok := authID.(uuid.UUID)
	if !ok {
		return uuid.Nil, jwt.ErrInvalidToken
	}

	return id, nil
}

// GetEmail extrai o email do contexto
func GetEmail(c *gin.Context) (string, error) {
	email, exists := c.Get("email")
	if !exists {
		return "", jwt.ErrInvalidToken
	}

	emailStr, ok := email.(string)
	if !ok {
		return "", jwt.ErrInvalidToken
	}

	return emailStr, nil
}

// GetRole extrai o role do contexto
func GetRole(c *gin.Context) (string, error) {
	role, exists := c.Get("role")
	if !exists {
		return "", jwt.ErrInvalidToken
	}

	roleStr, ok := role.(string)
	if !ok {
		return "", jwt.ErrInvalidToken
	}

	return roleStr, nil
}

// GetAuthPayload extrai as claims completas do contexto
func GetAuthPayload(c *gin.Context) (*jwt.Claims, error) {
	payload, exists := c.Get(AuthPayloadKey)
	if !exists {
		return nil, jwt.ErrInvalidToken
	}

	claims, ok := payload.(*jwt.Claims)
	if !ok {
		return nil, jwt.ErrInvalidToken
	}

	return claims, nil
}
