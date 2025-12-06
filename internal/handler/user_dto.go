package handler

import "github.com/hscHeric/go-potential-api/internal/domain"

// UpdateProfileRequest representa o payload para atualizar perfil
type UpdateProfileRequest struct {
	FullName  string         `json:"full_name" binding:"required,min=3"`
	BirthDate string         `json:"birth_date" binding:"required"`
	Address   domain.Address `json:"address" binding:"required"`
	Contact   domain.Contact `json:"contact" binding:"required"`
}
