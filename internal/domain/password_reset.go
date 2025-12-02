package domain

import (
	"time"

	"gorm.io/gorm"
)

type PasswordReset struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	AuthID    uint           `gorm:"not null;index" json:"auth_id"`
	Token     string         `gorm:"uniqueIndex;not null" json:"-"`
	ExpiresAt time.Time      `gorm:"not null" json:"expires_at"`
	UsedAt    *time.Time     `json:"used_at,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type RequestPasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}
