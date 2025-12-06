package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hscHeric/go-potential-api/internal/domain"
)

// RequireRole verifica se o usuário tem uma das roles permitidas
func RequireRole(allowedRoles ...domain.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, err := GetRole(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "não autorizado",
			})
			c.Abort()
			return
		}

		// Verificar se a role do usuário está nas permitidas
		hasPermission := false
		for _, allowedRole := range allowedRoles {
			if role == string(allowedRole) {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "você não tem permissão para acessar esse recurso",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin verifica se o usuário é admin
func RequireAdmin() gin.HandlerFunc {
	return RequireRole(domain.RoleAdmin)
}

// RequireTeacher verifica se o usuário é professor
func RequireTeacher() gin.HandlerFunc {
	return RequireRole(domain.RoleTeacher)
}

// RequireStudent verifica se o usuário é aluno
func RequireStudent() gin.HandlerFunc {
	return RequireRole(domain.RoleStudent)
}

// RequireTeacherOrAdmin verifica se o usuário é professor ou admin
func RequireTeacherOrAdmin() gin.HandlerFunc {
	return RequireRole(domain.RoleTeacher, domain.RoleAdmin)
}

// RequireStudentOrTeacher verifica se o usuário é aluno ou professor
func RequireStudentOrTeacher() gin.HandlerFunc {
	return RequireRole(domain.RoleStudent, domain.RoleTeacher)
}
