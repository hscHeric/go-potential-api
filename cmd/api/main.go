package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/hscHeric/go-potential-api/docs"
	_ "github.com/lib/pq"

	"github.com/hscHeric/go-potential-api/internal/config"
	"github.com/hscHeric/go-potential-api/internal/database"
	"github.com/hscHeric/go-potential-api/internal/handler"
	"github.com/hscHeric/go-potential-api/internal/repository"
	"github.com/hscHeric/go-potential-api/internal/router"
	"github.com/hscHeric/go-potential-api/internal/service"
	"github.com/hscHeric/go-potential-api/internal/storage"
	"github.com/hscHeric/go-potential-api/pkg/email"
	"github.com/hscHeric/go-potential-api/pkg/jwt"
)

// @title Potential Idiomas API
// @version 1.0
// @description API para gerenciamento de escola de idiomas
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.potential-idiomas.com/support
// @contact.email support@potential-idiomas.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Carregar configurações
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Conectar ao banco de dados
	dbCfg := database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
	}

	db, err := database.Connect(dbCfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close(db)

	// Inicializar repositories
	authRepo := repository.NewAuthRepository(db)
	userRepo := repository.NewUserRepository(db)
	activationTokenRepo := repository.NewActivationTokenRepository(db)
	passwordResetRepo := repository.NewPasswordResetTokenRepository(db)
	documentRepo := repository.NewDocumentRepository(db)

	// Inicializar services
	jwtService := jwt.NewService(cfg.JWT.Secret, cfg.JWT.ExpirationHours)
	emailService := email.NewService(
		cfg.Email.SMTPHost,
		cfg.Email.SMTPPort,
		cfg.Email.SMTPUsername,
		cfg.Email.SMTPPassword,
		cfg.Email.SMTPFrom,
		cfg.Tokens.FrontendURL,
	)

	// Inicializar storage (S3/MinIO)
	s3Storage, err := storage.NewS3Storage(storage.S3Config{
		Endpoint:  cfg.S3.Endpoint,
		Region:    cfg.S3.Region,
		AccessKey: cfg.S3.AccessKey,
		SecretKey: cfg.S3.SecretKey,
		Bucket:    cfg.S3.Bucket,
		UseSSL:    cfg.S3.UseSSL,
	})
	if err != nil {
		log.Fatalf("Failed to initialize S3 storage: %v", err)
	}

	authService := service.NewAuthService(
		authRepo,
		userRepo,
		activationTokenRepo,
		passwordResetRepo,
		jwtService,
		emailService,
		cfg.GetActivationTokenExpiration(),
		cfg.GetPasswordResetTokenExpiration(),
	)

	userService := service.NewUserService(authRepo, userRepo)

	documentService := service.NewDocumentService(
		userRepo,
		documentRepo,
		s3Storage,
	)

	// Inicializar handlers
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	documentHandler := handler.NewDocumentHandler(documentService)

	// Configurar router
	routerCfg := router.RouterConfig{
		AuthHandler:     authHandler,
		UserHandler:     userHandler,
		DocumentHandler: documentHandler,
		JWTService:      jwtService,
	}

	r := router.SetupRouter(routerCfg)

	// Iniciar servidor
	port := cfg.Server.Port
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s...", port)
	log.Printf("Environment: %s", cfg.Server.AppEnv)
	log.Printf("Swagger docs available at: http://localhost:%s/swagger/index.html", port)

	// Graceful shutdown
	go func() {
		if err := r.Run(":" + port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Aguardar sinal de interrupção
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	log.Println("Server stopped")
}
