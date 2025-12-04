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

type ActivationTokenRepository interface {
	Create(token *domain.ActivationToken) error
	GetByToken(token string) (*domain.ActivationToken, error)
	GetByAuthID(authID uuid.UUID) (*domain.ActivationToken, error)
	MarkAsUsed(id uuid.UUID) error
	DeleteExpired() error
	Delete(id uuid.UUID) error
}

type activationTokenRepository struct {
	db *sqlx.DB
}

func NewActivationTokenRepository(db *sqlx.DB) ActivationTokenRepository {
	return &activationTokenRepository{db: db}
}

func (r *activationTokenRepository) Create(token *domain.ActivationToken) error {
	query := `
		INSERT INTO activation_tokens (id, auth_id, token, expires_at, created_at)
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
		return fmt.Errorf("falha ao criar token de ativação: %w", err)
	}

	return nil
}

func (r *activationTokenRepository) GetByToken(token string) (*domain.ActivationToken, error) {
	query := `
		SELECT id, auth_id, token, expires_at, used_at, created_at
		FROM activation_tokens
		WHERE token = $1
	`

	var activationToken domain.ActivationToken
	err := r.db.Get(&activationToken, query, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("falha ao buscar token de ativação: %w", err)
	}

	return &activationToken, nil
}

func (r *activationTokenRepository) GetByAuthID(authID uuid.UUID) (*domain.ActivationToken, error) {
	query := `
		SELECT id, auth_id, token, expires_at, used_at, created_at
		FROM activation_tokens
		WHERE auth_id = $1 AND used_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`

	var activationToken domain.ActivationToken
	err := r.db.Get(&activationToken, query, authID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("falha ao buscar o token de ativação pelo auth_id: %w", err)
	}

	return &activationToken, nil
}

func (r *activationTokenRepository) MarkAsUsed(id uuid.UUID) error {
	query := `
		UPDATE activation_tokens
		SET used_at = $1
		WHERE id = $2
	`

	result, err := r.db.Exec(query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("falha ao marcar o token como usado: %w", err)
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

func (r *activationTokenRepository) DeleteExpired() error {
	query := `
		DELETE FROM activation_tokens
		WHERE expires_at < $1
	`

	_, err := r.db.Exec(query, time.Now())
	if err != nil {
		return fmt.Errorf("falha ao deletar os tokens expirados: %w", err)
	}

	return nil
}

func (r *activationTokenRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM activation_tokens WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("falha ao deletar token de ativação: %w", err)
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
