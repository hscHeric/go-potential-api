package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hscHeric/go-potential-api/internal/domain"
	"github.com/hscHeric/go-potential-api/internal/repository"
	"github.com/hscHeric/go-potential-api/pkg/email"
)

type ClassService interface {
	CreateClass(createdBy uuid.UUID, input *CreateClassInput) (*domain.Class, error)
	AddStudentToClass(classID, studentID, addedBy uuid.UUID) error
	RemoveStudentFromClass(classID, studentID uuid.UUID) error
	GetClass(id uuid.UUID) (*domain.ClassWithDetails, error)
	GetTeacherClasses(teacherID uuid.UUID, startDate, endDate time.Time) ([]domain.Class, error)
	GetStudentClasses(studentID uuid.UUID, startDate, endDate time.Time) ([]domain.Class, error)
	UpdateClass(id uuid.UUID, input *UpdateClassInput) error
	UpdateClassStatus(id uuid.UUID, status domain.ClassStatus) error
	CancelClass(id uuid.UUID, cancelledBy uuid.UUID) error
	MarkAttendance(classID, studentID uuid.UUID, attended bool) error
}

type CreateClassInput struct {
	TeacherID     uuid.UUID   `json:"teacher_id" binding:"required"`
	TimeSlotID    *uuid.UUID  `json:"time_slot_id"`
	ScheduledDate time.Time   `json:"scheduled_date" binding:"required"`
	StartTime     string      `json:"start_time" binding:"required"`
	EndTime       string      `json:"end_time" binding:"required"`
	Title         *string     `json:"title"`
	Description   *string     `json:"description"`
	StudentIDs    []uuid.UUID `json:"student_ids"`
}

type UpdateClassInput struct {
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	ClassLink   *string    `json:"class_link"`
	MaterialID  *uuid.UUID `json:"material_id"`
}

type classService struct {
	classRepo        repository.ClassRepository
	classStudentRepo repository.ClassStudentRepository
	timeSlotRepo     repository.TimeSlotRepository
	userRepo         repository.UserRepository
	authRepo         repository.AuthRepository
	emailService     *email.Service
}

func NewClassService(
	classRepo repository.ClassRepository,
	classStudentRepo repository.ClassStudentRepository,
	timeSlotRepo repository.TimeSlotRepository,
	userRepo repository.UserRepository,
	authRepo repository.AuthRepository,
	emailService *email.Service,
) ClassService {
	return &classService{
		classRepo:        classRepo,
		classStudentRepo: classStudentRepo,
		timeSlotRepo:     timeSlotRepo,
		userRepo:         userRepo,
		authRepo:         authRepo,
		emailService:     emailService,
	}
}

func (s *classService) CreateClass(createdBy uuid.UUID, input *CreateClassInput) (*domain.Class, error) {
	// Validar horários
	if input.StartTime >= input.EndTime {
		return nil, errors.New("start time must be before end time")
	}

	// Verificar disponibilidade do professor
	available, err := s.classRepo.CheckTeacherAvailability(
		input.TeacherID,
		input.ScheduledDate,
		input.StartTime,
		input.EndTime,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check availability: %w", err)
	}

	if !available {
		return nil, errors.New("teacher not available at this time")
	}

	// Se tiver time slot, validar capacidade
	if input.TimeSlotID != nil {
		slot, err := s.timeSlotRepo.GetByID(*input.TimeSlotID)
		if err != nil {
			return nil, fmt.Errorf("failed to get time slot: %w", err)
		}

		if len(input.StudentIDs) > slot.MaxStudents {
			return nil, fmt.Errorf("cannot add %d students, max is %d", len(input.StudentIDs), slot.MaxStudents)
		}
	}

	// Criar aula
	class := &domain.Class{
		TeacherID:     input.TeacherID,
		TimeSlotID:    input.TimeSlotID,
		ScheduledDate: input.ScheduledDate,
		StartTime:     input.StartTime,
		EndTime:       input.EndTime,
		Status:        domain.ClassStatusScheduled,
		Title:         input.Title,
		Description:   input.Description,
		CreatedBy:     createdBy,
	}

	if err := s.classRepo.Create(class); err != nil {
		return nil, fmt.Errorf("failed to create class: %w", err)
	}

	// Adicionar alunos
	for _, studentID := range input.StudentIDs {
		classStudent := &domain.ClassStudent{
			ClassID:   class.ID,
			StudentID: studentID,
			AddedBy:   createdBy,
		}

		if err := s.classStudentRepo.AddStudent(classStudent); err != nil {
			fmt.Printf("Warning: failed to add student %s: %v\n", studentID, err)
			continue
		}

		// Enviar email para o aluno
		go s.sendClassNotificationToStudent(studentID, class)
	}

	// Enviar email para o professor
	go s.sendClassNotificationToTeacher(input.TeacherID, class, len(input.StudentIDs))

	return class, nil
}

