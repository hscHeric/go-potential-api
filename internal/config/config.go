package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Email    EmailConfig
	Storage  StorageConfig
}

type ServerConfig struct {
	Port        string
	Env         string
	FrontendURL string
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
	Secret                            string
	ExpirationHours                   int
	InvitationTokenExpirationHours    int
	PasswordResetTokenExpirationHours int
}

type EmailConfig struct {
	SMTPHost string
	SMTPPort string
	SMTPUser string
	SMTPPass string
	From     string
}

type StorageConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	Region    string
	UseSSL    bool
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Error reading config file: %v", err)
		return nil, err
	}

	config := &Config{
		Server: ServerConfig{
			Port:        viper.GetString("PORT"),
			Env:         viper.GetString("ENV"),
			FrontendURL: viper.GetString("FRONTEND_URL"),
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
			Secret:                            viper.GetString("JWT_SECRET"),
			ExpirationHours:                   viper.GetInt("JWT_EXPIRATION_HOURS"),
			InvitationTokenExpirationHours:    viper.GetInt("INVITATION_TOKEN_EXPIRATION_HOURS"),
			PasswordResetTokenExpirationHours: viper.GetInt("PASSWORD_RESET_TOKEN_EXPIRATION_HOURS"),
		},
		Email: EmailConfig{
			SMTPHost: viper.GetString("SMTP_HOST"),
			SMTPPort: viper.GetString("SMTP_PORT"),
			SMTPUser: viper.GetString("SMTP_USER"),
			SMTPPass: viper.GetString("SMTP_PASSWORD"),
			From:     viper.GetString("SMTP_FROM"),
		},
		Storage: StorageConfig{
			Endpoint:  viper.GetString("S3_ENDPOINT"),
			AccessKey: viper.GetString("S3_ACCESS_KEY"),
			SecretKey: viper.GetString("S3_SECRET_KEY"),
			Bucket:    viper.GetString("S3_BUCKET"),
			Region:    viper.GetString("S3_REGION"),
			UseSSL:    viper.GetBool("S3_USE_SSL"),
		},
	}

	return config, nil
}

func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}
