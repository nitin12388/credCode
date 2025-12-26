package api

import (
	"errors"

	"credCode/models"
)

// RequestValidator defines the interface for request validation
type RequestValidator interface {
	ValidateSpamRequest(req *models.SpamDetectionRequest) error
}

// spamRequestValidator implements RequestValidator
type spamRequestValidator struct{}

// NewRequestValidator creates a new request validator
func NewRequestValidator() RequestValidator {
	return &spamRequestValidator{}
}

// ValidateSpamRequest validates a spam detection request
func (v *spamRequestValidator) ValidateSpamRequest(req *models.SpamDetectionRequest) error {
	if req.PhoneNumber == "" {
		return errors.New("phone number is required")
	}
	return nil
}

