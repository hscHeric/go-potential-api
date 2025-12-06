// Package service implementa as regras de negócio da aplicação
package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/hscHeric/go-potential-api/internal/domain"
	"github.com/hscHeric/go-potential-api/internal/repository"
	"github.com/hscHeric/go-potential-api/pkg/email"
	"github.com/hscHeric/go-potential-api/pkg/hash"
	"github.com/hscHeric/go-potential-api/pkg/jwt"
	"github.com/hscHeric/go-potential-api/pkg/token"
)

var (
	ErrInvalidCredentials   = errors.New("credenciais inválidas")
	ErrUserNotActive        = errors.New("usuário não está ativo")
	ErrTokenExpired         = errors.New("o token expirou")
	ErrTokenAlreadyUsed     = errors.New("o token já foi utilizado")
	ErrUserAlreadyCompleted = errors.New("o registro do usuário já foi concluído")
)

type AuthService interface {
	// Admin cria convite inicial de finindo se o email e se é aluno ou professor
	CreateInvitation(email string, role domain.Role) error

	// Usuário completa o cadastro usando um token de ativação de conta
	CompleteRegistration(activationToken string, userData *CompleteRegistrationInput) error

	// Login
	Login(email, password string) (*LoginResponse, error)

	// RequestPasswordReset solicita envio de email para recuperação de senha
	RequestPasswordReset(email string) error

	// ResetPassword altera a senha do usuário
	ResetPassword(resetToken, newPassword string) error

	// Validação de token
	ValidateActivationToken(token string) (*domain.Auth, error)
	ValidatePasswordResetToken(token string) (*domain.Auth, error)
}

type CompleteRegistrationInput struct {
	FullName  string         `json:"full_name" binding:"required,min=3"`
	CPF       string         `json:"cpf" binding:"required,len=11"`
	BirthDate time.Time      `json:"birth_date" binding:"required"`
	Address   domain.Address `json:"address" binding:"required"`
	Contact   domain.Contact `json:"contact" binding:"required"`
	Password  string         `json:"password" binding:"required,min=8"`
}

type LoginResponse struct {
	Token string               `json:"token"`
	User  *domain.UserWithAuth `json:"user"`
}

type authService struct {
	authRepo                repository.AuthRepository
	userRepo                repository.UserRepository
	activationTokenRepo     repository.ActivationTokenRepository
	passwordResetRepo       repository.PasswordResetTokenRepository
	jwtService              *jwt.Service
	emailService            *email.Service
	activationExpiration    time.Duration
	passwordResetExpiration time.Duration
}

func NewAuthService(
	authRepo repository.AuthRepository,
	userRepo repository.UserRepository,
	activationTokenRepo repository.ActivationTokenRepository,
	passwordResetRepo repository.PasswordResetTokenRepository,
	jwtService *jwt.Service,
	emailService *email.Service,
	activationExpiration time.Duration,
	passwordResetExpiration time.Duration,
) AuthService {
	return &authService{
		authRepo:                authRepo,
		userRepo:                userRepo,
		activationTokenRepo:     activationTokenRepo,
		passwordResetRepo:       passwordResetRepo,
		jwtService:              jwtService,
		emailService:            emailService,
		activationExpiration:    activationExpiration,
		passwordResetExpiration: passwordResetExpiration,
	}
}

func (s *authService) CreateInvitation(email string, role domain.Role) error {
	exists, err := s.authRepo.ExistsByEmail(email)
	if err != nil {
		return fmt.Errorf("falha ao verificar o e-mail: %w", err)
	}

	if exists {
		return repository.ErrEmailAlreadyExists
	}

	auth := &domain.Auth{
		Email:        email,
		PasswordHash: "", // Isso é definido no complete registration
		Role:         role,
		Status:       domain.StatusPending,
	}

	// Cria as credencias no banco de dados
	if err := s.authRepo.Create(auth); err != nil {
		return fmt.Errorf("falha ao criar credenciais (Auth): %w", err)
	}

	// Gera um token de ativação
	activationToken, err := token.GenerateActivationToken()
	if err != nil {
		return fmt.Errorf("falha ao gerar token de ativação: %w", err)
	}

	// Salva o token de ativação no DB
	tokenRecord := &domain.ActivationToken{
		AuthID:    auth.ID,
		Token:     activationToken,
		ExpiresAt: time.Now().Add(s.activationExpiration),
	}
	if err := s.activationTokenRepo.Create(tokenRecord); err != nil {
		return fmt.Errorf("falha ao salvar token de ativação: %w", err)
	}

	// Enviar email de ativação
	if err := s.emailService.SendActivationEmail(email, email, activationToken); err != nil {
		return fmt.Errorf("falha ao enviar e-mail de ativação: %w", err)
	}

	return nil
}

