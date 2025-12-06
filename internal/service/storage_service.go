package service

import (
	"mime/multipart"
)

// StorageService interface para upload de arquivos
type StorageService interface {
	UploadFile(file *multipart.FileHeader, path string) (string, error)
	DeleteFile(fileURL string) error
	GetFileURL(key string) string
}
