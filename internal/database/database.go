// Package database estabece uma conexeção com o banco de dados
package database

import (
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// Connect estabele conexeção com o banco de dados
func Connect(cfg Config) (*sqlx.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("falha ao estabelecer conexeção com o banco de dados: %w", err)
	}

	// Extrair isso para o arquivo .env se for o caso
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("falha ao testar conexeção com o banco de dados: %w", err)
	}

	log.Println("conexeção com o banco de dados estabelecida com sucesso")
	return db, nil
}

// Close fecha a conexão com o banco de dados
func Close(db *sqlx.DB) error {
	if db != nil {
		return db.Close()
	}
	return nil
}
