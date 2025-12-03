// Package hash genencia a geração e a comparação das hash's de senha
package hash

import "golang.org/x/crypto/bcrypt"

// HashPassword gera um hash bcrypt da senha usando o custo padrão definido pelo bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword verifica se a senha corresponde ao hash
func CheckPassword(password, hash string) bool {
	// Compare password hash compara e lança um erro caso n batam as senhas
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
