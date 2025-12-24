package handler

import "github.com/google/uuid"

// ToggleAvailabilityRequest para ativar/desativar time slot
type ToggleAvailabilityRequest struct {
	IsAvailable bool `json:"is_available"`
}

// AddStudentRequest para adicionar aluno a uma aula
type AddStudentRequest struct {
	StudentID uuid.UUID `json:"student_id" binding:"required"`
}

// AttendanceRequest para marcar presença
type AttendanceRequest struct {
	Attended bool `json:"attended"`
}
