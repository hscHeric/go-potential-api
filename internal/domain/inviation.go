package domain

import (
	"time"

	"gorm.io/gorm"
)

type InvitationStatus string

const (
	InvitationStatusPending   InvitationStatus = "pending"
	InvitationStatusAccepted  InvitationStatus = "accepted"
	InvitationStatusExpired   InvitationStatus = "expired"
	InvitationStatusCancelled InvitationStatus = "cancelled"
)

type Invitation struct {
	ID         uint             `gorm:"primarykey" json:"id"`
	Email      string           `gorm:"not null;index" json:"email"`
	Role       UserRole         `gorm:"type:varchar(20);not null" json:"role"`
	Token      string           `gorm:"uniqueIndex;not null" json:"-"`
	Status     InvitationStatus `gorm:"type:varchar(20);default:'pending'" json:"status"`
	ExpiresAt  time.Time        `gorm:"not null" json:"expires_at"`
	AcceptedAt *time.Time       `json:"accepted_at,omitempty"`
	CreatedBy  uint             `gorm:"not null" json:"created_by"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
	DeletedAt  gorm.DeletedAt   `gorm:"index" json:"-"`
}

type CreateInvitationRequest struct {
	Email string   `json:"email" binding:"required,email"`
	Role  UserRole `json:"role" binding:"required,oneof=admin teacher student"`
}

type ValidateInvitationResponse struct {
	Email     string    `json:"email"`
	Role      UserRole  `json:"role"`
	ExpiresAt time.Time `json:"expires_at"`
	IsValid   bool      `json:"is_valid"`
}
