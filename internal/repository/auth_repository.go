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

func NewAuthRepository(db *sqlx.DB) AuthRepository {
	return &authRepository{db: db}
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

func (r *authRepository) GetByEmail(email string) (*domain.Auth, error) {
	query := `
		SELECT id, email, password_hash, role, status, created_at, updated_at
		FROM auth
		WHERE email = $1
	`

	var auth domain.Auth
	err := r.db.Get(&auth, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get auth by email: %w", err)
	}

	return &auth, nil
}

func (r *authRepository) Update(auth *domain.Auth) error {
	query := `
		UPDATE auth
		SET email = $1, role = $2, status = $3, updated_at = $4
		WHERE id = $5
	`

	auth.UpdatedAt = time.Now()

	result, err := r.db.Exec(
		query,
		auth.Email,
		auth.Role,
		auth.Status,
		auth.UpdatedAt,
		auth.ID,
	)
	if err != nil {
		if IsDuplicateKeyError(err) {
			return ErrEmailAlreadyExists
		}
		return fmt.Errorf("failed to update auth: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *authRepository) UpdateStatus(id uuid.UUID, status domain.UserStatus) error {
	query := `
		UPDATE auth
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.Exec(query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update auth status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *authRepository) UpdatePassword(id uuid.UUID, passwordHash string) error {
	query := `
		UPDATE auth
		SET password_hash = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.Exec(query, passwordHash, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *authRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM auth WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete auth: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *authRepository) ExistsByEmail(email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM auth WHERE email = $1)`

	var exists bool
	err := r.db.Get(&exists, query, email)
	if err != nil {
		return false, fmt.Errorf("failed to check if email exists: %w", err)
	}

	return exists, nil
}