func (s *classService) AddStudentToClass(classID, studentID, addedBy uuid.UUID) error {
	// Buscar aula
	class, err := s.classRepo.GetByID(classID)
	if err != nil {
		return err
	}

	// Verificar se aula já passou
	if class.ScheduledDate.Before(time.Now()) {
		return errors.New("cannot add student to past class")
	}

	// Verificar capacidade se tiver time slot
	if class.TimeSlotID != nil {
		slot, err := s.timeSlotRepo.GetByID(*class.TimeSlotID)
		if err != nil {
			return fmt.Errorf("failed to get time slot: %w", err)
		}

		currentCount, err := s.classStudentRepo.CountStudentsInClass(classID)
		if err != nil {
			return fmt.Errorf("failed to count students: %w", err)
		}

		if currentCount >= slot.MaxStudents {
			return errors.New("class is full")
		}
	}

	// Adicionar aluno
	classStudent := &domain.ClassStudent{
		ClassID:   classID,
		StudentID: studentID,
		AddedBy:   addedBy,
	}

	if err := s.classStudentRepo.AddStudent(classStudent); err != nil {
		return fmt.Errorf("failed to add student: %w", err)
	}

	// Enviar email de notificação
	go s.sendClassNotificationToStudent(studentID, class)

	return nil
}

func (s *classService) RemoveStudentFromClass(classID, studentID uuid.UUID) error {
	return s.classStudentRepo.RemoveStudent(classID, studentID)
}

func (s *classService) GetClass(id uuid.UUID) (*domain.ClassWithDetails, error) {
	class, err := s.classRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Buscar professor
	teacher, _ := s.userRepo.GetByAuthID(class.TeacherID)

	// Buscar alunos
	studentIDs, err := s.classStudentRepo.GetStudentsByClass(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get students: %w", err)
	}

	var students []domain.User
	for _, studentID := range studentIDs {
		student, err := s.userRepo.GetByAuthID(studentID)
		if err != nil {
			continue
		}
		students = append(students, *student)
	}

	return &domain.ClassWithDetails{
		Class:    *class,
		Teacher:  teacher,
		Students: students,
	}, nil
}

func (s *classService) GetTeacherClasses(teacherID uuid.UUID, startDate, endDate time.Time) ([]domain.Class, error) {
	return s.classRepo.GetByTeacher(teacherID, startDate, endDate)
}

func (s *classService) GetStudentClasses(studentID uuid.UUID, startDate, endDate time.Time) ([]domain.Class, error) {
	return s.classRepo.GetByStudent(studentID, startDate, endDate)
}

func (s *classService) UpdateClass(id uuid.UUID, input *UpdateClassInput) error {
	class, err := s.classRepo.GetByID(id)
	if err != nil {
		return err
	}

	// Atualizar campos
	if input.Title != nil {
		class.Title = input.Title
	}
	if input.Description != nil {
		class.Description = input.Description
	}
	if input.ClassLink != nil {
		class.ClassLink = input.ClassLink
	}
	if input.MaterialID != nil {
		class.MaterialID = input.MaterialID
	}

	return s.classRepo.Update(class)
}

func (s *classService) UpdateClassStatus(id uuid.UUID, status domain.ClassStatus) error {
	return s.classRepo.UpdateStatus(id, status)
}

func (s *classService) CancelClass(id uuid.UUID, cancelledBy uuid.UUID) error {
	class, err := s.classRepo.GetByID(id)
	if err != nil {
		return err
	}

	if err := s.classRepo.UpdateStatus(id, domain.ClassStatusCancelled); err != nil {
		return err
	}

	// Notificar alunos
	studentIDs, err := s.classStudentRepo.GetStudentsByClass(id)
	if err == nil {
		for _, studentID := range studentIDs {
			go s.sendClassCancellationToStudent(studentID, class)
		}
	}

	// Notificar professor
	go s.sendClassCancellationToTeacher(class.TeacherID, class)

	return nil
}

