package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hscHeric/go-potential-api/internal/middleware"
	"github.com/hscHeric/go-potential-api/internal/service"
	"github.com/hscHeric/go-potential-api/pkg/validator"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetProfile godoc
// @Summary Obter perfil do usuário
// @Description Retorna o perfil do usuário autenticado
// @Tags usuários
// @Produce json
// @Success 200 {object} domain.UserWithAuth
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/users/me [get]
// @Security BearerAuth
func (h *UserHandler) GetProfile(c *gin.Context) {
	authID, err := middleware.GetAuthID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Não autorizado",
		})
		return
	}

	profile, err := h.userService.GetProfile(authID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Falha ao obter o perfil",
		})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// UpdateProfile godoc
// @Summary Atualizar perfil do usuário
// @Description Atualiza as informações do perfil do usuário autenticado
// @Tags usuários
// @Accept json
// @Produce json
// @Param request body UpdateProfileRequest true "Dados do perfil"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/users/me [put]
// @Security BearerAuth
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	authID, err := middleware.GetAuthID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Não autorizado",
		})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Formato de requisição inválido",
		})
		return
	}

	if err := validator.Validate(req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Falha na validação dos dados",
			Details: validator.FormatValidationErrors(err),
		})
		return
	}

	input := &service.UpdateProfileInput{
		FullName:  req.FullName,
		BirthDate: req.BirthDate,
		Address:   req.Address,
		Contact:   req.Contact,
	}

	if err := h.userService.UpdateProfile(authID, input); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Falha ao atualizar o perfil",
		})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{
		Message: "Perfil atualizado com sucesso",
	})
}

// UpdateProfilePicture godoc
// @Summary Atualizar foto de perfil
// @Description Atualiza a foto de perfil do usuário autenticado
// @Tags usuários
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Foto de perfil"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/users/me/profile-picture [put]
// @Security BearerAuth
func (h *UserHandler) UpdateProfilePicture(c *gin.Context) {
	// Aqui você pode implementar a lógica de upload
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error: "Upload de foto de perfil ainda não foi implementado",
	})
}
