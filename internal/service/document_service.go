package service

import (
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/hscHeric/go-potential-api/internal/domain"
	"github.com/hscHeric/go-potential-api/internal/repository"
)

var (
	ErrInvalidFileType = errors.New("tipo de arquivo inválido")
	ErrFileTooLarge    = errors.New("arquivo muito grande")
)

const (
	MaxFileSize = 10 * 1024 * 1024 // 10MB
)

var allowedMimeTypes = map[string]bool{
	"application/pdf":    true,
	"image/jpeg":         true,
	"image/jpg":          true,
	"image/png":          true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
}

type DocumentService interface {
	UploadDocument(authID uuid.UUID, docType domain.DocumentType, file *multipart.FileHeader) (*domain.Document, error)
	GetUserDocuments(authID uuid.UUID) ([]domain.Document, error)
	GetDocumentByID(authID uuid.UUID, documentID uuid.UUID) (*domain.Document, error)
	ApproveDocument(adminAuthID, documentID uuid.UUID) error
	RejectDocument(adminAuthID, documentID uuid.UUID, reason string) error
	GetPendingDocuments() ([]domain.Document, error)
	DeleteDocument(authID uuid.UUID, documentID uuid.UUID) error
}

type documentService struct {
	userRepo       repository.UserRepository
	documentRepo   repository.DocumentRepository
	storageService StorageService
}

func NewDocumentService(
	userRepo repository.UserRepository,
	documentRepo repository.DocumentRepository,
	storageService StorageService,
) DocumentService {
	return &documentService{
		userRepo:       userRepo,
		documentRepo:   documentRepo,
		storageService: storageService,
	}
}

func (s *documentService) UploadDocument(authID uuid.UUID, docType domain.DocumentType, file *multipart.FileHeader) (*domain.Document, error) {
	// Buscar user
	user, err := s.userRepo.GetByAuthID(authID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, errors.New("usuário não encontrado")
		}
		return nil, fmt.Errorf("falha ao obter usuário: %w", err)
	}

	// Validar arquivo
	if err := s.validateFile(file); err != nil {
		return nil, err
	}

	// Fazer upload para S3
	fileURL, err := s.storageService.UploadFile(file, fmt.Sprintf("documents/%s", user.ID.String()))
	if err != nil {
		return nil, fmt.Errorf("falha ao fazer upload do arquivo: %w", err)
	}

	// Criar registro do documento
	document := &domain.Document{
		UserID:   user.ID,
		Type:     docType,
		FileName: file.Filename,
		FileURL:  fileURL,
		FileSize: file.Size,
		MimeType: file.Header.Get("Content-Type"),
		Status:   domain.DocumentStatusPending,
	}

	if err := s.documentRepo.Create(document); err != nil {
		// Tentar deletar arquivo do S3 em caso de erro
		_ = s.storageService.DeleteFile(fileURL)
		return nil, fmt.Errorf("falha ao criar registro do documento: %w", err)
	}

	return document, nil
}

func (s *documentService) GetUserDocuments(authID uuid.UUID) ([]domain.Document, error) {
	// Buscar user
	user, err := s.userRepo.GetByAuthID(authID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, errors.New("usuário não encontrado")
		}
		return nil, fmt.Errorf("falha ao obter usuário: %w", err)
	}

	documents, err := s.documentRepo.GetByUserID(user.ID)
	if err != nil {
		return nil, fmt.Errorf("falha ao obter documentos: %w", err)
	}

	return documents, nil
}

func (s *documentService) GetDocumentByID(authID uuid.UUID, documentID uuid.UUID) (*domain.Document, error) {
	// Buscar user
	user, err := s.userRepo.GetByAuthID(authID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, errors.New("usuário não encontrado")
		}
		return nil, fmt.Errorf("falha ao obter usuário: %w", err)
	}

	// Buscar documento
	document, err := s.documentRepo.GetByID(documentID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, errors.New("documento não encontrado")
		}
		return nil, fmt.Errorf("falha ao obter documento: %w", err)
	}

	// Verificar se o documento pertence ao usuário
	if document.UserID != user.ID {
		return nil, errors.New("acesso não autorizado ao documento")
	}

	return document, nil
}

func (s *documentService) ApproveDocument(adminAuthID, documentID uuid.UUID) error {
	// Buscar documento
	document, err := s.documentRepo.GetByID(documentID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return errors.New("documento não encontrado")
		}
		return fmt.Errorf("falha ao obter documento: %w", err)
	}

	// Atualizar status
	if err := s.documentRepo.UpdateStatus(
		document.ID,
		domain.DocumentStatusApproved,
		nil,
		adminAuthID,
	); err != nil {
		return fmt.Errorf("falha ao aprovar documento: %w", err)
	}

	return nil
}

func (s *documentService) RejectDocument(adminAuthID, documentID uuid.UUID, reason string) error {
	// Buscar documento
	document, err := s.documentRepo.GetByID(documentID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return errors.New("documento não encontrado")
		}
		return fmt.Errorf("falha ao obter documento: %w", err)
	}

	// Atualizar status
	if err := s.documentRepo.UpdateStatus(
		document.ID,
		domain.DocumentStatusRejected,
		&reason,
		adminAuthID,
	); err != nil {
		return fmt.Errorf("falha ao rejeitar documento: %w", err)
	}

	return nil
}

func (s *documentService) GetPendingDocuments() ([]domain.Document, error) {
	documents, err := s.documentRepo.GetPendingDocuments()
	if err != nil {
		return nil, fmt.Errorf("falha ao obter documentos pendentes: %w", err)
	}

	return documents, nil
}

func (s *documentService) DeleteDocument(authID uuid.UUID, documentID uuid.UUID) error {
	// Buscar user
	user, err := s.userRepo.GetByAuthID(authID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return errors.New("usuário não encontrado")
		}
		return fmt.Errorf("falha ao obter usuário: %w", err)
	}

	// Buscar documento
	document, err := s.documentRepo.GetByID(documentID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return errors.New("documento não encontrado")
		}
		return fmt.Errorf("falha ao obter documento: %w", err)
	}

	// Verificar se o documento pertence ao usuário
	if document.UserID != user.ID {
		return errors.New("não autorizado a deletar este documento")
	}

	// Deletar arquivo do S3
	if err := s.storageService.DeleteFile(document.FileURL); err != nil {
		// Logar erro mas continuar com a deleção do registro
		fmt.Printf("Aviso: falha ao deletar arquivo do storage: %v\n", err)
	}

	// Deletar registro
	if err := s.documentRepo.Delete(documentID); err != nil {
		return fmt.Errorf("falha ao deletar documento: %w", err)
	}

	return nil
}

func (s *documentService) validateFile(file *multipart.FileHeader) error {
	// Validar tamanho
	if file.Size > MaxFileSize {
		return ErrFileTooLarge
	}

	// Validar tipo MIME
	contentType := file.Header.Get("Content-Type")
	if !allowedMimeTypes[contentType] {
		return ErrInvalidFileType
	}

	// Validar extensão
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExtensions := map[string]bool{
		".pdf":  true,
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".doc":  true,
		".docx": true,
	}

	if !allowedExtensions[ext] {
		return ErrInvalidFileType
	}

	return nil
}
