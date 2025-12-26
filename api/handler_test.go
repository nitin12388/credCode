package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"credCode/models"
	"credCode/repository"
	"credCode/service"
	"credCode/service/rules"
)

func TestNewSpamDetectionHandler(t *testing.T) {
	graphRepo := repository.NewInMemoryGraphRepository()
	spamService := service.NewSpamDetectionService(graphRepo, 0.5)

	handler := NewSpamDetectionHandler(spamService)

	if handler == nil {
		t.Fatal("Expected handler to be created, got nil")
	}

	if handler.spamService == nil {
		t.Error("Expected spamService to be set")
	}

	if handler.validator == nil {
		t.Error("Expected validator to be set")
	}
}

func TestSpamDetectionHandler_DetectSpam_POST(t *testing.T) {
	graphRepo := repository.NewInMemoryGraphRepository()
	spamService := service.NewSpamDetectionService(graphRepo, 0.5)
	spamService.RegisterRule(rules.NewContactCountRule(3, 0.7))

	handler := NewSpamDetectionHandler(spamService)

	// Create request body
	reqBody := models.SpamDetectionRequest{
		PhoneNumber:     "7379037972",
		UserPhoneNumber: "",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/spam/detect", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.DetectSpam(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result models.SpamDetectionResult
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if result.PhoneNumber != "7379037972" {
		t.Errorf("Expected phone '7379037972', got '%s'", result.PhoneNumber)
	}
}

func TestSpamDetectionHandler_DetectSpam_WithUserPhone(t *testing.T) {
	graphRepo := repository.NewInMemoryGraphRepository()
	spamService := service.NewSpamDetectionService(graphRepo, 0.5)
	spamService.RegisterRule(rules.NewContactCountRule(3, 0.7))
	spamService.RegisterRule(rules.NewSecondLevelContactRule(2, 0.5))

	handler := NewSpamDetectionHandler(spamService)

	reqBody := models.SpamDetectionRequest{
		PhoneNumber:     "7379037972",
		UserPhoneNumber: "9876543210",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/spam/detect", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.DetectSpam(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestSpamDetectionHandler_DetectSpam_WrongMethod(t *testing.T) {
	graphRepo := repository.NewInMemoryGraphRepository()
	spamService := service.NewSpamDetectionService(graphRepo, 0.5)
	handler := NewSpamDetectionHandler(spamService)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/spam/detect", nil)
	w := httptest.NewRecorder()

	handler.DetectSpam(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestSpamDetectionHandler_DetectSpam_InvalidBody(t *testing.T) {
	graphRepo := repository.NewInMemoryGraphRepository()
	spamService := service.NewSpamDetectionService(graphRepo, 0.5)
	handler := NewSpamDetectionHandler(spamService)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/spam/detect", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.DetectSpam(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestSpamDetectionHandler_DetectSpam_MissingPhoneNumber(t *testing.T) {
	graphRepo := repository.NewInMemoryGraphRepository()
	spamService := service.NewSpamDetectionService(graphRepo, 0.5)
	handler := NewSpamDetectionHandler(spamService)

	reqBody := models.SpamDetectionRequest{
		PhoneNumber: "", // Missing
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/spam/detect", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.DetectSpam(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestSpamDetectionHandler_GetSpamScore_GET(t *testing.T) {
	graphRepo := repository.NewInMemoryGraphRepository()
	spamService := service.NewSpamDetectionService(graphRepo, 0.5)
	spamService.RegisterRule(rules.NewContactCountRule(3, 0.7))

	handler := NewSpamDetectionHandler(spamService)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/spam/score?phone_number=7379037972", nil)
	w := httptest.NewRecorder()

	handler.GetSpamScore(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result models.SpamDetectionResult
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if result.PhoneNumber != "7379037972" {
		t.Errorf("Expected phone '7379037972', got '%s'", result.PhoneNumber)
	}
}

func TestSpamDetectionHandler_GetSpamScore_WithUserPhone(t *testing.T) {
	graphRepo := repository.NewInMemoryGraphRepository()
	spamService := service.NewSpamDetectionService(graphRepo, 0.5)
	spamService.RegisterRule(rules.NewContactCountRule(3, 0.7))
	spamService.RegisterRule(rules.NewSecondLevelContactRule(2, 0.5))

	handler := NewSpamDetectionHandler(spamService)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/spam/score?phone_number=7379037972&user_phone_number=9876543210", nil)
	w := httptest.NewRecorder()

	handler.GetSpamScore(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestSpamDetectionHandler_GetSpamScore_MissingPhoneNumber(t *testing.T) {
	graphRepo := repository.NewInMemoryGraphRepository()
	spamService := service.NewSpamDetectionService(graphRepo, 0.5)
	handler := NewSpamDetectionHandler(spamService)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/spam/score", nil) // No phone_number param
	w := httptest.NewRecorder()

	handler.GetSpamScore(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestSpamDetectionHandler_GetSpamScore_WrongMethod(t *testing.T) {
	graphRepo := repository.NewInMemoryGraphRepository()
	spamService := service.NewSpamDetectionService(graphRepo, 0.5)
	handler := NewSpamDetectionHandler(spamService)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/spam/score?phone_number=7379037972", nil)
	w := httptest.NewRecorder()

	handler.GetSpamScore(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestSpamDetectionHandler_GetRules(t *testing.T) {
	graphRepo := repository.NewInMemoryGraphRepository()
	spamService := service.NewSpamDetectionService(graphRepo, 0.5)
	spamService.RegisterRule(rules.NewContactCountRule(3, 0.7))
	spamService.RegisterRule(rules.NewCallPatternRule(30, 60*time.Minute, 0.6))

	handler := NewSpamDetectionHandler(spamService)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/spam/rules", nil)
	w := httptest.NewRecorder()

	handler.GetRules(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	rules, ok := response["rules"].([]interface{})
	if !ok {
		t.Error("Expected 'rules' field in response")
	}

	if len(rules) != 2 {
		t.Errorf("Expected 2 rules, got %d", len(rules))
	}
}

func TestSpamDetectionHandler_GetRules_WrongMethod(t *testing.T) {
	graphRepo := repository.NewInMemoryGraphRepository()
	spamService := service.NewSpamDetectionService(graphRepo, 0.5)
	handler := NewSpamDetectionHandler(spamService)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/spam/rules", nil)
	w := httptest.NewRecorder()

	handler.GetRules(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

