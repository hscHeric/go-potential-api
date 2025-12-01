package database

import (
	"log"

	"github.com/hscHeric/go-potential-api/internal/config"
	"github.com/hscHeric/go-potential-api/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewPostgresDB(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	dsn := cfg.GetDSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Printf("Falha ao se conectar com o banco de dados: %v", err)
		return nil, err
	}

	log.Println("Conexão com o banco de dados estabelecida com sucesso")
	return db, nil
}

func RunMigrations(db *gorm.DB) error {
	log.Println("Iniciando migrações do banco de dados...")

	err := db.AutoMigrate(
		&domain.User{},
	)
	if err != nil {
		return err
	}

	log.Println("Migrações do banco de dados concluídas com sucesso")

	// Seed default admin
	if err := SeedDefaultAdmin(db); err != nil {
		log.Printf("Warning: Falha ao criar usuário admin padrão: %v", err)
	}

	return nil
}
