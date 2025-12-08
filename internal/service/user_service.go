package service

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/hscHeric/go-potential-api/internal/domain"
	"github.com/hscHeric/go-potential-api/internal/repository"
)

type UserService interface {
	GetProfile(authID uuid.UUID) (*domain.UserWithAuth, error)
	UpdateProfile(authID uuid.UUID, input *UpdateProfileInput) error
	GetByID(userID uuid.UUID) (*domain.User, error)
}

type UpdateProfileInput struct {
	FullName  string         `json:"full_name" binding:"required,min=3"`
	BirthDate string         `json:"birth_date" binding:"required"`
	Address   domain.Address `json:"address" binding:"required"`
	Contact   domain.Contact `json:"contact" binding:"required"`
}

type userService struct {
	authRepo repository.AuthRepository
	userRepo repository.UserRepository
}

func NewUserService(
	authRepo repository.AuthRepository,
	userRepo repository.UserRepository,
) UserService {
	return &userService{
		authRepo: authRepo,
		userRepo: userRepo,
	}
}

func (s *userService) GetProfile(authID uuid.UUID) (*domain.UserWithAuth, error) {
	// Buscar auth
	auth, err := s.authRepo.GetByID(authID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, errors.New("usuário não encontrado")
		}
		return nil, fmt.Errorf("falha ao obter o registro de autenticação: %w", err)
	}

	// Buscar user
	user, err := s.userRepo.GetByAuthID(authID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			// Usuário ainda não completou o cadastro
			return &domain.UserWithAuth{
				Auth: *auth,
				User: nil,
			}, nil
		}
		return nil, fmt.Errorf("falha ao obter o usuário associado: %w", err)
	}

	return &domain.UserWithAuth{
		Auth: *auth,
		User: user,
	}, nil
}

func (s *userService) UpdateProfile(authID uuid.UUID, input *UpdateProfileInput) error {
	// Verificar se user existe
	user, err := s.userRepo.GetByAuthID(authID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return errors.New("usuário não encontrado")
		}
		return fmt.Errorf("falha ao obter o usuário: %w", err)
	}

	// Atualizar campos
	user.FullName = input.FullName
	user.Address = input.Address
	user.Contact = input.Contact

	// Parse birth date se necessário
	// (você pode adicionar lógica de parse aqui se BirthDate vier como string)

	// Salvar alterações
	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("falha ao atualizar o usuário: %w", err)
	}

	return nil
}

func (s *userService) GetByID(userID uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, errors.New("usuário não encontrado")
		}
		return nil, fmt.Errorf("falha ao obter o usuário: %w", err)
	}

	return user, nil
}
