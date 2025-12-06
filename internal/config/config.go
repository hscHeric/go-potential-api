// Package config carrega as informações do arquivo .env usando viper
package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Email    EmailConfig
	S3       S3Config
	Tokens   TokensConfig
}

type ServerConfig struct {
	Port   string
	AppEnv string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type JWTConfig struct {
	Secret          string
	ExpirationHours int
}

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
}

type S3Config struct {
	Endpoint  string
	Region    string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

type TokensConfig struct {
	ActivationExpirationHours    int
	PasswordResetExpirationHours int
	FrontendURL                  string
}

// LoadConfig carrega as configurações usando Viper
func LoadConfig(path string) (*Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(path)
	viper.AddConfigPath(".")

	// Permite sobrescrever com variáveis de ambiente
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo de configuração: %w", err)
	}

	config := &Config{
		Server: ServerConfig{
			Port:   viper.GetString("PORT"),
			AppEnv: viper.GetString("APP_ENV"),
		},
		Database: DatabaseConfig{
			Host:     viper.GetString("DB_HOST"),
			Port:     viper.GetString("DB_PORT"),
			User:     viper.GetString("DB_USER"),
			Password: viper.GetString("DB_PASSWORD"),
			DBName:   viper.GetString("DB_NAME"),
			SSLMode:  viper.GetString("DB_SSLMODE"),
		},
		JWT: JWTConfig{
			Secret:          viper.GetString("JWT_SECRET"),
			ExpirationHours: viper.GetInt("JWT_EXPIRATION_HOURS"),
		},
		Email: EmailConfig{
			SMTPHost:     viper.GetString("SMTP_HOST"),
			SMTPPort:     viper.GetInt("SMTP_PORT"),
			SMTPUsername: viper.GetString("SMTP_USERNAME"),
			SMTPPassword: viper.GetString("SMTP_PASSWORD"),
			SMTPFrom:     viper.GetString("SMTP_FROM"),
		},
		S3: S3Config{
			Endpoint:  viper.GetString("S3_ENDPOINT"),
			Region:    viper.GetString("S3_REGION"),
			AccessKey: viper.GetString("S3_ACCESS_KEY"),
			SecretKey: viper.GetString("S3_SECRET_KEY"),
			Bucket:    viper.GetString("S3_BUCKET"),
			UseSSL:    viper.GetBool("S3_USE_SSL"),
		},
		Tokens: TokensConfig{
			ActivationExpirationHours:    viper.GetInt("ACTIVATION_TOKEN_EXPIRATION_HOURS"),
			PasswordResetExpirationHours: viper.GetInt("PASSWORD_RESET_TOKEN_EXPIRATION_HOURS"),
			FrontendURL:                  viper.GetString("FRONTEND_URL"),
		},
	}

	return config, nil
}

// GetDatabaseURL retorna a URL de conexão do banco
func (c *Config) GetDatabaseURL() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}

// GetJWTExpiration retorna a duração de expiração do JWT
func (c *Config) GetJWTExpiration() time.Duration {
	return time.Duration(c.JWT.ExpirationHours) * time.Hour
}

// GetActivationTokenExpiration retorna a duração de expiração do token de ativação
func (c *Config) GetActivationTokenExpiration() time.Duration {
	return time.Duration(c.Tokens.ActivationExpirationHours) * time.Hour
}

// GetPasswordResetTokenExpiration retorna a duração de expiração do token de reset
func (c *Config) GetPasswordResetTokenExpiration() time.Duration {
	return time.Duration(c.Tokens.PasswordResetExpirationHours) * time.Hour
}
