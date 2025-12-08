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

type FileRepository interface {
	Create(file *domain.File) error
	GetByID(id uuid.UUID) (*domain.File, error)
	GetByEntity(entityType domain.EntityType, entityID uuid.UUID) ([]domain.File, error)
	GetLatestByEntity(entityType domain.EntityType, entityID uuid.UUID) (*domain.File, error)
	Delete(id uuid.UUID) error
	DeleteByEntity(entityType domain.EntityType, entityID uuid.UUID) error
}

type fileRepository struct {
	db *sqlx.DB
}

func NewFileRepository(db *sqlx.DB) FileRepository {
	return &fileRepository{db: db}
}

func (r *fileRepository) Create(file *domain.File) error {
	query := `
		INSERT INTO files (
			id, filename, original_filename, file_url, file_path, file_size, 
			mime_type, entity_type, entity_id, metadata, uploaded_by, 
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	file.ID = uuid.New()
	file.CreatedAt = time.Now()
	file.UpdatedAt = time.Now()

	// Serializar metadata
	metadataJSON, err := json.Marshal(file.Metadata)
	if err != nil {
		return fmt.Errorf("falha ao serializar metadata: %w", err)
	}

	_, err = r.db.Exec(
		query,
		file.ID,
		file.Filename,
		file.OriginalFilename,
		file.FileURL,
		file.FilePath,
		file.FileSize,
		file.MimeType,
		file.EntityType,
		file.EntityID,
		metadataJSON,
		file.UploadedBy,
		file.CreatedAt,
		file.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("falha ao criar file: %w", err)
	}

	return nil
}

func (r *fileRepository) GetByID(id uuid.UUID) (*domain.File, error) {
	query := `
		SELECT id, filename, original_filename, file_url, file_path, file_size,
		       mime_type, entity_type, entity_id, metadata, uploaded_by,
		       created_at, updated_at
		FROM files
		WHERE id = $1
	`

	var file domain.File
	var metadataJSON []byte

	err := r.db.QueryRow(query, id).Scan(
		&file.ID,
		&file.Filename,
		&file.OriginalFilename,
		&file.FileURL,
		&file.FilePath,
		&file.FileSize,
		&file.MimeType,
		&file.EntityType,
		&file.EntityID,
		&metadataJSON,
		&file.UploadedBy,
		&file.CreatedAt,
		&file.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("falha ao buscar file: %w", err)
	}

	// Deserializar metadata
	if err := json.Unmarshal(metadataJSON, &file.Metadata); err != nil {
		return nil, fmt.Errorf("falha ao desserializar metadata: %w", err)
	}

	return &file, nil
}

func (r *fileRepository) GetByEntity(entityType domain.EntityType, entityID uuid.UUID) ([]domain.File, error) {
	query := `
		SELECT id, filename, original_filename, file_url, file_path, file_size,
		       mime_type, entity_type, entity_id, metadata, uploaded_by,
		       created_at, updated_at
		FROM files
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar arquivos pelo entityType: %w", err)
	}

	defer func() {
		_ = rows.Close()
	}()

	var files []domain.File
	for rows.Next() {
		var file domain.File
		var metadataJSON []byte

		err := rows.Scan(
			&file.ID,
			&file.Filename,
			&file.OriginalFilename,
			&file.FileURL,
			&file.FilePath,
			&file.FileSize,
			&file.MimeType,
			&file.EntityType,
			&file.EntityID,
			&metadataJSON,
			&file.UploadedBy,
			&file.CreatedAt,
			&file.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("falha ao ler dados do file: %w", err)
		}

		if err := json.Unmarshal(metadataJSON, &file.Metadata); err != nil {
			return nil, fmt.Errorf("falha ao desserializar metadata: %w", err)
		}

		files = append(files, file)
	}

	return files, nil
}

func (r *fileRepository) GetLatestByEntity(entityType domain.EntityType, entityID uuid.UUID) (*domain.File, error) {
	query := `
		SELECT id, filename, original_filename, file_url, file_path, file_size,
		       mime_type, entity_type, entity_id, metadata, uploaded_by,
		       created_at, updated_at
		FROM files
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	var file domain.File
	var metadataJSON []byte

	err := r.db.QueryRow(query, entityType, entityID).Scan(
		&file.ID,
		&file.Filename,
		&file.OriginalFilename,
		&file.FileURL,
		&file.FilePath,
		&file.FileSize,
		&file.MimeType,
		&file.EntityType,
		&file.EntityID,
		&metadataJSON,
		&file.UploadedBy,
		&file.CreatedAt,
		&file.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("falha ao buscar o file mais recente: %w", err)
	}

	if err := json.Unmarshal(metadataJSON, &file.Metadata); err != nil {
		return nil, fmt.Errorf("falha ao desserializar metadata: %w", err)
	}

	return &file, nil
}

func (r *fileRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM files WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("falha ao excluir o file: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("falha ao obter número de linhas afetadas: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *fileRepository) DeleteByEntity(entityType domain.EntityType, entityID uuid.UUID) error {
	query := `DELETE FROM files WHERE entity_type = $1 AND entity_id = $2`

	_, err := r.db.Exec(query, entityType, entityID)
	if err != nil {
		return fmt.Errorf("falha ao excluir arquivos: %w", err)
	}

	return nil
}
