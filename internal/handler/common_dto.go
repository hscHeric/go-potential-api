package handler

type ErrorResponse struct {
	Error   string            `json:"error"`
	Details map[string]string `json:"details,omitempty"`
}

// MessageResponse representa uma resposta de sucesso simples
type MessageResponse struct {
	Message string `json:"message"`
}
