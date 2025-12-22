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

type ClassRepository interface {
	Create(class *domain.Class) error
	GetByID(id uuid.UUID) (*domain.Class, error)
	GetByTeacher(teacherID uuid.UUID, startDate, endDate time.Time) ([]domain.Class, error)
	GetByStudent(studentID uuid.UUID, startDate, endDate time.Time) ([]domain.Class, error)
	Update(class *domain.Class) error
	UpdateStatus(id uuid.UUID, status domain.ClassStatus) error
	Delete(id uuid.UUID) error
	CheckTeacherAvailability(teacherID uuid.UUID, date time.Time, startTime, endTime string) (bool, error)
}

type classRepository struct {
	db *sqlx.DB
}

func NewClassRepository(db *sqlx.DB) ClassRepository {
	return &classRepository{db: db}
}

func (r *classRepository) Create(class *domain.Class) error {
	query := `
		INSERT INTO classes (
			id, teacher_id, time_slot_id, scheduled_date, start_time, end_time,
			status, title, description, class_link, material_id, created_by,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	class.ID = uuid.New()
	class.CreatedAt = time.Now()
	class.UpdatedAt = time.Now()

	_, err := r.db.Exec(
		query,
		class.ID,
		class.TeacherID,
		class.TimeSlotID,
		class.ScheduledDate,
		class.StartTime,
		class.EndTime,
		class.Status,
		class.Title,
		class.Description,
		class.ClassLink,
		class.MaterialID,
		class.CreatedBy,
		class.CreatedAt,
		class.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create class: %w", err)
	}

	return nil
}

func (r *classRepository) GetByID(id uuid.UUID) (*domain.Class, error) {
	query := `
		SELECT id, teacher_id, time_slot_id, scheduled_date, start_time, end_time,
		       status, title, description, class_link, material_id, created_by,
		       created_at, updated_at
		FROM classes
		WHERE id = $1
	`

	var class domain.Class
	err := r.db.Get(&class, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get class: %w", err)
	}

	return &class, nil
}

func (r *classRepository) GetByTeacher(teacherID uuid.UUID, startDate, endDate time.Time) ([]domain.Class, error) {
	query := `
		SELECT id, teacher_id, time_slot_id, scheduled_date, start_time, end_time,
		       status, title, description, class_link, material_id, created_by,
		       created_at, updated_at
		FROM classes
		WHERE teacher_id = $1
		  AND scheduled_date >= $2
		  AND scheduled_date <= $3
		ORDER BY scheduled_date, start_time
	`

	var classes []domain.Class
	err := r.db.Select(&classes, query, teacherID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get classes by teacher: %w", err)
	}

	return classes, nil
}

func (r *classRepository) GetByStudent(studentID uuid.UUID, startDate, endDate time.Time) ([]domain.Class, error) {
	query := `
		SELECT c.id, c.teacher_id, c.time_slot_id, c.scheduled_date, c.start_time, c.end_time,
		       c.status, c.title, c.description, c.class_link, c.material_id, c.created_by,
		       c.created_at, c.updated_at
		FROM classes c
		INNER JOIN class_students cs ON cs.class_id = c.id
		WHERE cs.student_id = $1
		  AND c.scheduled_date >= $2
		  AND c.scheduled_date <= $3
		ORDER BY c.scheduled_date, c.start_time
	`

	var classes []domain.Class
	err := r.db.Select(&classes, query, studentID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get classes by student: %w", err)
	}

	return classes, nil
}

func (r *classRepository) Update(class *domain.Class) error {
	query := `
		UPDATE classes
		SET title = $1, description = $2, class_link = $3, material_id = $4, updated_at = $5
		WHERE id = $6
	`

	class.UpdatedAt = time.Now()

	result, err := r.db.Exec(
		query,
		class.Title,
		class.Description,
		class.ClassLink,
		class.MaterialID,
		class.UpdatedAt,
		class.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update class: %w", err)
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

func (r *classRepository) UpdateStatus(id uuid.UUID, status domain.ClassStatus) error {
	query := `
		UPDATE classes
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.Exec(query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update class status: %w", err)
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

func (r *classRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM classes WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete class: %w", err)
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

func (r *classRepository) CheckTeacherAvailability(teacherID uuid.UUID, date time.Time, startTime, endTime string) (bool, error) {
	query := `
		SELECT COUNT(*) FROM classes
		WHERE teacher_id = $1
		  AND scheduled_date = $2
		  AND status != 'cancelled'
		  AND (
		    (start_time < $4 AND end_time > $3) OR
		    (start_time >= $3 AND start_time < $4)
		  )
	`

	var count int
	err := r.db.Get(&count, query, teacherID, date, startTime, endTime)
	if err != nil {
		return false, fmt.Errorf("failed to check availability: %w", err)
	}

	return count == 0, nil
}
