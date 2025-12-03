// Package repository é responsável pela persistencia dos dados
package repository

import (
	"github.com/google/uuid"
	"github.com/hscHeric/go-potential-api/internal/domain"
	"github.com/jmoiron/sqlx"
)

type AuthRepository interface {
	Create(auth *domain.Auth) error
	GetByID(id uuid.UUID) (*domain.Auth, error)
	GetByEmail(email string) (*domain.Auth, error)
	Update(auth *domain.Auth) error
	UpdateStatus(id uuid.UUID, status domain.UserStatus) error
	UpdatePassword(id uuid.UUID, passwordHash string) error
	Delete(id uuid.UUID) error
	ExistsByEmail(email string) (bool, error)
}

type authRepository struct {
	db *sqlx.DB
}
