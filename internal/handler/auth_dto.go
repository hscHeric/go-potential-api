package handler

import (
	"time"

	"github.com/hscHeric/go-potential-api/internal/domain"
)

// CreateInvitationRequest representa o payload para criar convite
type CreateInvitationRequest struct {
	Email string      `json:"email" binding:"required,email"`
	Role  domain.Role `json:"role" binding:"required,oneof=admin teacher student"`
}

// CompleteRegistrationRequest representa o payload para completar registro
type CompleteRegistrationRequest struct {
	Token     string         `json:"token" binding:"required"`
	FullName  string         `json:"full_name" binding:"required,min=3"`
	CPF       string         `json:"cpf" binding:"required,len=11"`
	BirthDate time.Time      `json:"birth_date" binding:"required"`
	Address   domain.Address `json:"address" binding:"required"`
	Contact   domain.Contact `json:"contact" binding:"required"`
	Password  string         `json:"password" binding:"required,min=8"`
}

// LoginRequest representa o payload de login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RequestPasswordResetRequest representa o payload para solicitar reset de senha
type RequestPasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest representa o payload para resetar senha
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// ValidateTokenResponse representa a resposta de validação de token
type ValidateTokenResponse struct {
	Valid bool   `json:"valid"`
	Email string `json:"email"`
	Role  string `json:"role"`
}
