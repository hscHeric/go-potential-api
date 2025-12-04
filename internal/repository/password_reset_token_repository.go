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

type PasswordResetTokenRepository interface {
	Create(token *domain.PasswordResetToken) error
	GetByToken(token string) (*domain.PasswordResetToken, error)
	GetByAuthID(authID uuid.UUID) (*domain.PasswordResetToken, error)
	MarkAsUsed(id uuid.UUID) error
	DeleteExpired() error
	Delete(id uuid.UUID) error
}

type passwordResetTokenRepository struct {
	db *sqlx.DB
}

func NewPasswordResetTokenRepository(db *sqlx.DB) PasswordResetTokenRepository {
	return &passwordResetTokenRepository{db: db}
}

func (r *passwordResetTokenRepository) Create(token *domain.PasswordResetToken) error {
	query := `
		INSERT INTO password_reset_tokens (id, auth_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	token.ID = uuid.New()
	token.CreatedAt = time.Now()

	_, err := r.db.Exec(
		query,
		token.ID,
		token.AuthID,
		token.Token,
		token.ExpiresAt,
		token.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("falha ao criar token de recuperação de senha: %w", err)
	}

	return nil
}

func (r *passwordResetTokenRepository) GetByToken(token string) (*domain.PasswordResetToken, error) {
	query := `
		SELECT id, auth_id, token, expires_at, used_at, created_at
		FROM password_reset_tokens
		WHERE token = $1
	`

	var resetToken domain.PasswordResetToken
	err := r.db.Get(&resetToken, query, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("falha ao buscar token de recuperação de senha: %w", err)
	}

	return &resetToken, nil
}

func (r *passwordResetTokenRepository) GetByAuthID(authID uuid.UUID) (*domain.PasswordResetToken, error) {
	query := `
		SELECT id, auth_id, token, expires_at, used_at, created_at
		FROM password_reset_tokens
		WHERE auth_id = $1 AND used_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`

	var resetToken domain.PasswordResetToken
	err := r.db.Get(&resetToken, query, authID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("falha ao buscar o token de recuperação de senha pelo auth_id: %w", err)
	}

	return &resetToken, nil
}

func (r *passwordResetTokenRepository) MarkAsUsed(id uuid.UUID) error {
	query := `
		UPDATE password_reset_tokens
		SET used_at = $1
		WHERE id = $2
	`

	result, err := r.db.Exec(query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("falha ao marcar token como usado: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("falha ao verificar as linhas afetadas: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *passwordResetTokenRepository) DeleteExpired() error {
	query := `
		DELETE FROM password_reset_tokens
		WHERE expires_at < $1
	`

	_, err := r.db.Exec(query, time.Now())
	if err != nil {
		return fmt.Errorf("falha ao deletar tokens expirados: %w", err)
	}

	return nil
}

func (r *passwordResetTokenRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM password_reset_tokens WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("falha ao deletar token de recuperação de senha: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("falha ao verificar as linhas afetadas: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