func (s *classService) MarkAttendance(classID, studentID uuid.UUID, attended bool) error {
	return s.classStudentRepo.MarkAttendance(classID, studentID, attended)
}

// Funções de envio de email
func (s *classService) sendClassNotificationToStudent(studentID uuid.UUID, class *domain.Class) {
	studentAuth, err := s.authRepo.GetByID(studentID)
	if err != nil {
		fmt.Printf("Failed to get student auth: %v\n", err)
		return
	}

	student, err := s.userRepo.GetByAuthID(studentID)
	if err != nil {
		fmt.Printf("Failed to get student user: %v\n", err)
		return
	}

	teacher, _ := s.userRepo.GetByAuthID(class.TeacherID)

	teacherName := "Seu Professor"
	if teacher != nil {
		teacherName = teacher.FullName
	}

	title := ""
	if class.Title != nil {
		title = *class.Title
	}

	classLink := ""
	if class.ClassLink != nil {
		classLink = *class.ClassLink
	}

	subject := "Nova Aula Agendada - Potential Idiomas"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Aula Agendada</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #3498db;">Nova Aula Agendada!</h2>
        <p>Olá, %s!</p>
        <p>Uma nova aula foi agendada para você.</p>
        <div style="background-color: #f8f9fa; padding: 20px; border-radius: 5px; margin: 20px 0;">
            <p style="margin: 5px 0;"><strong>Data:</strong> %s</p>
            <p style="margin: 5px 0;"><strong>Horário:</strong> %s - %s</p>
            <p style="margin: 5px 0;"><strong>Professor:</strong> %s</p>
            %s
            %s
        </div>
        <p style="color: #7f8c8d; font-size: 14px;">
            Não esqueça de comparecer no horário marcado!
        </p>
    </div>
</body>
</html>
	`,
		student.FullName,
		class.ScheduledDate.Format("02/01/2006"),
		class.StartTime[:5], // "14:00:00" -> "14:00"
		class.EndTime[:5],
		teacherName,
		func() string {
			if title != "" {
				return fmt.Sprintf("<p style=\"margin: 5px 0;\"><strong>Tema:</strong> %s</p>", title)
			}
			return ""
		}(),
		func() string {
			if classLink != "" {
				return fmt.Sprintf(`
            <div style="text-align: center; margin-top: 20px;">
                <a href="%s" 
                   style="background-color: #27ae60; color: white; padding: 12px 30px; 
                          text-decoration: none; border-radius: 5px; display: inline-block;">
                    Entrar na Aula
                </a>
            </div>`, classLink)
			}
			return ""
		}(),
	)

	if err := s.emailService.SendCustomEmail(studentAuth.Email, subject, body); err != nil {
		fmt.Printf("Failed to send email to student: %v\n", err)
	}
}

func (s *classService) sendClassNotificationToTeacher(teacherID uuid.UUID, class *domain.Class, studentCount int) {
	teacherAuth, err := s.authRepo.GetByID(teacherID)
	if err != nil {
		fmt.Printf("Failed to get teacher auth: %v\n", err)
		return
	}

	teacher, err := s.userRepo.GetByAuthID(teacherID)
	if err != nil {
		fmt.Printf("Failed to get teacher user: %v\n", err)
		return
	}

	title := ""
	if class.Title != nil {
		title = *class.Title
	}

	subject := "Nova Aula Agendada - Potential Idiomas"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Aula Agendada</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #3498db;">Nova Aula Criada!</h2>
        <p>Olá, %s!</p>
        <p>Uma nova aula foi agendada no seu calendário.</p>
        <div style="background-color: #f8f9fa; padding: 20px; border-radius: 5px; margin: 20px 0;">
            <p style="margin: 5px 0;"><strong>Data:</strong> %s</p>
            <p style="margin: 5px 0;"><strong>Horário:</strong> %s - %s</p>
            <p style="margin: 5px 0;"><strong>Alunos matriculados:</strong> %d</p>
            %s
        </div>
        <p style="color: #7f8c8d; font-size: 14px;">
            Acesse a plataforma para adicionar o link da aula e materiais.
        </p>
    </div>
</body>
</html>
	`,
		teacher.FullName,
		class.ScheduledDate.Format("02/01/2006"),
		class.StartTime[:5],
		class.EndTime[:5],
		studentCount,
		func() string {
			if title != "" {
				return fmt.Sprintf("<p style=\"margin: 5px 0;\"><strong>Tema:</strong> %s</p>", title)
			}
			return ""
		}(),
	)

	if err := s.emailService.SendCustomEmail(teacherAuth.Email, subject, body); err != nil {
		fmt.Printf("Failed to send email to teacher: %v\n", err)
	}
}

