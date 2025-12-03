// Package domain define o formato dos dados
package domain

import (
	"time"

	"github.com/google/uuid"
)

// Role representa o papel do usuário no sistema
type Role string

const (
	RoleAdmin   Role = "admin"
	RoleTeacher Role = "teacher"
	RoleStudent Role = "student"
)

// UserStatus representa o status de ativação do usuário
type UserStatus string

const (
	StatusPending  UserStatus = "pending"  // Aguardando ativação
	StatusActive   UserStatus = "active"   // Ativo
	StatusInactive UserStatus = "inactive" // Desativado
)

// Auth trata-se da entidade da tabela de credenciais
type Auth struct {
	ID           uuid.UUID  `db:"id" json:"id"`
	Email        string     `db:"email" json:"email" binding:"required,email"`
	PasswordHash string     `db:"password_hash" json:"-"`
	Role         Role       `db:"role" json:"role" binding:"required,oneof=admin teacher student"`
	Status       UserStatus `db:"status" json:"status"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
}

// Address são as informações de endereço do aluno, aqui deixei como binding required
type Address struct {
	Street     string `json:"street" binding:"required"`
	Number     string `json:"number" binding:"required"`
	Complement string `json:"complement"`
	District   string `json:"district" binding:"required"`
	City       string `json:"city" binding:"required"`
	State      string `json:"state" binding:"required,len=2"`
	ZipCode    string `json:"zip_code" binding:"required"`
	Country    string `json:"country" binding:"required"`
}

// Contact representa os contatos (armazenado como JSONB)
type Contact struct {
	Phone       string `json:"phone"`
	MobilePhone string `json:"mobile_phone" binding:"required"`
	WhatsApp    string `json:"whatsapp"`
}

// User representa as informações pessoais do usuário
type User struct {
	ID         uuid.UUID `db:"id" json:"id"`
	AuthID     uuid.UUID `db:"auth_id" json:"auth_id"`
	FullName   string    `db:"full_name" json:"full_name" binding:"required,min=3"`
	CPF        string    `db:"cpf" json:"cpf" binding:"required,len=11"`
	BirthDate  time.Time `db:"birth_date" json:"birth_date" binding:"required"`
	Address    Address   `db:"address" json:"address" binding:"required"`
	Contact    Contact   `db:"contact" json:"contact" binding:"required"`
	ProfilePic *string   `db:"profile_pic" json:"profile_pic"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

// ActivationToken representa o token de ativação de conta
type ActivationToken struct {
	ID        uuid.UUID  `db:"id" json:"id"`
	AuthID    uuid.UUID  `db:"auth_id" json:"auth_id"`
	Token     string     `db:"token" json:"token"`
	ExpiresAt time.Time  `db:"expires_at" json:"expires_at"`
	UsedAt    *time.Time `db:"used_at" json:"used_at"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
}

// PasswordResetToken representa o token de recuperação de senha
type PasswordResetToken struct {
	ID        uuid.UUID  `db:"id" json:"id"`
	AuthID    uuid.UUID  `db:"auth_id" json:"auth_id"`
	Token     string     `db:"token" json:"token"`
	ExpiresAt time.Time  `db:"expires_at" json:"expires_at"`
	UsedAt    *time.Time `db:"used_at" json:"used_at"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
}

// UserWithAuth combina Auth e User para respostas completas
type UserWithAuth struct {
	Auth
	User *User `json:"user,omitempty"`
}
