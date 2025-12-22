package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hscHeric/go-potential-api/internal/domain"
	"github.com/jmoiron/sqlx"
)

type TimeSlotRepository interface {
	Create(slot *domain.TimeSlot) error
	GetByID(id uuid.UUID) (*domain.TimeSlot, error)
	GetByTeacher(teacherID uuid.UUID) ([]domain.TimeSlot, error)
	GetByTeacherAndDay(teacherID uuid.UUID, dayOfWeek domain.DayOfWeek) ([]domain.TimeSlot, error)
	Update(slot *domain.TimeSlot) error
	Delete(id uuid.UUID) error
	ToggleAvailability(id uuid.UUID, isAvailable bool) error
}

type timeSlotRepository struct {
	db *sqlx.DB
}

func NewTimeSlotRepository(db *sqlx.DB) TimeSlotRepository {
	return &timeSlotRepository{db: db}
}

func (r *timeSlotRepository) Create(slot *domain.TimeSlot) error {
	query := `
		INSERT INTO time_slots (id, teacher_id, day_of_week, start_time, end_time, max_students, is_available, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	slot.ID = uuid.New()
	slot.CreatedAt = time.Now()
	slot.UpdatedAt = time.Now()

	_, err := r.db.Exec(
		query,
		slot.ID,
		slot.TeacherID,
		slot.DayOfWeek,
		slot.StartTime,
		slot.EndTime,
		slot.MaxStudents,
		slot.IsAvailable,
		slot.CreatedAt,
		slot.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create time slot: %w", err)
	}

	return nil
}

func (r *timeSlotRepository) GetByID(id uuid.UUID) (*domain.TimeSlot, error) {
	query := `
		SELECT id, teacher_id, day_of_week, start_time, end_time, max_students, is_available, created_at, updated_at
		FROM time_slots
		WHERE id = $1
	`

	var slot domain.TimeSlot
	err := r.db.Get(&slot, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get time slot: %w", err)
	}

	return &slot, nil
}

func (r *timeSlotRepository) GetByTeacher(teacherID uuid.UUID) ([]domain.TimeSlot, error) {
	query := `
		SELECT id, teacher_id, day_of_week, start_time, end_time, max_students, is_available, created_at, updated_at
		FROM time_slots
		WHERE teacher_id = $1
		ORDER BY day_of_week, start_time
	`

	var slots []domain.TimeSlot
	err := r.db.Select(&slots, query, teacherID)
	if err != nil {
		return nil, fmt.Errorf("failed to get time slots by teacher: %w", err)
	}

	return slots, nil
}

func (r *timeSlotRepository) GetByTeacherAndDay(teacherID uuid.UUID, dayOfWeek domain.DayOfWeek) ([]domain.TimeSlot, error) {
	query := `
		SELECT id, teacher_id, day_of_week, start_time, end_time, max_students, is_available, created_at, updated_at
		FROM time_slots
		WHERE teacher_id = $1 AND day_of_week = $2 AND is_available = true
		ORDER BY start_time
	`

	var slots []domain.TimeSlot
	err := r.db.Select(&slots, query, teacherID, dayOfWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to get time slots by teacher and day: %w", err)
	}

	return slots, nil
}

func (r *timeSlotRepository) Update(slot *domain.TimeSlot) error {
	query := `
		UPDATE time_slots
		SET day_of_week = $1, start_time = $2, end_time = $3, max_students = $4, is_available = $5, updated_at = $6
		WHERE id = $7
	`

	slot.UpdatedAt = time.Now()

	result, err := r.db.Exec(
		query,
		slot.DayOfWeek,
		slot.StartTime,
		slot.EndTime,
		slot.MaxStudents,
		slot.IsAvailable,
		slot.UpdatedAt,
		slot.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update time slot: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *timeSlotRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM time_slots WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete time slot: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *timeSlotRepository) ToggleAvailability(id uuid.UUID, isAvailable bool) error {
	query := `
		UPDATE time_slots
		SET is_available = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.Exec(query, isAvailable, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to toggle availability: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
