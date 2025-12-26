package api

import (
	"fmt"
	"log"
	"net/http"

	"credCode/service"
)

// Server represents the HTTP server
type Server struct {
	handler *SpamDetectionHandler
	port    string
}

// NewServer creates a new HTTP server
func NewServer(spamService *service.SpamDetectionService, port string) *Server {
	return &Server{
		handler: NewSpamDetectionHandler(spamService),
		port:    port,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Register routes
	http.HandleFunc("/api/v1/spam/detect", s.handler.DetectSpam)
	http.HandleFunc("/api/v1/spam/score", s.handler.GetSpamScore)
	http.HandleFunc("/api/v1/spam/rules", s.handler.GetRules)
	http.HandleFunc("/health", s.healthCheck)

	addr := fmt.Sprintf(":%s", s.port)
	log.Printf("Starting server on port %s", s.port)
	log.Printf("API endpoints:")
	log.Printf("  POST /api/v1/spam/detect - Detect spam (JSON: phone_number, user_phone_number)")
	log.Printf("  GET  /api/v1/spam/score  - Get spam score (query: phone_number, user_phone_number)")
	log.Printf("  GET  /api/v1/spam/rules  - Get registered rules")
	log.Printf("  GET  /health             - Health check")

	return http.ListenAndServe(addr, nil)
}

// healthCheck handles health check requests
func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy"}`)
}
