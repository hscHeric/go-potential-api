package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hscHeric/go-potential-api/internal/domain"
	"github.com/hscHeric/go-potential-api/internal/middleware"
	"github.com/hscHeric/go-potential-api/internal/service"
)

type DocumentHandler struct {
	documentService service.DocumentService
}

func NewDocumentHandler(documentService service.DocumentService) *DocumentHandler {
	return &DocumentHandler{
		documentService: documentService,
	}
}

// UploadDocument godoc
// @Summary Enviar documento
// @Description Envia um documento para validação
// @Tags documentos
// @Accept multipart/form-data
// @Produce json
// @Param type formData string true "Tipo de documento" Enums(rg, cpf, comprovante_endereco, outro)
// @Param file formData file true "Arquivo do documento"
// @Success 201 {object} domain.Document
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/documents [post]
// @Security BearerAuth
func (h *DocumentHandler) UploadDocument(c *gin.Context) {
	authID, err := middleware.GetAuthID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Não autorizado",
		})
		return
	}

	docType := domain.DocumentType(c.PostForm("type"))
	if docType == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Tipo de documento é obrigatório",
		})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Arquivo é obrigatório",
		})
		return
	}

	document, err := h.documentService.UploadDocument(authID, docType, file)
	if err != nil {
		if err == service.ErrInvalidFileType {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "Tipo de arquivo inválido",
			})
			return
		}
		if err == service.ErrFileTooLarge {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "Arquivo muito grande (máx. 10MB)",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Falha ao enviar documento",
		})
		return
	}

	c.JSON(http.StatusCreated, document)
}

// GetUserDocuments godoc
// @Summary Obter documentos do usuário
// @Description Retorna todos os documentos enviados pelo usuário autenticado
// @Tags documentos
// @Produce json
// @Success 200 {array} domain.Document
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/documents [get]
// @Security BearerAuth
func (h *DocumentHandler) GetUserDocuments(c *gin.Context) {
	authID, err := middleware.GetAuthID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Não autorizado",
		})
		return
	}

	documents, err := h.documentService.GetUserDocuments(authID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Falha ao obter documentos",
		})
		return
	}

	c.JSON(http.StatusOK, documents)
}

// GetDocumentByID godoc
// @Summary Obter documento por ID
// @Description Retorna um documento específico pelo ID
// @Tags documentos
// @Produce json
// @Param id path string true "ID do documento"
// @Success 200 {object} domain.Document
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/documents/{id} [get]
// @Security BearerAuth
func (h *DocumentHandler) GetDocumentByID(c *gin.Context) {
	authID, err := middleware.GetAuthID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Não autorizado",
		})
		return
	}

	documentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "ID do documento inválido",
		})
		return
	}

	document, err := h.documentService.GetDocumentByID(authID, documentID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error: "Documento não encontrado",
		})
		return
	}

	c.JSON(http.StatusOK, document)
}

// GetPendingDocuments godoc
// @Summary Obter documentos pendentes
// @Description Retorna todos os documentos pendentes de revisão (somente Admin/Professor)
// @Tags documentos
// @Produce json
// @Success 200 {array} domain.Document
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /api/v1/documents/pending [get]
// @Security BearerAuth
func (h *DocumentHandler) GetPendingDocuments(c *gin.Context) {
	documents, err := h.documentService.GetPendingDocuments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Falha ao obter documentos pendentes",
		})
		return
	}

	c.JSON(http.StatusOK, documents)
}

// ApproveDocument godoc
// @Summary Aprovar documento
// @Description Aprova um documento (somente Admin)
// @Tags documentos
// @Produce json
// @Param id path string true "ID do documento"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /api/v1/documents/{id}/approve [patch]
// @Security BearerAuth
func (h *DocumentHandler) ApproveDocument(c *gin.Context) {
	adminAuthID, err := middleware.GetAuthID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Não autorizado",
		})
		return
	}

	documentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "ID do documento inválido",
		})
		return
	}

	if err := h.documentService.ApproveDocument(adminAuthID, documentID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Falha ao aprovar documento",
		})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{
		Message: "Documento aprovado com sucesso",
	})
}

// RejectDocument godoc
// @Summary Rejeitar documento
// @Description Rejeita um documento com motivo (somente Admin)
// @Tags documentos
// @Produce json
// @Param id path string true "ID do documento"
// @Param request body RejectDocumentRequest true "Motivo da rejeição"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /api/v1/documents/{id}/reject [patch]
// @Security BearerAuth
func (h *DocumentHandler) RejectDocument(c *gin.Context) {
	adminAuthID, err := middleware.GetAuthID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Não autorizado",
		})
		return
	}

	documentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "ID do documento inválido",
		})
		return
	}

	var req RejectDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Formato de requisição inválido",
		})
		return
	}

	if err := h.documentService.RejectDocument(adminAuthID, documentID, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Falha ao rejeitar documento",
		})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{
		Message: "Documento rejeitado com sucesso",
	})
}

// DeleteDocument godoc
// @Summary Excluir documento
// @Description Exclui um documento
// @Tags documentos
// @Produce json
// @Param id path string true "ID do documento"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/documents/{id} [delete]
// @Security BearerAuth
func (h *DocumentHandler) DeleteDocument(c *gin.Context) {
	authID, err := middleware.GetAuthID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Não autorizado",
		})
		return
	}

	documentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "ID do documento inválido",
		})
		return
	}

	if err := h.documentService.DeleteDocument(authID, documentID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Falha ao excluir documento",
		})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{
		Message: "Documento excluído com sucesso",
	})
}
