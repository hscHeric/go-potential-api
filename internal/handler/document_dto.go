package handler

// RejectDocumentRequest representa o payload para rejeitar documento
type RejectDocumentRequest struct {
	Reason string `json:"reason" binding:"required,min=10"`
}
