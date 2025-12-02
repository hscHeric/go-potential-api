package domain

import (
	"time"

	"gorm.io/gorm"
)

type UserRole string

const (
	RoleAdmin   UserRole = "admin"
	RoleTeacher UserRole = "teacher"
	RoleStudent UserRole = "student"
)

// Tabela de credenciais de autenticação
type Auth struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	Email        string         `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string         `gorm:"not null" json:"-"`
	Role         UserRole       `gorm:"type:varchar(20);not null" json:"role"`
	IsActive     bool           `gorm:"default:false" json:"is_active"` // Inativo até completar cadastro
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relacionamento com a tabela de usuários
	User *User `gorm:"foreignKey:AuthID" json:"user,omitempty"`
}

// request do login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// Formato da response do login
type LoginResponse struct {
	Token string   `json:"token"`
	User  UserInfo `json:"user"`
}

// Informações básicas do usuário retornadas no login
type UserInfo struct {
	ID                  uint     `json:"id"`
	Email               string   `json:"email"`
	Role                UserRole `json:"role"`
	IsActive            bool     `json:"is_active"`
	HasCompletedProfile bool     `json:"has_completed_profile"`
}
