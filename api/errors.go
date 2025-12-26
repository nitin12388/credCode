package api

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// WriteError writes an error response to the HTTP response writer
func WriteError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	})
}

// WriteBadRequest writes a 400 Bad Request error
func WriteBadRequest(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusBadRequest, message)
}

// WriteInternalServerError writes a 500 Internal Server Error
func WriteInternalServerError(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusInternalServerError, message)
}

// WriteMethodNotAllowed writes a 405 Method Not Allowed error
func WriteMethodNotAllowed(w http.ResponseWriter) {
	WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
}

