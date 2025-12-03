package domain

import (
	"time"

	"github.com/google/uuid"
)

// DocumentType representa o tipo de documento
type DocumentType string

const (
	DocumentTypeRG           DocumentType = "rg"
	DocumentTypeCPF          DocumentType = "cpf"
	DocumentTypeProofAddress DocumentType = "proof_address"
	DocumentTypeOther        DocumentType = "other"
)

// DocumentStatus representa o status de validação do documento
type DocumentStatus string

const (
	DocumentStatusPending  DocumentStatus = "pending"
	DocumentStatusApproved DocumentStatus = "approved"
	DocumentStatusRejected DocumentStatus = "rejected"
)

// Document representa um documento enviado por aluno ou professor
type Document struct {
	ID              uuid.UUID      `db:"id" json:"id"`
	UserID          uuid.UUID      `db:"user_id" json:"user_id"`
	Type            DocumentType   `db:"type" json:"type" binding:"required,oneof=rg cpf proof_address other"`
	FileName        string         `db:"file_name" json:"file_name"`
	FileURL         string         `db:"file_url" json:"file_url"`
	FileSize        int64          `db:"file_size" json:"file_size"`
	MimeType        string         `db:"mime_type" json:"mime_type"`
	Status          DocumentStatus `db:"status" json:"status"`
	RejectionReason *string        `db:"rejection_reason" json:"rejection_reason,omitempty"`
	UploadedAt      time.Time      `db:"uploaded_at" json:"uploaded_at"`
	ReviewedAt      *time.Time     `db:"reviewed_at" json:"reviewed_at,omitempty"`
	ReviewedBy      *uuid.UUID     `db:"reviewed_by" json:"reviewed_by,omitempty"`
}
