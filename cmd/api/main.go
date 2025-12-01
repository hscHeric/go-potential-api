package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/hscHeric/go-potential-api/docs"
	"github.com/hscHeric/go-potential-api/internal/config"
	"github.com/hscHeric/go-potential-api/internal/database"
	"github.com/hscHeric/go-potential-api/internal/handler"
	"github.com/hscHeric/go-potential-api/internal/middleware"
	"github.com/hscHeric/go-potential-api/internal/repository"
	"github.com/hscHeric/go-potential-api/internal/service"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Potential API
// @version 1.0
// @description API para sistema de gestão de escola de idiomas
// @termsOfService http://swagger.io/terms/

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Connect to database
	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, cfg)
	userService := service.NewUserService(userRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)

	// Setup Gin
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Middlewares
	router.Use(middleware.CORSMiddleware())

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"env":    cfg.Server.Env,
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(cfg))
		{
			protected.GET("/me", authHandler.Me)
		}

		// Admin only routes
		admin := v1.Group("/users")
		admin.Use(middleware.AuthMiddleware(cfg))
		admin.Use(middleware.RequireRole("admin"))
		{
			admin.POST("", userHandler.CreateUser)
			admin.GET("", userHandler.ListUsers)
			admin.GET("/:id", userHandler.GetUser)
			admin.PATCH("/:id/status", userHandler.UpdateUserStatus)
			admin.DELETE("/:id", userHandler.DeleteUser)
		}
	}

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	log.Printf("Swagger docs available at http://localhost:%s/swagger/index.html", cfg.Server.Port)

	if err := router.Run(addr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
