// Package service implementa as regras de negócio da aplicação
package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/hscHeric/go-potential-api/internal/domain"
	"github.com/hscHeric/go-potential-api/internal/repository"
	"github.com/hscHeric/go-potential-api/pkg/email"
	"github.com/hscHeric/go-potential-api/pkg/jwt"
	"github.com/hscHeric/go-potential-api/pkg/token"
)

var (
	ErrInvalidCredentials   = errors.New("credenciais inválidas")
	ErrUserNotActive        = errors.New("usuário não está ativo")
	ErrTokenExpired         = errors.New("o token expirou")
	ErrTokenAlreadyUsed     = errors.New("o token já foi utilizado")
	ErrUserAlreadyCompleted = errors.New("o registro do usuário já foi concluído")
)

type AuthService interface {
	// Admin cria convite inicial de finindo se o email e se é aluno ou professor
	CreateInvitation(email string, role domain.Role) error

	// Usuário completa o cadastro usando um token de ativação de conta
	CompleteRegistration(activationToken string, userData *CompleteRegistrationInput) error

	// Login
	Login(email, password string) (*LoginResponse, error)

	// RequestPasswordReset solicita envio de email para recuperação de senha
	RequestPasswordReset(email string) error

	// ResetPassword altera a senha do usuário
	ResetPassword(resetToken, newPassword string) error

	// Validação de token
	ValidateActivationToken(token string) (*domain.Auth, error)
	ValidatePasswordResetToken(token string) (*domain.Auth, error)
}

type CompleteRegistrationInput struct {
	FullName  string         `json:"full_name" binding:"required,min=3"`
	CPF       string         `json:"cpf" binding:"required,len=11"`
	BirthDate time.Time      `json:"birth_date" binding:"required"`
	Address   domain.Address `json:"address" binding:"required"`
	Contact   domain.Contact `json:"contact" binding:"required"`
	Password  string         `json:"password" binding:"required,min=8"`
}

type LoginResponse struct {
	Token string               `json:"token"`
	User  *domain.UserWithAuth `json:"user"`
}

type authService struct {
	authRepo                repository.AuthRepository
	userRepo                repository.UserRepository
	activationTokenRepo     repository.ActivationTokenRepository
	passwordResetRepo       repository.PasswordResetTokenRepository
	jwtService              *jwt.Service
	emailService            *email.Service
	activationExpiration    time.Duration
	passwordResetExpiration time.Duration
}

func NewAuthService(
	authRepo repository.AuthRepository,
	userRepo repository.UserRepository,
	activationTokenRepo repository.ActivationTokenRepository,
	passwordResetRepo repository.PasswordResetTokenRepository,
	jwtService *jwt.Service,
	emailService *email.Service,
	activationExpiration time.Duration,
	passwordResetExpiration time.Duration,
) AuthService {
	return &authService{
		authRepo:                authRepo,
		userRepo:                userRepo,
		activationTokenRepo:     activationTokenRepo,
		passwordResetRepo:       passwordResetRepo,
		jwtService:              jwtService,
		emailService:            emailService,
		activationExpiration:    activationExpiration,
		passwordResetExpiration: passwordResetExpiration,
	}
}

func (s *authService) CreateInvitation(email string, role domain.Role) error {
	exists, err := s.authRepo.ExistsByEmail(email)
	if err != nil {
		return fmt.Errorf("falha ao verificar o e-mail: %w", err)
	}

	if exists {
		return repository.ErrEmailAlreadyExists
	}

	auth := &domain.Auth{
		Email:        email,
		PasswordHash: "", // Isso é definido no complete registration
		Role:         role,
		Status:       domain.StatusPending,
	}

	// Cria as credencias no banco de dados
	if err := s.authRepo.Create(auth); err != nil {
		return fmt.Errorf("falha ao criar credenciais (Auth): %w", err)
	}

	// Gera um token de ativação
	activationToken, err := token.GenerateActivationToken()
	if err != nil {
		return fmt.Errorf("falha ao gerar token de ativação: %w", err)
	}

	// Salva o token de ativação no DB
	tokenRecord := &domain.ActivationToken{
		AuthID:    auth.ID,
		Token:     activationToken,
		ExpiresAt: time.Now().Add(s.activationExpiration),
	}
	if err := s.activationTokenRepo.Create(tokenRecord); err != nil {
		return fmt.Errorf("falha ao salvar token de ativação: %w", err)
	}

	// Enviar email de ativação
	if err := s.emailService.SendActivationEmail(email, email, activationToken); err != nil {
		return fmt.Errorf("falha ao enviar e-mail de ativação: %w", err)
	}

	return nil
}

func (s *authService) CompleteRegistration(activationToken string, userData *CompleteRegistrationInput) error {
}
func (s *authService) Login(email, password string) (*LoginResponse, error)          {}
func (s *authService) RequestPasswordReset(email string) error                       {}
func (s *authService) ResetPassword(resetToken, newPassword string) error            {}
func (s *authService) ValidateActivationToken(token string) (*domain.Auth, error)    {}
func (s *authService) ValidatePasswordResetToken(token string) (*domain.Auth, error) {}
