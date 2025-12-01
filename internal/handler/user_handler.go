package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hscHeric/go-potential-api/internal/domain"
	"github.com/hscHeric/go-potential-api/internal/service"
	"github.com/hscHeric/go-potential-api/pkg/response"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// CreateUser godoc
// @Summary Criar usuário (Admin only)
// @Description Admin pode criar professores ou alunos. Não é possível criar admin pela API.
// @Tags users
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body domain.CreateUserRequest true "Dados do usuário"
// @Success 201 {object} response.Response{data=domain.User} "Usuário criado com sucesso"
// @Failure 400 {object} response.Response "Erro de validação"
// @Failure 401 {object} response.Response "Não autorizado"
// @Failure 403 {object} response.Response "Sem permissão"
// @Router /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req domain.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	user, err := h.userService.CreateUser(&req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, 201, "User created successfully", user)
}

// ListUsers godoc
// @Summary Listar todos os usuários (Admin only)
// @Description Retorna lista de todos os usuários cadastrados
// @Tags users
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response{data=[]domain.User} "Lista de usuários"
// @Failure 401 {object} response.Response "Não autorizado"
// @Failure 403 {object} response.Response "Sem permissão"
// @Failure 500 {object} response.Response "Erro interno"
// @Router /users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	users, err := h.userService.ListUsers()
	if err != nil {
		response.Error(c, 500, "Failed to fetch users")
		return
	}

	response.Success(c, 200, "Users retrieved successfully", users)
}

// GetUser godoc
// @Summary Obter usuário por ID (Admin only)
// @Description Retorna detalhes de um usuário específico
// @Tags users
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "ID do usuário"
// @Success 200 {object} response.Response{data=domain.User} "Detalhes do usuário"
// @Failure 400 {object} response.Response "ID inválido"
// @Failure 401 {object} response.Response "Não autorizado"
// @Failure 403 {object} response.Response "Sem permissão"
// @Failure 404 {object} response.Response "Usuário não encontrado"
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, 400, "Invalid user ID")
		return
	}

	user, err := h.userService.GetUser(uint(id))
	if err != nil {
		response.Error(c, 404, "User not found")
		return
	}

	response.Success(c, 200, "User retrieved successfully", user)
}

// UpdateUserStatus godoc
// @Summary Ativar/Desativar usuário (Admin only)
// @Description Atualiza o status de ativo/inativo de um usuário. Não é possível desativar admins.
// @Tags users
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "ID do usuário"
// @Param request body object{is_active=bool} true "Status"
// @Success 200 {object} response.Response "Status atualizado com sucesso"
// @Failure 400 {object} response.Response "Erro de validação"
// @Failure 401 {object} response.Response "Não autorizado"
// @Failure 403 {object} response.Response "Sem permissão"
// @Router /users/{id}/status [patch]
func (h *UserHandler) UpdateUserStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, 400, "Invalid user ID")
		return
	}

	var req struct {
		IsActive bool `json:"is_active" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	if err := h.userService.UpdateUserStatus(uint(id), req.IsActive); err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, 200, "User status updated successfully", nil)
}

// DeleteUser godoc
// @Summary Deletar usuário (Admin only)
// @Description Remove um usuário do sistema. Não é possível deletar admins.
// @Tags users
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "ID do usuário"
// @Success 200 {object} response.Response "Usuário deletado com sucesso"
// @Failure 400 {object} response.Response "Erro ao deletar"
// @Failure 401 {object} response.Response "Não autorizado"
// @Failure 403 {object} response.Response "Sem permissão"
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, 400, "Invalid user ID")
		return
	}

	if err := h.userService.DeleteUser(uint(id)); err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, 200, "User deleted successfully", nil)
}
