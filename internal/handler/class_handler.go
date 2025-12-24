package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hscHeric/go-potential-api/internal/domain"
	"github.com/hscHeric/go-potential-api/internal/middleware"
	"github.com/hscHeric/go-potential-api/internal/service"
	"github.com/hscHeric/go-potential-api/pkg/validator"
)

type ClassHandler struct {
	classService service.ClassService
}

func NewClassHandler(classService service.ClassService) *ClassHandler {
	return &ClassHandler{
		classService: classService,
	}
}

// CreateClass godoc
// @Summary Create class
// @Description Create a new class (Teacher or Admin)
// @Tags classes
// @Accept json
// @Produce json
// @Param request body service.CreateClassInput true "Class data"
// @Success 201 {object} domain.Class
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/classes [post]
// @Security BearerAuth
func (h *ClassHandler) CreateClass(c *gin.Context) {
	authID, err := middleware.GetAuthID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Unauthorized",
		})
		return
	}

	var req service.CreateClassInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
		})
		return
	}

	if err := validator.Validate(req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Validation failed",
			Details: validator.FormatValidationErrors(err),
		})
		return
	}

	class, err := h.classService.CreateClass(authID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, class)
}

// AddStudentToClass godoc
// @Summary Add student to class
// @Description Add a student to an existing class
// @Tags classes
// @Accept json
// @Produce json
// @Param id path string true "Class ID"
// @Param request body AddStudentRequest true "Student ID"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/classes/{id}/students [post]
// @Security BearerAuth
func (h *ClassHandler) AddStudentToClass(c *gin.Context) {
	authID, err := middleware.GetAuthID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Unauthorized",
		})
		return
	}

	classID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid class ID",
		})
		return
	}

	var req AddStudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
		})
		return
	}

	if err := h.classService.AddStudentToClass(classID, req.StudentID, authID); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{
		Message: "Student added to class successfully",
	})
}

// RemoveStudentFromClass godoc
// @Summary Remove student from class
// @Description Remove a student from a class
// @Tags classes
// @Produce json
// @Param id path string true "Class ID"
// @Param student_id path string true "Student ID"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/classes/{id}/students/{student_id} [delete]
// @Security BearerAuth
func (h *ClassHandler) RemoveStudentFromClass(c *gin.Context) {
	classID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid class ID",
		})
		return
	}

	studentID, err := uuid.Parse(c.Param("student_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid student ID",
		})
		return
	}

	if err := h.classService.RemoveStudentFromClass(classID, studentID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to remove student",
		})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{
		Message: "Student removed from class successfully",
	})
}

// GetClass godoc
// @Summary Get class details
// @Description Get detailed information about a class
// @Tags classes
// @Produce json
// @Param id path string true "Class ID"
// @Success 200 {object} domain.ClassWithDetails
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/classes/{id} [get]
// @Security BearerAuth
func (h *ClassHandler) GetClass(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid class ID",
		})
		return
	}

	class, err := h.classService.GetClass(id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error: "Class not found",
		})
		return
	}

	c.JSON(http.StatusOK, class)
}

// GetMyClasses godoc
// @Summary Get my classes
// @Description Get classes for authenticated user (as teacher or student)
// @Tags classes
// @Produce json
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Success 200 {array} domain.Class
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/classes/me [get]
// @Security BearerAuth
func (h *ClassHandler) GetMyClasses(c *gin.Context) {
	authID, err := middleware.GetAuthID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Unauthorized",
		})
		return
	}

	role, _ := middleware.GetRole(c)

	// Parse dates
	startDate := time.Now().AddDate(0, 0, -30) // Default: 30 days ago
	endDate := time.Now().AddDate(0, 0, 30)    // Default: 30 days ahead

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = parsed
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = parsed
		}
	}

	var classes []domain.Class
	if role == string(domain.RoleTeacher) {
		classes, err = h.classService.GetTeacherClasses(authID, startDate, endDate)
	} else {
		classes, err = h.classService.GetStudentClasses(authID, startDate, endDate)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to get classes",
		})
		return
	}

	c.JSON(http.StatusOK, classes)
}

// UpdateClass godoc
// @Summary Update class
// @Description Update class details (title, description, link, material)
// @Tags classes
// @Accept json
// @Produce json
// @Param id path string true "Class ID"
// @Param request body service.UpdateClassInput true "Class data"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/classes/{id} [put]
// @Security BearerAuth
func (h *ClassHandler) UpdateClass(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid class ID",
		})
		return
	}

	var req service.UpdateClassInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
		})
		return
	}

	if err := h.classService.UpdateClass(id, &req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{
		Message: "Class updated successfully",
	})
}

// CancelClass godoc
// @Summary Cancel class
// @Description Cancel a scheduled class
// @Tags classes
// @Produce json
// @Param id path string true "Class ID"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/classes/{id}/cancel [patch]
// @Security BearerAuth
func (h *ClassHandler) CancelClass(c *gin.Context) {
	authID, err := middleware.GetAuthID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Unauthorized",
		})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid class ID",
		})
		return
	}

	if err := h.classService.CancelClass(id, authID); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{
		Message: "Class cancelled successfully",
	})
}

// MarkAttendance godoc
// @Summary Mark student attendance
// @Description Mark if a student attended the class
// @Tags classes
// @Accept json
// @Produce json
// @Param id path string true "Class ID"
// @Param student_id path string true "Student ID"
// @Param request body AttendanceRequest true "Attendance data"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/classes/{id}/students/{student_id}/attendance [patch]
// @Security BearerAuth
func (h *ClassHandler) MarkAttendance(c *gin.Context) {
	classID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid class ID",
		})
		return
	}

	studentID, err := uuid.Parse(c.Param("student_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid student ID",
		})
		return
	}

	var req AttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
		})
		return
	}

	if err := h.classService.MarkAttendance(classID, studentID, req.Attended); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to mark attendance",
		})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{
		Message: "Attendance marked successfully",
	})
}
