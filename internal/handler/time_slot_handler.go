package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hscHeric/go-potential-api/internal/domain"
	"github.com/hscHeric/go-potential-api/internal/middleware"
	"github.com/hscHeric/go-potential-api/internal/service"
	"github.com/hscHeric/go-potential-api/pkg/validator"
)

type TimeSlotHandler struct {
	timeSlotService service.TimeSlotService
}

func NewTimeSlotHandler(timeSlotService service.TimeSlotService) *TimeSlotHandler {
	return &TimeSlotHandler{
		timeSlotService: timeSlotService,
	}
}

// CreateTimeSlot godoc
// @Summary Create time slot
// @Description Teacher creates a recurring time slot
// @Tags time-slots
// @Accept json
// @Produce json
// @Param request body service.CreateTimeSlotInput true "Time slot data"
// @Success 201 {object} domain.TimeSlot
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/time-slots [post]
// @Security BearerAuth
func (h *TimeSlotHandler) CreateTimeSlot(c *gin.Context) {
	authID, err := middleware.GetAuthID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Unauthorized",
		})
		return
	}

	var req service.CreateTimeSlotInput
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

	timeSlot, err := h.timeSlotService.CreateTimeSlot(authID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, timeSlot)
}

// GetMyTimeSlots godoc
// @Summary Get teacher's time slots
// @Description Get all time slots for the authenticated teacher
// @Tags time-slots
// @Produce json
// @Success 200 {array} domain.TimeSlot
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/time-slots/me [get]
// @Security BearerAuth
func (h *TimeSlotHandler) GetMyTimeSlots(c *gin.Context) {
	authID, err := middleware.GetAuthID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Unauthorized",
		})
		return
	}

	timeSlots, err := h.timeSlotService.GetTeacherTimeSlots(authID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to get time slots",
		})
		return
	}

	c.JSON(http.StatusOK, timeSlots)
}

// GetTeacherTimeSlots godoc
// @Summary Get teacher's time slots by ID
// @Description Get all time slots for a specific teacher
// @Tags time-slots
// @Produce json
// @Param teacher_id path string true "Teacher ID"
// @Success 200 {array} domain.TimeSlot
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/time-slots/teacher/{teacher_id} [get]
// @Security BearerAuth
func (h *TimeSlotHandler) GetTeacherTimeSlots(c *gin.Context) {
	teacherID, err := uuid.Parse(c.Param("teacher_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid teacher ID",
		})
		return
	}

	timeSlots, err := h.timeSlotService.GetTeacherTimeSlots(teacherID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to get time slots",
		})
		return
	}

	c.JSON(http.StatusOK, timeSlots)
}

// GetAvailableSlots godoc
// @Summary Get available slots
// @Description Get available time slots for a teacher on a specific day
// @Tags time-slots
// @Produce json
// @Param teacher_id path string true "Teacher ID"
// @Param day_of_week query int true "Day of week (0=Sunday, 1=Monday, ..., 6=Saturday)"
// @Success 200 {array} domain.TimeSlot
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/time-slots/teacher/{teacher_id}/available [get]
// @Security BearerAuth
func (h *TimeSlotHandler) GetAvailableSlots(c *gin.Context) {
	teacherID, err := uuid.Parse(c.Param("teacher_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid teacher ID",
		})
		return
	}

	dayOfWeekStr := c.Query("day_of_week")
	dayOfWeekInt, err := strconv.Atoi(dayOfWeekStr)
	if err != nil || dayOfWeekInt < 0 || dayOfWeekInt > 6 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid day_of_week (must be 0-6)",
		})
		return
	}

	dayOfWeek := domain.DayOfWeek(dayOfWeekInt)

	timeSlots, err := h.timeSlotService.GetAvailableSlots(teacherID, dayOfWeek)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to get available slots",
		})
		return
	}

	c.JSON(http.StatusOK, timeSlots)
}

// UpdateTimeSlot godoc
// @Summary Update time slot
// @Description Update a time slot
// @Tags time-slots
// @Accept json
// @Produce json
// @Param id path string true "Time Slot ID"
// @Param request body service.UpdateTimeSlotInput true "Time slot data"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/time-slots/{id} [put]
// @Security BearerAuth
func (h *TimeSlotHandler) UpdateTimeSlot(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid time slot ID",
		})
		return
	}

	var req service.UpdateTimeSlotInput
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

	if err := h.timeSlotService.UpdateTimeSlot(id, &req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{
		Message: "Time slot updated successfully",
	})
}

// DeleteTimeSlot godoc
// @Summary Delete time slot
// @Description Delete a time slot
// @Tags time-slots
// @Produce json
// @Param id path string true "Time Slot ID"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/time-slots/{id} [delete]
// @Security BearerAuth
func (h *TimeSlotHandler) DeleteTimeSlot(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid time slot ID",
		})
		return
	}

	if err := h.timeSlotService.DeleteTimeSlot(id); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to delete time slot",
		})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{
		Message: "Time slot deleted successfully",
	})
}

// ToggleTimeSlotAvailability godoc
// @Summary Toggle time slot availability
// @Description Enable or disable a time slot
// @Tags time-slots
// @Accept json
// @Produce json
// @Param id path string true "Time Slot ID"
// @Param request body ToggleAvailabilityRequest true "Availability status"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/time-slots/{id}/toggle [patch]
// @Security BearerAuth
func (h *TimeSlotHandler) ToggleTimeSlotAvailability(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid time slot ID",
		})
		return
	}

	var req ToggleAvailabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
		})
		return
	}

	if err := h.timeSlotService.ToggleAvailability(id, req.IsAvailable); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to toggle availability",
		})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{
		Message: "Availability updated successfully",
	})
}
