// Package repository é responsavel pela persistencia dos dados
package repository

import (
	"errors"
	"strings"
)

var (
	ErrNotFound           = errors.New("registro não encontrado")
	ErrAlreadyExists      = errors.New("registro já existe")
	ErrEmailAlreadyExists = errors.New("email já existe")
	ErrCPFAlreadyExists   = errors.New("cpf já existe")
	ErrInvalidData        = errors.New("dados inválidos")
)

// IsDuplicateKeyError verifica se o erro do banco é de chave duplicada
func IsDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()

	// checks em inglês, pois o DB retorna assim, verificar se funciona
	return strings.Contains(errMsg, "duplicate key") ||
		strings.Contains(errMsg, "UNIQUE constraint failed")
}

// IsNotFoundError verifica se o erro é de registro não encontrado
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrNotFound)
}
