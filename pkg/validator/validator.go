// Package validator implementa validações customizadas e função auxiliar para formatar erros.
package validator

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	// Registrar validações customizadas
	_ = validate.RegisterValidation("cpf", validateCPF)
	_ = validate.RegisterValidation("phone", validatePhone)
}

// Validate valida uma struct usando as tags de validação
func Validate(s any) error {
	return validate.Struct(s)
}

// GetValidator retorna a instância do validator
func GetValidator() *validator.Validate {
	return validate
}

// validateCPF valida um CPF brasileiro
func validateCPF(fl validator.FieldLevel) bool {
	cpf := fl.Field().String()

	// Remove caracteres não numéricos
	cpf = regexp.MustCompile(`[^0-9]`).ReplaceAllString(cpf, "")

	if len(cpf) != 11 {
		return false
	}

	// Verifica se todos os dígitos são iguais
	allEqual := true
	for i := 1; i < len(cpf); i++ {
		if cpf[i] != cpf[0] {
			allEqual = false
			break
		}
	}
	if allEqual {
		return false
	}

	// Validação do primeiro dígito verificador
	sum := 0
	for i := range 9 {
		sum += int(cpf[i]-'0') * (10 - i)
	}
	digit1 := (sum * 10) % 11
	if digit1 == 10 {
		digit1 = 0
	}
	if digit1 != int(cpf[9]-'0') {
		return false
	}

	// Validação do segundo dígito verificador
	sum = 0
	for i := range 10 {
		sum += int(cpf[i]-'0') * (11 - i)
	}
	digit2 := (sum * 10) % 11
	if digit2 == 10 {
		digit2 = 0
	}
	if digit2 != int(cpf[10]-'0') {
		return false
	}

	return true
}

// validatePhone valida um telefone brasileiro
func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()

	// Remove caracteres não numéricos
	phone = regexp.MustCompile(`[^0-9]`).ReplaceAllString(phone, "")

	// Telefone fixo: 10 dígitos (DDD + 8 dígitos)
	// Celular: 11 dígitos (DDD + 9 dígitos)
	return len(phone) == 10 || len(phone) == 11
}

// FormatValidationErrors formata erros de validação para mensagens amigáveis em português
func FormatValidationErrors(err error) map[string]string {
	errors := make(map[string]string)

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := strings.ToLower(e.Field())

			switch e.Tag() {
			case "required":
				errors[field] = "Este campo é obrigatório"
			case "email":
				errors[field] = "Formato de e-mail inválido"
			case "min":
				errors[field] = "O tamanho mínimo é " + e.Param()
			case "max":
				errors[field] = "O tamanho máximo é " + e.Param()
			case "len":
				errors[field] = "O tamanho deve ser exatamente " + e.Param()
			case "cpf":
				errors[field] = "CPF inválido"
			case "phone":
				errors[field] = "Número de telefone inválido"
			case "oneof":
				errors[field] = "O valor deve ser um dos seguintes: " + e.Param()
			default:
				errors[field] = "Valor inválido"
			}
		}
	}

	return errors
}
