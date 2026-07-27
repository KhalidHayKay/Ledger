package render

import (
	"encoding/json"
	"log"
	"net/http"
)

const contentTypeJSON = "application/json; charset=utf-8"

type APIError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

type Response struct {
	Success bool      `json:"success"`
	Message string    `json:"message,omitempty"`
	Data    any       `json:"data,omitempty"`
	Error   *APIError `json:"error,omitempty"`
}

// JSON writes a successful response.
func JSON(w http.ResponseWriter, status int, message string, data any) {
	writeJSON(w, status, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// ErrorJSON writes an error response.
func ErrorJSON(w http.ResponseWriter, message string, status int) {
	writeJSON(w, status, Response{
		Success: false,
		Error: &APIError{
			Message: message,
		},
	})
}

func writeJSON(w http.ResponseWriter, status int, payload Response) {
	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("ERROR: failed to marshal response: %v", err)

		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)

		return
	}

	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(status)

	if _, err := w.Write(body); err != nil {
		log.Printf("ERROR: failed to write response: %v", err)
	}
}
