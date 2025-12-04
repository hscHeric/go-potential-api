// Package repository é responsável pela persistencia dos dados
package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

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

func (r *authRepository) Create(auth *domain.Auth) error {
	query := `
		INSERT INTO auth (id, email, password_hash, role, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	auth.ID = uuid.New()
	auth.CreatedAt = time.Now()
	auth.UpdatedAt = time.Now()

	_, err := r.db.Exec(
		query,
		auth.ID,
		auth.Email,
		auth.PasswordHash,
		auth.Role,
		auth.Status,
		auth.CreatedAt,
		auth.UpdatedAt,
	)
	if err != nil {
		if IsDuplicateKeyError(err) {
			return ErrEmailAlreadyExists
		}

		return fmt.Errorf("falha ao criar entidade auth: %w", err)
	}

	return nil
}

func (r *authRepository) GetByID(id uuid.UUID) (*domain.Auth, error) {
	query := `
		SELECT id, email, password_hash, role, status, created_at, updated_at
		FROM auth
		WHERE id = $1
	`

	var auth domain.Auth
	err := r.db.Get(&auth, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("falha ao procurar entidade auth pelo ID: %w", err)
	}

	return &auth, nil
}
