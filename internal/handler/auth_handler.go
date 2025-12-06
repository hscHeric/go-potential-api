// Package handler controllers para as rotas
package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hscHeric/go-potential-api/internal/repository"
	"github.com/hscHeric/go-potential-api/internal/service"
	"github.com/hscHeric/go-potential-api/pkg/validator"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// CreateInvitation godoc
// @Summary Criar convite para novo usuário
// @Description Admin cria um convite informando email e papel (role)
// @Tags Autenticação
// @Accept json
// @Produce json
// @Param request body CreateInvitationRequest true "Dados do convite"
// @Success 201 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /api/v1/invitations [post]
// @Security BearerAuth
func (h *AuthHandler) CreateInvitation(c *gin.Context) {
	var req CreateInvitationRequest

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

	if err := h.authService.CreateInvitation(req.Email, req.Role); err != nil {
		if errors.Is(err, repository.ErrEmailAlreadyExists) {
			c.JSON(http.StatusConflict, ErrorResponse{
				Error: "O e-mail informado já está em uso",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Erro ao criar convite",
		})
		return
	}

	c.JSON(http.StatusCreated, MessageResponse{
		Message: "Convite enviado com sucesso",
	})
}

// ValidateActivationToken godoc
// @Summary Validar token de ativação
// @Description Verifica se o token de ativação é válido
// @Tags Autenticação
// @Accept json
// @Produce json
// @Param token query string true "Token de ativação"
// @Success 200 {object} ValidateTokenResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/auth/validate-activation-token [get]
func (h *AuthHandler) ValidateActivationToken(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "O token é obrigatório",
		})
		return
	}

	auth, err := h.authService.ValidateActivationToken(token)
	if err != nil {
		if errors.Is(err, service.ErrTokenExpired) {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "O token expirou",
			})
			return
		}
		if errors.Is(err, service.ErrTokenAlreadyUsed) {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "O token já foi utilizado",
			})
			return
		}

		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Token inválido",
		})
		return
	}

	c.JSON(http.StatusOK, ValidateTokenResponse{
		Valid: true,
		Email: auth.Email,
		Role:  string(auth.Role),
	})
}

// CompleteRegistration godoc
// @Summary Completar cadastro de usuário
// @Description Usuário finaliza seu cadastro informando dados pessoais e senha
// @Tags Autenticação
// @Accept json
// @Produce json
// @Param request body CompleteRegistrationRequest true "Dados de cadastro"
// @Success 201 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/auth/complete-registration [post]
func (h *AuthHandler) CompleteRegistration(c *gin.Context) {
	var req CompleteRegistrationRequest

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

	userData := &service.CompleteRegistrationInput{
		FullName:  req.FullName,
		CPF:       req.CPF,
		BirthDate: req.BirthDate,
		Address:   req.Address,
		Contact:   req.Contact,
		Password:  req.Password,
	}

	if err := h.authService.CompleteRegistration(req.Token, userData); err != nil {
		if errors.Is(err, service.ErrTokenExpired) {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "O token expirou",
			})
			return
		}
		if errors.Is(err, service.ErrTokenAlreadyUsed) {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "O token já foi utilizado",
			})
			return
		}
		if errors.Is(err, service.ErrUserAlreadyCompleted) {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "O usuário já completou o cadastro anteriormente",
			})
			return
		}
		if errors.Is(err, repository.ErrCPFAlreadyExists) {
			c.JSON(http.StatusConflict, ErrorResponse{
				Error: "O CPF informado já está cadastrado",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Erro ao completar cadastro",
		})
		return
	}

	c.JSON(http.StatusCreated, MessageResponse{
		Message: "Cadastro concluído com sucesso",
	})
}

// Login godoc
// @Summary Login do usuário
// @Description Autentica usuário com email e senha
// @Tags Autenticação
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Credenciais de login"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest

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

	response, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error: "E-mail ou senha inválidos",
			})
			return
		}
		if errors.Is(err, service.ErrUserNotActive) {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error: "O usuário não está ativo",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Erro ao realizar login",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// RequestPasswordReset godoc
// @Summary Solicitar redefinição de senha
// @Description Envia email com instruções para redefinir senha
// @Tags Autenticação
// @Accept json
// @Produce json
// @Param request body RequestPasswordResetRequest true "E-mail"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/auth/request-password-reset [post]
func (h *AuthHandler) RequestPasswordReset(c *gin.Context) {
	var req RequestPasswordResetRequest

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

	if err := h.authService.RequestPasswordReset(req.Email); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Erro ao solicitar redefinição de senha",
		})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{
		Message: "Se o email existir, um link para redefinição de senha foi enviado",
	})
}

// ResetPassword godoc
// @Summary Redefinir senha
// @Description Redefine a senha do usuário utilizando um token
// @Tags Autenticação
// @Accept json
// @Produce json
// @Param request body ResetPasswordRequest true "Dados para redefinição"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest

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

	if err := h.authService.ResetPassword(req.Token, req.NewPassword); err != nil {
		if errors.Is(err, service.ErrTokenExpired) {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "O token expirou",
			})
			return
		}
		if errors.Is(err, service.ErrTokenAlreadyUsed) {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "O token já foi utilizado",
			})
			return
		}

		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Token inválido",
		})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{
		Message: "Senha redefinida com sucesso",
	})
}
