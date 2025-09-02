// filepath: /Users/potato/Desktop/Dev/Go_server/locker-server/internal/api/handlers/common.go

package handlers

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error string `json:"error" example:"invalid credentials"`
}
