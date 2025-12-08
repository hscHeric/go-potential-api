// Package router define as rotas da aplicação
package router

import (
	"github.com/gin-gonic/gin"
	"github.com/hscHeric/go-potential-api/internal/handler"
	"github.com/hscHeric/go-potential-api/internal/middleware"
	"github.com/hscHeric/go-potential-api/pkg/jwt"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type RouterConfig struct {
	AuthHandler *handler.AuthHandler
	UserHandler *handler.UserHandler
	JWTService  *jwt.Service
}

func SetupRouter(cfg RouterConfig) *gin.Engine {
	router := gin.Default()

	// Middlewares globais
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.ErrorHandlerMiddleware())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1
	v1 := router.Group("/api/v1")
	{
		// Auth routes (públicas)
		auth := v1.Group("/auth")
		{
			auth.POST("/login", cfg.AuthHandler.Login)
			auth.POST("/complete-registration", cfg.AuthHandler.CompleteRegistration)
			auth.GET("/validate-activation-token", cfg.AuthHandler.ValidateActivationToken)
			auth.POST("/request-password-reset", cfg.AuthHandler.RequestPasswordReset)
			auth.POST("/reset-password", cfg.AuthHandler.ResetPassword)
		}

		// Invitations (apenas admin)
		invitations := v1.Group("/invitations")
		invitations.Use(middleware.AuthMiddleware(cfg.JWTService))
		invitations.Use(middleware.RequireAdmin())
		{
			invitations.POST("", cfg.AuthHandler.CreateInvitation)
		}

		// User routes (autenticadas)
		users := v1.Group("/users")
		users.Use(middleware.AuthMiddleware(cfg.JWTService))
		{
			users.GET("/me", cfg.UserHandler.GetProfile)
			users.PUT("/me", cfg.UserHandler.UpdateProfile)
		}
	}

	return router
}
