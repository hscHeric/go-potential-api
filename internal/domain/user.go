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

type User struct {
	ID           uint           `gorm:"primarykey" json:"id" example:"1"`
	Email        string         `gorm:"uniqueIndex;not null" json:"email" example:"user@escola.com"`
	PasswordHash string         `gorm:"not null" json:"-"`
	Role         UserRole       `gorm:"type:varchar(20);not null" json:"role" example:"student"`
	IsActive     bool           `gorm:"default:true" json:"is_active" example:"true"`
	CreatedAt    time.Time      `json:"created_at" example:"2024-12-01T10:00:00Z"`
	UpdatedAt    time.Time      `json:"updated_at" example:"2024-12-01T10:00:00Z"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"admin@escola.com"`
	Password string `json:"password" binding:"required,min=6" example:"admin123456"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email" example:"aluno@escola.com"`
	Password string `json:"password" binding:"required,min=6" example:"123456"`
}

type CreateUserRequest struct {
	Email    string   `json:"email" binding:"required,email" example:"professor@escola.com"`
	Password string   `json:"password" binding:"required,min=6" example:"123456"`
	Role     UserRole `json:"role" binding:"required,oneof=teacher student" example:"teacher"`
}

type LoginResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User  User   `json:"user"`
}
