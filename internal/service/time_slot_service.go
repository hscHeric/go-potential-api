package service

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/hscHeric/go-potential-api/internal/domain"
	"github.com/hscHeric/go-potential-api/internal/repository"
)

type TimeSlotService interface {
	CreateTimeSlot(teacherID uuid.UUID, input *CreateTimeSlotInput) (*domain.TimeSlot, error)
	GetTimeSlot(id uuid.UUID) (*domain.TimeSlot, error)
	GetTeacherTimeSlots(teacherID uuid.UUID) ([]domain.TimeSlot, error)
	GetAvailableSlots(teacherID uuid.UUID, dayOfWeek domain.DayOfWeek) ([]domain.TimeSlot, error)
	UpdateTimeSlot(id uuid.UUID, input *UpdateTimeSlotInput) error
	DeleteTimeSlot(id uuid.UUID) error
	ToggleAvailability(id uuid.UUID, isAvailable bool) error
}

type CreateTimeSlotInput struct {
	DayOfWeek   domain.DayOfWeek `json:"day_of_week" binding:"required,min=0,max=6"`
	StartTime   string           `json:"start_time" binding:"required"`
	EndTime     string           `json:"end_time" binding:"required"`
	MaxStudents int              `json:"max_students" binding:"required,min=1"`
}

type UpdateTimeSlotInput struct {
	DayOfWeek   domain.DayOfWeek `json:"day_of_week" binding:"required,min=0,max=6"`
	StartTime   string           `json:"start_time" binding:"required"`
	EndTime     string           `json:"end_time" binding:"required"`
	MaxStudents int              `json:"max_students" binding:"required,min=1"`
	IsAvailable bool             `json:"is_available"`
}

type timeSlotService struct {
	timeSlotRepo repository.TimeSlotRepository
	authRepo     repository.AuthRepository
}

func NewTimeSlotService(
	timeSlotRepo repository.TimeSlotRepository,
	authRepo repository.AuthRepository,
) TimeSlotService {
	return &timeSlotService{
		timeSlotRepo: timeSlotRepo,
		authRepo:     authRepo,
	}
}

func (s *timeSlotService) CreateTimeSlot(teacherID uuid.UUID, input *CreateTimeSlotInput) (*domain.TimeSlot, error) {
	// Verificar se o usuário é professor
	auth, err := s.authRepo.GetByID(teacherID)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth: %w", err)
	}

	if auth.Role != domain.RoleTeacher {
		return nil, errors.New("only teachers can create time slots")
	}

	// Validar horários
	if input.StartTime >= input.EndTime {
		return nil, errors.New("start time must be before end time")
	}

	timeSlot := &domain.TimeSlot{
		TeacherID:   teacherID,
		DayOfWeek:   input.DayOfWeek,
		StartTime:   input.StartTime,
		EndTime:     input.EndTime,
		MaxStudents: input.MaxStudents,
		IsAvailable: true,
	}

	if err := s.timeSlotRepo.Create(timeSlot); err != nil {
		return nil, fmt.Errorf("failed to create time slot: %w", err)
	}

	return timeSlot, nil
}

func (s *timeSlotService) GetTimeSlot(id uuid.UUID) (*domain.TimeSlot, error) {
	return s.timeSlotRepo.GetByID(id)
}

func (s *timeSlotService) GetTeacherTimeSlots(teacherID uuid.UUID) ([]domain.TimeSlot, error) {
	return s.timeSlotRepo.GetByTeacher(teacherID)
}

func (s *timeSlotService) GetAvailableSlots(teacherID uuid.UUID, dayOfWeek domain.DayOfWeek) ([]domain.TimeSlot, error) {
	return s.timeSlotRepo.GetByTeacherAndDay(teacherID, dayOfWeek)
}

func (s *timeSlotService) UpdateTimeSlot(id uuid.UUID, input *UpdateTimeSlotInput) error {
	// Buscar time slot existente
	slot, err := s.timeSlotRepo.GetByID(id)
	if err != nil {
		return err
	}

	// Validar horários
	if input.StartTime >= input.EndTime {
		return errors.New("start time must be before end time")
	}

	// Atualizar campos
	slot.DayOfWeek = input.DayOfWeek
	slot.StartTime = input.StartTime
	slot.EndTime = input.EndTime
	slot.MaxStudents = input.MaxStudents
	slot.IsAvailable = input.IsAvailable

	return s.timeSlotRepo.Update(slot)
}

func (s *timeSlotService) DeleteTimeSlot(id uuid.UUID) error {
	return s.timeSlotRepo.Delete(id)
}

func (s *timeSlotService) ToggleAvailability(id uuid.UUID, isAvailable bool) error {
	return s.timeSlotRepo.ToggleAvailability(id, isAvailable)
}
