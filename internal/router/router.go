// Package router define as rotas da aplicação
package router

import (
	"github.com/gin-gonic/gin"
	"github.com/hscHeric/go-potential-api/internal/domain"
	"github.com/hscHeric/go-potential-api/internal/handler"
	"github.com/hscHeric/go-potential-api/internal/middleware"
	"github.com/hscHeric/go-potential-api/pkg/jwt"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type RouterConfig struct {
	AuthHandler     *handler.AuthHandler
	UserHandler     *handler.UserHandler
	TimeSlotHandler *handler.TimeSlotHandler
	ClassHandler    *handler.ClassHandler
	JWTService      *jwt.Service
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

		// TimeSlot routes (Professor)
		timeSlots := v1.Group("/time-slots")
		timeSlots.Use(middleware.AuthMiddleware(cfg.JWTService))
		{
			// Rotas para professores criarem/gerenciarem seus horários
			timeSlots.POST("", middleware.RequireRole(domain.RoleTeacher), cfg.TimeSlotHandler.CreateTimeSlot)
			timeSlots.GET("/me", middleware.RequireRole(domain.RoleTeacher), cfg.TimeSlotHandler.GetMyTimeSlots)
			timeSlots.PUT("/:id", middleware.RequireRole(domain.RoleTeacher), cfg.TimeSlotHandler.UpdateTimeSlot)
			timeSlots.DELETE("/:id", middleware.RequireRole(domain.RoleTeacher), cfg.TimeSlotHandler.DeleteTimeSlot)
			timeSlots.PATCH("/:id/toggle", middleware.RequireRole(domain.RoleTeacher), cfg.TimeSlotHandler.ToggleTimeSlotAvailability)

			// Rotas para alunos visualizarem horários disponíveis
			timeSlots.GET("/teacher/:teacher_id", cfg.TimeSlotHandler.GetTeacherTimeSlots)
			timeSlots.GET("/teacher/:teacher_id/available", cfg.TimeSlotHandler.GetAvailableSlots)
		}

		// Class routes
		classes := v1.Group("/classes")
		classes.Use(middleware.AuthMiddleware(cfg.JWTService))
		{
			// Criar aula (Professor ou Admin)
			classes.POST("", middleware.RequireRole(domain.RoleTeacher, domain.RoleAdmin), cfg.ClassHandler.CreateClass)

			// Ver minhas aulas (Professor vê as que ensina, Aluno vê as que está matriculado)
			classes.GET("/me", cfg.ClassHandler.GetMyClasses)

			// Ver detalhes de uma aula
			classes.GET("/:id", cfg.ClassHandler.GetClass)

			// Atualizar aula (Professor ou Admin)
			classes.PUT("/:id", middleware.RequireRole(domain.RoleTeacher, domain.RoleAdmin), cfg.ClassHandler.UpdateClass)

			// Cancelar aula (Professor ou Admin)
			classes.PATCH("/:id/cancel", middleware.RequireRole(domain.RoleTeacher, domain.RoleAdmin), cfg.ClassHandler.CancelClass)

			// Gerenciar alunos na aula
			classes.POST("/:id/students", middleware.RequireRole(domain.RoleTeacher, domain.RoleAdmin, domain.RoleStudent), cfg.ClassHandler.AddStudentToClass)
			classes.DELETE("/:id/students/:student_id", middleware.RequireRole(domain.RoleTeacher, domain.RoleAdmin), cfg.ClassHandler.RemoveStudentFromClass)

			// Marcar presença (apenas Professor)
			classes.PATCH("/:id/students/:student_id/attendance", middleware.RequireRole(domain.RoleTeacher), cfg.ClassHandler.MarkAttendance)
		}
	}

	return router
}
