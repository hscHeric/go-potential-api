// Package service implementa as regras de negócio da aplicação
package service

import (
	"errors"
	"time"

	"github.com/hscHeric/go-potential-api/internal/domain"
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
