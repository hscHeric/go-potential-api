package database

import (
	"log"

	"github.com/hscHeric/go-potential-api/internal/domain"
	"github.com/hscHeric/go-potential-api/pkg/utils"
	"gorm.io/gorm"
)

func SeedDefaultAdmin(db *gorm.DB) error {
	log.Println("Verificando existência de admin user...")

	// Verificar se já existe um admin
	var adminCount int64
	db.Model(&domain.User{}).Where("role = ?", domain.RoleAdmin).Count(&adminCount)

	if adminCount > 0 {
		log.Println("Já existe um usuário admin.")
		return nil
	}

	// Criar admin padrão
	log.Println("Criando um usuário admin padrão...")

	hashedPassword, err := utils.HashPassword("admin123456")
	if err != nil {
		return err
	}

	admin := &domain.User{
		Email:        "admin@potential.com",
		PasswordHash: hashedPassword,
		Role:         domain.RoleAdmin,
		IsActive:     true,
	}

	if err := db.Create(admin).Error; err != nil {
		return err
	}

	log.Println(" Default admin criado com sucesso!")
	log.Println(" Email: admin@potential.com")
	log.Println(" Password: admin123456")
	log.Println("  CHANGE THIS PASSWORD IN PRODUCTION!")

	return nil
}
