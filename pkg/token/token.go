// Package token gera tokens criptograficamente seguros para confirmação de email e reset de senha
package token

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// Generate gera um token aleatório criptograficamente seguro com o tamanho especificado
func Generate(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateActivationToken gera um token de ativação (32 bytes)
func GenerateActivationToken() (string, error) {
	return Generate(32)
}

// GeneratePasswordResetToken gera um token de recuperação de senha (32 bytes)
func GeneratePasswordResetToken() (string, error) {
	return Generate(32)
}
