package domain

import (
	"time"

	"github.com/google/uuid"
)

// EntityType representa o tipo de entidade associada ao arquivo
type EntityType string

const (
	EntityTypeUserProfile        EntityType = "user_profile"
	EntityTypeUserDocument       EntityType = "user_document"
	EntityTypeActivityAttachment EntityType = "activity_attachment"
	EntityTypeClassMaterial      EntityType = "class_material"
)

// FileMetadata representa metadados adicionais do arquivo
type FileMetadata map[string]any

type File struct {
	ID               uuid.UUID    `db:"id" json:"id"`
	Filename         string       `db:"filename" json:"filename"`
	OriginalFilename string       `db:"original_filename" json:"original_filename"`
	FileURL          string       `db:"file_url" json:"file_url"`
	FilePath         string       `db:"file_path" json:"file_path"`
	FileSize         int64        `db:"file_size" json:"file_size"`
	MimeType         string       `db:"mime_type" json:"mime_type"`
	EntityType       EntityType   `db:"entity_type" json:"entity_type"`
	EntityID         uuid.UUID    `db:"entity_id" json:"entity_id"`
	Metadata         FileMetadata `db:"metadata" json:"metadata"`
	UploadedBy       *uuid.UUID   `db:"uploaded_by" json:"uploaded_by,omitempty"`
	CreatedAt        time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time    `db:"updated_at" json:"updated_at"`
}
