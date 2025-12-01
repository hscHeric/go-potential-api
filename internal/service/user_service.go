package service

import (
	"errors"

	"github.com/hscHeric/go-potential-api/internal/domain"
	"github.com/hscHeric/go-potential-api/internal/repository"
	"github.com/hscHeric/go-potential-api/pkg/utils"
)

/**
* Esse serviço é usado apenas por usuários administradores para gerenciar usuários.
* Os admins podem criar, listar, atualizar status (ativo/inativo) e deletar usuários.
* */

type UserService interface {
	CreateUser(req *domain.CreateUserRequest) (*domain.User, error)
	GetUser(id uint) (*domain.User, error)
	ListUsers() ([]domain.User, error)
	UpdateUserStatus(id uint, isActive bool) error
	DeleteUser(id uint) error
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) CreateUser(req *domain.CreateUserRequest) (*domain.User, error) {
	// Validar que role não seja admin
	if req.Role == domain.RoleAdmin {
		return nil, errors.New("cannot create admin users")
	}

	// Check if user already exists
	existingUser, _ := s.userRepo.FindByEmail(req.Email)
	if existingUser != nil {
		return nil, errors.New("email already registered")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Create user
	user := &domain.User{
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         req.Role,
		IsActive:     true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, errors.New("failed to create user")
	}

	return user, nil
}

func (s *userService) GetUser(id uint) (*domain.User, error) {
	return s.userRepo.FindByID(id)
}

func (s *userService) ListUsers() ([]domain.User, error) {
	return s.userRepo.List()
}

func (s *userService) UpdateUserStatus(id uint, isActive bool) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return err
	}

	// Não permitir desativar admin
	if user.Role == domain.RoleAdmin {
		return errors.New("cannot deactivate admin users")
	}

	user.IsActive = isActive
	return s.userRepo.Update(user)
}

func (s *userService) DeleteUser(id uint) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return err
	}

	// Não permitir deletar admin
	if user.Role == domain.RoleAdmin {
		return errors.New("cannot delete admin users")
	}

	return s.userRepo.Delete(id)
}
