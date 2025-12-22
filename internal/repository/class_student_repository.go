package repository

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hscHeric/go-potential-api/internal/domain"
	"github.com/jmoiron/sqlx"
)

type ClassStudentRepository interface {
	AddStudent(classStudent *domain.ClassStudent) error
	RemoveStudent(classID, studentID uuid.UUID) error
	GetStudentsByClass(classID uuid.UUID) ([]uuid.UUID, error)
	GetClassesByStudent(studentID uuid.UUID) ([]uuid.UUID, error)
	CountStudentsInClass(classID uuid.UUID) (int, error)
	MarkAttendance(classID, studentID uuid.UUID, attended bool) error
	IsStudentInClass(classID, studentID uuid.UUID) (bool, error)
}

type classStudentRepository struct {
	db *sqlx.DB
}

func NewClassStudentRepository(db *sqlx.DB) ClassStudentRepository {
	return &classStudentRepository{db: db}
}

func (r *classStudentRepository) AddStudent(classStudent *domain.ClassStudent) error {
	query := `
		INSERT INTO class_students (id, class_id, student_id, added_by, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (class_id, student_id) DO NOTHING
	`

	classStudent.ID = uuid.New()
	classStudent.CreatedAt = time.Now()

	_, err := r.db.Exec(
		query,
		classStudent.ID,
		classStudent.ClassID,
		classStudent.StudentID,
		classStudent.AddedBy,
		classStudent.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to add student to class: %w", err)
	}

	return nil
}

func (r *classStudentRepository) RemoveStudent(classID, studentID uuid.UUID) error {
	query := `DELETE FROM class_students WHERE class_id = $1 AND student_id = $2`

	result, err := r.db.Exec(query, classID, studentID)
	if err != nil {
		return fmt.Errorf("failed to remove student from class: %w", err)
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

func (r *classStudentRepository) GetStudentsByClass(classID uuid.UUID) ([]uuid.UUID, error) {
	query := `SELECT student_id FROM class_students WHERE class_id = $1`

	var studentIDs []uuid.UUID
	err := r.db.Select(&studentIDs, query, classID)
	if err != nil {
		return nil, fmt.Errorf("failed to get students by class: %w", err)
	}

	return studentIDs, nil
}

func (r *classStudentRepository) GetClassesByStudent(studentID uuid.UUID) ([]uuid.UUID, error) {
	query := `SELECT class_id FROM class_students WHERE student_id = $1`

	var classIDs []uuid.UUID
	err := r.db.Select(&classIDs, query, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get classes by student: %w", err)
	}

	return classIDs, nil
}

func (r *classStudentRepository) CountStudentsInClass(classID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM class_students WHERE class_id = $1`

	var count int
	err := r.db.Get(&count, query, classID)
	if err != nil {
		return 0, fmt.Errorf("failed to count students: %w", err)
	}

	return count, nil
}

func (r *classStudentRepository) MarkAttendance(classID, studentID uuid.UUID, attended bool) error {
	query := `
		UPDATE class_students
		SET attended = $1
		WHERE class_id = $2 AND student_id = $3
	`

	result, err := r.db.Exec(query, attended, classID, studentID)
	if err != nil {
		return fmt.Errorf("failed to mark attendance: %w", err)
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

func (r *classStudentRepository) IsStudentInClass(classID, studentID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM class_students WHERE class_id = $1 AND student_id = $2)`

	var exists bool
	err := r.db.Get(&exists, query, classID, studentID)
	if err != nil {
		return false, fmt.Errorf("failed to check if student in class: %w", err)
	}

	return exists, nil
}
