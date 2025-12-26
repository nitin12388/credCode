package api

import (
	"encoding/json"
	"net/http"

	"credCode/models"
	"credCode/service"
)

// SpamDetectionHandler handles spam detection API requests
type SpamDetectionHandler struct {
	spamService *service.SpamDetectionService
	validator   RequestValidator
}

// NewSpamDetectionHandler creates a new spam detection handler
func NewSpamDetectionHandler(spamService *service.SpamDetectionService) *SpamDetectionHandler {
	return &SpamDetectionHandler{
		spamService: spamService,
		validator:   NewRequestValidator(),
	}
}

// DetectSpam handles POST /api/v1/spam/detect
func (h *SpamDetectionHandler) DetectSpam(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		WriteMethodNotAllowed(w)
		return
	}

	// Parse request body
	var req models.SpamDetectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "Invalid request body: "+err.Error())
		return
	}

	// Validate request
	if err := h.validator.ValidateSpamRequest(&req); err != nil {
		WriteBadRequest(w, err.Error())
		return
	}

	// Detect spam (pass user phone number if provided)
	result, err := h.spamService.DetectSpam(req.PhoneNumber, req.UserPhoneNumber)
	if err != nil {
		WriteInternalServerError(w, "Error detecting spam: "+err.Error())
		return
	}

	// Return result
	WriteSuccess(w, result)
}

// GetSpamScore handles GET /api/v1/spam/score?phone_number=...
func (h *SpamDetectionHandler) GetSpamScore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteMethodNotAllowed(w)
		return
	}

	// Get phone numbers from query parameters
	phoneNumber := r.URL.Query().Get("phone_number")
	if phoneNumber == "" {
		WriteBadRequest(w, "phone_number query parameter is required")
		return
	}

	// User phone number is optional
	userPhoneNumber := r.URL.Query().Get("user_phone_number")

	// Create request object for validation
	req := models.SpamDetectionRequest{
		PhoneNumber:     phoneNumber,
		UserPhoneNumber: userPhoneNumber,
	}

	// Validate request
	if err := h.validator.ValidateSpamRequest(&req); err != nil {
		WriteBadRequest(w, err.Error())
		return
	}

	// Detect spam
	result, err := h.spamService.DetectSpam(phoneNumber, userPhoneNumber)
	if err != nil {
		WriteInternalServerError(w, "Error detecting spam: "+err.Error())
		return
	}

	// Return result
	WriteSuccess(w, result)
}

// GetRules handles GET /api/v1/spam/rules
func (h *SpamDetectionHandler) GetRules(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteMethodNotAllowed(w)
		return
	}

	rules := h.spamService.GetRegisteredRules()

	WriteSuccess(w, map[string]interface{}{
		"rules": rules,
		"count": len(rules),
	})
}