func (s *authService) CompleteRegistration(activationToken string, userData *CompleteRegistrationInput) error {
	// Validar token de ativação
	tokenRecord, err := s.activationTokenRepo.GetByToken(activationToken)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return errors.New("token de ativação inválido")
		}
		return fmt.Errorf("falha ao obter o token de ativação: %w", err)
	}

	// Verificar se token já foi usado
	if tokenRecord.UsedAt != nil {
		return ErrTokenAlreadyUsed
	}

	// Verificar se token expirou
	if time.Now().After(tokenRecord.ExpiresAt) {
		return ErrTokenExpired
	}

	// Buscar auth
	auth, err := s.authRepo.GetByID(tokenRecord.AuthID)
	if err != nil {
		return fmt.Errorf("falha ao obter o registro de autenticação: %w", err)
	}

	// Verificar se usuário já completou cadastro
	userExists, err := s.userRepo.ExistsByAuthID(auth.ID)
	if err != nil {
		return fmt.Errorf("falha ao verificar se o usuário já existe: %w", err)
	}
	if userExists {
		return ErrUserAlreadyCompleted
	}

	// Verificar se CPF já existe
	cpfExists, err := s.userRepo.ExistsByCPF(userData.CPF)
	if err != nil {
		return fmt.Errorf("falha ao verificar existência do CPF: %w", err)
	}
	if cpfExists {
		return repository.ErrCPFAlreadyExists
	}

	// Hash da senha
	passwordHash, err := hash.HashPassword(userData.Password)
	if err != nil {
		return fmt.Errorf("falha ao gerar o hash da senha: %w", err)
	}

	// Atualizar senha no auth
	if err := s.authRepo.UpdatePassword(auth.ID, passwordHash); err != nil {
		return fmt.Errorf("falha ao atualizar a senha: %w", err)
	}

	// Criar user com informações pessoais
	user := &domain.User{
		AuthID:    auth.ID,
		FullName:  userData.FullName,
		CPF:       userData.CPF,
		BirthDate: userData.BirthDate,
		Address:   userData.Address,
		Contact:   userData.Contact,
	}

	if err := s.userRepo.Create(user); err != nil {
		return fmt.Errorf("falha ao criar o usuário: %w", err)
	}

	// Ativar usuário
	if err := s.authRepo.UpdateStatus(auth.ID, domain.StatusActive); err != nil {
		return fmt.Errorf("falha ao ativar o usuário: %w", err)
	}

	// Marcar token como usado
	if err := s.activationTokenRepo.MarkAsUsed(tokenRecord.ID); err != nil {
		return fmt.Errorf("falha ao marcar o token como utilizado: %w", err)
	}

	// Enviar email de boas-vindas
	if err := s.emailService.SendWelcomeEmail(auth.Email, userData.FullName); err != nil {
		// Não retornar erro aqui, apenas logar
		fmt.Printf("Aviso: falha ao enviar o email de boas-vindas: %v\n", err)
	}

	return nil
}