func (s *classService) sendClassCancellationToStudent(studentID uuid.UUID, class *domain.Class) {
	studentAuth, err := s.authRepo.GetByID(studentID)
	if err != nil {
		return
	}

	student, err := s.userRepo.GetByAuthID(studentID)
	if err != nil {
		return
	}

	teacher, _ := s.userRepo.GetByAuthID(class.TeacherID)
	teacherName := "Seu Professor"
	if teacher != nil {
		teacherName = teacher.FullName
	}

	title := ""
	if class.Title != nil {
		title = *class.Title
	}

	subject := "Aula Cancelada - Potential Idiomas"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Aula Cancelada</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #e74c3c;">Aula Cancelada</h2>
        <p>Olá, %s!</p>
        <p>Infelizmente, a aula abaixo foi cancelada:</p>
        <div style="background-color: #fee; padding: 20px; border-radius: 5px; margin: 20px 0; border-left: 4px solid #e74c3c;">
            <p style="margin: 5px 0;"><strong>Data:</strong> %s</p>
            <p style="margin: 5px 0;"><strong>Horário:</strong> %s - %s</p>
            <p style="margin: 5px 0;"><strong>Professor:</strong> %s</p>
            %s
        </div>
        <p>Entre em contato com seu professor para reagendar.</p>
        <p style="color: #7f8c8d; font-size: 14px;">
            Desculpe pelo transtorno.
        </p>
    </div>
</body>
</html>
	`,
		student.FullName,
		class.ScheduledDate.Format("02/01/2006"),
		class.StartTime[:5],
		class.EndTime[:5],
		teacherName,
		func() string {
			if title != "" {
				return fmt.Sprintf("<p style=\"margin: 5px 0;\"><strong>Tema:</strong> %s</p>", title)
			}
			return ""
		}(),
	)

	if err := s.emailService.SendCustomEmail(studentAuth.Email, subject, body); err != nil {
		fmt.Printf("Failed to send cancellation email to student: %v\n", err)
	}
}

func (s *classService) sendClassCancellationToTeacher(teacherID uuid.UUID, class *domain.Class) {
	teacherAuth, err := s.authRepo.GetByID(teacherID)
	if err != nil {
		return
	}

	teacher, err := s.userRepo.GetByAuthID(teacherID)
	if err != nil {
		return
	}

	title := ""
	if class.Title != nil {
		title = *class.Title
	}

	subject := "Aula Cancelada - Potential Idiomas"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Aula Cancelada</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #e74c3c;">Aula Cancelada</h2>
        <p>Olá, %s!</p>
        <p>A aula abaixo foi cancelada:</p>
        <div style="background-color: #fee; padding: 20px; border-radius: 5px; margin: 20px 0; border-left: 4px solid #e74c3c;">
            <p style="margin: 5px 0;"><strong>Data:</strong> %s</p>
            <p style="margin: 5px 0;"><strong>Horário:</strong> %s - %s</p>
            %s
        </div>
    </div>
</body>
</html>
	`,
		teacher.FullName,
		class.ScheduledDate.Format("02/01/2006"),
		class.StartTime[:5],
		class.EndTime[:5],
		func() string {
			if title != "" {
				return fmt.Sprintf("<p style=\"margin: 5px 0;\"><strong>Tema:</strong> %s</p>", title)
			}
			return ""
		}(),
	)

	if err := s.emailService.SendCustomEmail(teacherAuth.Email, subject, body); err != nil {
		fmt.Printf("Failed to send cancellation email to teacher: %v\n", err)
	}
}
