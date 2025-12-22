package domain

import (
	"time"

	"github.com/google/uuid"
)

// DayOfWeek representa o dia da semana
type DayOfWeek int

const (
	Sunday DayOfWeek = iota
	Monday
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
)

// TimeSlot representa um horário disponível recorrente do professor
type TimeSlot struct {
	ID          uuid.UUID `db:"id" json:"id"`
	TeacherID   uuid.UUID `db:"teacher_id" json:"teacher_id"`
	DayOfWeek   DayOfWeek `db:"day_of_week" json:"day_of_week"`
	StartTime   string    `db:"start_time" json:"start_time"` // Format: "14:00:00"
	EndTime     string    `db:"end_time" json:"end_time"`     // Format: "15:00:00"
	MaxStudents int       `db:"max_students" json:"max_students"`
	IsAvailable bool      `db:"is_available" json:"is_available"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// ClassStatus representa o status de uma aula
type ClassStatus string

const (
	ClassStatusScheduled ClassStatus = "scheduled"
	ClassStatusCompleted ClassStatus = "completed"
	ClassStatusCancelled ClassStatus = "cancelled"
	ClassStatusNoShow    ClassStatus = "no_show"
)

// Class representa uma aula agendada
type Class struct {
	ID            uuid.UUID   `db:"id" json:"id"`
	TeacherID     uuid.UUID   `db:"teacher_id" json:"teacher_id"`
	TimeSlotID    *uuid.UUID  `db:"time_slot_id" json:"time_slot_id,omitempty"`
	ScheduledDate time.Time   `db:"scheduled_date" json:"scheduled_date"`
	StartTime     string      `db:"start_time" json:"start_time"` // Format: "14:00:00"
	EndTime       string      `db:"end_time" json:"end_time"`     // Format: "15:00:00"
	Status        ClassStatus `db:"status" json:"status"`
	Title         *string     `db:"title" json:"title,omitempty"`
	Description   *string     `db:"description" json:"description,omitempty"`
	ClassLink     *string     `db:"class_link" json:"class_link,omitempty"`
	MaterialID    *uuid.UUID  `db:"material_id" json:"material_id,omitempty"`
	CreatedBy     uuid.UUID   `db:"created_by" json:"created_by"`
	CreatedAt     time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time   `db:"updated_at" json:"updated_at"`
}

// ClassStudent representa a relação entre aula e aluno
type ClassStudent struct {
	ID        uuid.UUID `db:"id" json:"id"`
	ClassID   uuid.UUID `db:"class_id" json:"class_id"`
	StudentID uuid.UUID `db:"student_id" json:"student_id"`
	AddedBy   uuid.UUID `db:"added_by" json:"added_by"`
	Attended  *bool     `db:"attended" json:"attended,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// ClassWithDetails combina Class com detalhes
type ClassWithDetails struct {
	Class
	Teacher  *User  `json:"teacher,omitempty"`
	Students []User `json:"students,omitempty"`
	Material *File  `json:"material,omitempty"`
}
