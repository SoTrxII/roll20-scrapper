package http_helpers

import "encoding/json"

// swagger:model ErrorTemplate
type ErrorTemplate struct {
	Message string `json:"message"`
}

func NewError(message string) *ErrorTemplate {
	return &ErrorTemplate{Message: message}
}

// FormatError Return a JSON stringified error
func FormatError(message string) []byte {
	et := NewError(message)
	formatted, _ := json.Marshal(et)
	return formatted
}