func (s *authService) Login(email, password string) (*LoginResponse, error) {
	// Buscar auth por email
	auth, err := s.authRepo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("falha ao obter o registro de autenticação: %w", err)
	}

	// Verificar senha
	if !hash.CheckPassword(password, auth.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	// Verificar se usuário está ativo
	if auth.Status != domain.StatusActive {
		return nil, ErrUserNotActive
	}

	// Buscar informações do usuário
	user, err := s.userRepo.GetByAuthID(auth.ID)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, fmt.Errorf("falha ao obter o usuário associado: %w", err)
	}

	// Gerar JWT token
	token, err := s.jwtService.GenerateToken(auth.ID, auth.Email, string(auth.Role))
	if err != nil {
		return nil, fmt.Errorf("falha ao gerar o token JWT: %w", err)
	}

	// Montar resposta
	userWithAuth := &domain.UserWithAuth{
		Auth: *auth,
		User: user,
	}

	return &LoginResponse{
		Token: token,
		User:  userWithAuth,
	}, nil
}

func (s *authService) RequestPasswordReset(email string) error {
	// Buscar auth por email
	auth, err := s.authRepo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			// Por segurança, não revelar se o email existe
			return nil
		}
		return fmt.Errorf("falha ao obter o registro de autenticação: %w", err)
	}

	// Buscar user para pegar o nome
	user, err := s.userRepo.GetByAuthID(auth.ID)
	if err != nil {
		return fmt.Errorf("falha ao obter o usuário associado: %w", err)
	}

	// Gerar token de reset
	resetToken, err := token.GeneratePasswordResetToken()
	if err != nil {
		return fmt.Errorf("falha ao gerar o token de redefinição de senha: %w", err)
	}

	// Salvar token
	tokenRecord := &domain.PasswordResetToken{
		AuthID:    auth.ID,
		Token:     resetToken,
		ExpiresAt: time.Now().Add(s.passwordResetExpiration),
	}

	if err := s.passwordResetRepo.Create(tokenRecord); err != nil {
		return fmt.Errorf("falha ao salvar o token de redefinição de senha: %w", err)
	}

	// Enviar email
	if err := s.emailService.SendPasswordResetEmail(auth.Email, user.FullName, resetToken); err != nil {
		return fmt.Errorf("falha ao enviar o email de redefinição de senha: %w", err)
	}

	return nil
}

func (s *authService) ResetPassword(resetToken, newPassword string) error {
	// Buscar token
	tokenRecord, err := s.passwordResetRepo.GetByToken(resetToken)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return errors.New("token de redefinição inválido")
		}
		return fmt.Errorf("falha ao obter o token de redefinição: %w", err)
	}

	// Verificar se token já foi usado
	if tokenRecord.UsedAt != nil {
		return ErrTokenAlreadyUsed
	}

	// Verificar se token expirou
	if time.Now().After(tokenRecord.ExpiresAt) {
		return ErrTokenExpired
	}

	// Hash da nova senha
	passwordHash, err := hash.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("falha ao gerar o hash da senha: %w", err)
	}

	// Atualizar senha
	if err := s.authRepo.UpdatePassword(tokenRecord.AuthID, passwordHash); err != nil {
		return fmt.Errorf("falha ao atualizar a senha: %w", err)
	}

	// Marcar token como usado
	if err := s.passwordResetRepo.MarkAsUsed(tokenRecord.ID); err != nil {
		return fmt.Errorf("falha ao marcar o token como utilizado: %w", err)
	}

	return nil
}

func (s *authService) ValidateActivationToken(token string) (*domain.Auth, error) {
	tokenRecord, err := s.activationTokenRepo.GetByToken(token)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, errors.New("token de ativação inválido")
		}
		return nil, err
	}

	if tokenRecord.UsedAt != nil {
		return nil, ErrTokenAlreadyUsed
	}

	if time.Now().After(tokenRecord.ExpiresAt) {
		return nil, ErrTokenExpired
	}

	return s.authRepo.GetByID(tokenRecord.AuthID)
}

func (s *authService) ValidatePasswordResetToken(token string) (*domain.Auth, error) {
	tokenRecord, err := s.passwordResetRepo.GetByToken(token)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, errors.New("token de redefinição inválido")
		}
		return nil, err
	}

	if tokenRecord.UsedAt != nil {
		return nil, ErrTokenAlreadyUsed
	}

	if time.Now().After(tokenRecord.ExpiresAt) {
		return nil, ErrTokenExpired
	}

	return s.authRepo.GetByID(tokenRecord.AuthID)
}
