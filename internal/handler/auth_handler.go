package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/hscHeric/go-potential-api/internal/domain"
	"github.com/hscHeric/go-potential-api/internal/service"
	"github.com/hscHeric/go-potential-api/pkg/response"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register godoc
// @Summary Registrar novo usuário
// @Description Cria uma nova conta de aluno (student). Qualquer pessoa pode se registrar.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.RegisterRequest true "Dados de registro"
// @Success 201 {object} response.Response{data=domain.User} "Usuário criado com sucesso"
// @Failure 400 {object} response.Response "Erro de validação"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	user, err := h.authService.Register(&req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, 201, "Usuario registrado com sucesso", user)
}

// Login godoc
// @Summary Login
// @Description Autentica um usuário e retorna um token JWT
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.LoginRequest true "Credenciais de login"
// @Success 200 {object} response.Response{data=domain.LoginResponse} "Login realizado com sucesso"
// @Failure 401 {object} response.Response "Credenciais inválidas"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	loginResponse, err := h.authService.Login(&req)
	if err != nil {
		response.Error(c, 401, err.Error())
		return
	}

	response.Success(c, 200, "Login realizado com sucesso", loginResponse)
}

// Me godoc
// @Summary Obter perfil do usuário autenticado
// @Description Retorna informações do usuário logado
// @Tags auth
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response "Informações do usuário"
// @Failure 401 {object} response.Response "Não autorizado"
// @Router /me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userEmail, _ := c.Get("user_email")
	userRole, _ := c.Get("user_role")

	response.Success(c, 200, "User info", gin.H{
		"id":    userID,
		"email": userEmail,
		"role":  userRole,
	})
}
