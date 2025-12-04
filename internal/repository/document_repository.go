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

type DocumentRepository interface {
	Create(document *domain.Document) error
	GetByID(id uuid.UUID) (*domain.Document, error)
	GetByUserID(userID uuid.UUID) ([]domain.Document, error)
	GetByUserIDAndType(userID uuid.UUID, docType domain.DocumentType) ([]domain.Document, error)
	GetPendingDocuments() ([]domain.Document, error)
	UpdateStatus(id uuid.UUID, status domain.DocumentStatus, rejectionReason *string, reviewedBy uuid.UUID) error
	Delete(id uuid.UUID) error
}

type documentRepository struct {
	db *sqlx.DB
}

func NewDocumentRepository(db *sqlx.DB) DocumentRepository {
	return &documentRepository{db: db}
}

func (r *documentRepository) Create(document *domain.Document) error {
	query := `
		INSERT INTO documents (id, user_id, type, file_name, file_url, file_size, mime_type, status, uploaded_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	document.ID = uuid.New()
	document.UploadedAt = time.Now()

	_, err := r.db.Exec(
		query,
		document.ID,
		document.UserID,
		document.Type,
		document.FileName,
		document.FileURL,
		document.FileSize,
		document.MimeType,
		document.Status,
		document.UploadedAt,
	)
	if err != nil {
		return fmt.Errorf("falha ao criar documento: %w", err)
	}

	return nil
}

func (r *documentRepository) GetByID(id uuid.UUID) (*domain.Document, error) {
	query := `
		SELECT id, user_id, type, file_name, file_url, file_size, mime_type, status, 
		       rejection_reason, uploaded_at, reviewed_at, reviewed_by
		FROM documents
		WHERE id = $1
	`

	var document domain.Document
	err := r.db.Get(&document, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("falha ao buscar documento: %w", err)
	}

	return &document, nil
}

func (r *documentRepository) GetByUserID(userID uuid.UUID) ([]domain.Document, error) {
	query := `
		SELECT id, user_id, type, file_name, file_url, file_size, mime_type, status, 
		       rejection_reason, uploaded_at, reviewed_at, reviewed_by
		FROM documents
		WHERE user_id = $1
		ORDER BY uploaded_at DESC
	`

	var documents []domain.Document
	err := r.db.Select(&documents, query, userID)
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar documentos pelo user_id: %w", err)
	}

	return documents, nil
}

func (r *documentRepository) GetByUserIDAndType(userID uuid.UUID, docType domain.DocumentType) ([]domain.Document, error) {
	query := `
		SELECT id, user_id, type, file_name, file_url, file_size, mime_type, status, 
		       rejection_reason, uploaded_at, reviewed_at, reviewed_by
		FROM documents
		WHERE user_id = $1 AND type = $2
		ORDER BY uploaded_at DESC
	`

	var documents []domain.Document
	err := r.db.Select(&documents, query, userID, docType)
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar documentos pelo user_id e pelo tipo: %w", err)
	}

	return documents, nil
}

func (r *documentRepository) GetPendingDocuments() ([]domain.Document, error) {
	query := `
		SELECT id, user_id, type, file_name, file_url, file_size, mime_type, status, 
		       rejection_reason, uploaded_at, reviewed_at, reviewed_by
		FROM documents
		WHERE status = $1
		ORDER BY uploaded_at ASC
	`

	var documents []domain.Document
	err := r.db.Select(&documents, query, domain.DocumentStatusPending)
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar documentos pendentes: %w", err)
	}

	return documents, nil
}

func (r *documentRepository) UpdateStatus(id uuid.UUID, status domain.DocumentStatus, rejectionReason *string, reviewedBy uuid.UUID) error {
	query := `
		UPDATE documents
		SET status = $1, rejection_reason = $2, reviewed_at = $3, reviewed_by = $4
		WHERE id = $5
	`

	result, err := r.db.Exec(query, status, rejectionReason, time.Now(), reviewedBy, id)
	if err != nil {
		return fmt.Errorf("falha ao atualizar o status do documento: %w", err)
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

func (r *documentRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM documents WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("falha ao deletar documento: %w", err)
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
