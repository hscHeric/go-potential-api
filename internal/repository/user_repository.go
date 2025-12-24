package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hscHeric/go-potential-api/internal/domain"
	"github.com/jmoiron/sqlx"
)

type UserRepository interface {
	Create(user *domain.User) error
	GetByID(id uuid.UUID) (*domain.User, error)
	GetByAuthID(authID uuid.UUID) (*domain.User, error)
	GetByCPF(cpf string) (*domain.User, error)
	Update(user *domain.User) error
	UpdateProfilePic(id uuid.UUID, profilePic string) error
	Delete(id uuid.UUID) error
	ExistsByCPF(cpf string) (bool, error)
	ExistsByAuthID(authID uuid.UUID) (bool, error)
}

func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{db: db}
}

type userRepository struct {
	db *sqlx.DB
}

func (r *userRepository) Create(user *domain.User) error {
	query := `
		INSERT INTO users (id, auth_id, full_name, cpf, birth_date, address, contact, profile_pic, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// Converter structs para JSONB
	addressJSON, err := json.Marshal(user.Address)
	if err != nil {
		return fmt.Errorf("falha ao converter dados de endereço para json: %w", err)
	}

	contactJSON, err := json.Marshal(user.Contact)
	if err != nil {
		return fmt.Errorf("falha ao converter dados de contato para json: %w", err)
	}

	_, err = r.db.Exec(
		query,
		user.ID,
		user.AuthID,
		user.FullName,
		user.CPF,
		user.BirthDate,
		addressJSON,
		contactJSON,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		if IsDuplicateKeyError(err) {
			return ErrCPFAlreadyExists
		}
		return fmt.Errorf("falha ao criar usuário: %w", err)
	}

	return nil
}

func (r *userRepository) GetByID(id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, auth_id, full_name, cpf, birth_date, address, contact, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user domain.User
	var addressJSON, contactJSON []byte

	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.AuthID,
		&user.FullName,
		&user.CPF,
		&user.BirthDate,
		&addressJSON,
		&contactJSON,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("falha ao buscar o usuário pelo id: %w", err)
	}

	// Unmarshal JSONB
	if err := json.Unmarshal(addressJSON, &user.Address); err != nil {
		return nil, fmt.Errorf("falha ao converter os dados de endereço em struct: %w", err)
	}

	if err := json.Unmarshal(contactJSON, &user.Contact); err != nil {
		return nil, fmt.Errorf("falha ao converter os dados de contato em struct: %w", err)
	}

	return &user, nil
}

func (r *userRepository) GetByAuthID(authID uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, auth_id, full_name, cpf, birth_date, address, contact, created_at, updated_at
		FROM users
		WHERE auth_id = $1
	`

	var user domain.User
	var addressJSON, contactJSON []byte

	err := r.db.QueryRow(query, authID).Scan(
		&user.ID,
		&user.AuthID,
		&user.FullName,
		&user.CPF,
		&user.BirthDate,
		&addressJSON,
		&contactJSON,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("falha ao buscar usuario pelo auth_id: %w", err)
	}

	if err := json.Unmarshal(addressJSON, &user.Address); err != nil {
		return nil, fmt.Errorf("falha ao converter o endereço em struct: %w", err)
	}

	if err := json.Unmarshal(contactJSON, &user.Contact); err != nil {
		return nil, fmt.Errorf("falha ao converter os contatos em struct: %w", err)
	}

	return &user, nil
}

func (r *userRepository) GetByCPF(cpf string) (*domain.User, error) {
	query := `
		SELECT id, auth_id, full_name, cpf, birth_date, address, contact, created_at, updated_at
		FROM users
		WHERE cpf = $1
	`

	var user domain.User
	var addressJSON, contactJSON []byte

	err := r.db.QueryRow(query, cpf).Scan(
		&user.ID,
		&user.AuthID,
		&user.FullName,
		&user.CPF,
		&user.BirthDate,
		&addressJSON,
		&contactJSON,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("falha ao selecionar o cpf pelo usuário: %w", err)
	}

	if err := json.Unmarshal(addressJSON, &user.Address); err != nil {
		return nil, fmt.Errorf("falha ao  converter o endereçoJSON para struct: %w", err)
	}

	if err := json.Unmarshal(contactJSON, &user.Contact); err != nil {
		return nil, fmt.Errorf("falha ao converter contatoJSON para struct: %w", err)
	}

	return &user, nil
}

func (r *userRepository) Update(user *domain.User) error {
	query := `
		UPDATE users
		SET full_name = $1, cpf = $2, birth_date = $3, address = $4, contact = $5, updated_at = $6
		WHERE id = $7
	`

	user.UpdatedAt = time.Now()

	addressJSON, err := json.Marshal(user.Address)
	if err != nil {
		return fmt.Errorf("falha ao converter endereço para json: %w", err)
	}

	contactJSON, err := json.Marshal(user.Contact)
	if err != nil {
		return fmt.Errorf("falha ao converter dados de contato para json: %w", err)
	}

	result, err := r.db.Exec(
		query,
		user.FullName,
		user.CPF,
		user.BirthDate,
		addressJSON,
		contactJSON,
		user.UpdatedAt,
		user.ID,
	)
	if err != nil {
		if IsDuplicateKeyError(err) {
			return ErrCPFAlreadyExists
		}
		return fmt.Errorf("falha ao atualizar o usuário: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("falha ao verificar o número de linhas afetadas: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *userRepository) UpdateProfilePic(id uuid.UUID, profilePic string) error {
	query := `
		UPDATE users
		SET profile_pic = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.Exec(query, profilePic, time.Now(), id)
	if err != nil {
		return fmt.Errorf("falha ao atualizar a foto de perfil do usuário: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("falha ao verificar o número de linhas afetadas: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *userRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("falha ao deletar user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("falha ao verificar o números de linhas afetadas: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *userRepository) ExistsByCPF(cpf string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE cpf = $1)`

	var exists bool
	err := r.db.Get(&exists, query, cpf)
	if err != nil {
		return false, fmt.Errorf("falha ao checar se o cpf existe: %w", err)
	}

	return exists, nil
}

func (r *userRepository) ExistsByAuthID(authID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE auth_id = $1)`

	var exists bool
	err := r.db.Get(&exists, query, authID)
	if err != nil {
		return false, fmt.Errorf("falha ao verificar se o auth_id associado ao usuário existe: %w", err)
	}

	return exists, nil
}
