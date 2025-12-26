package models

import (
	"testing"
	"time"
)

func TestSpamScore_Structure(t *testing.T) {
	score := SpamScore{
		RuleName: "test_rule",
		Score:    0.5,
		Reason:   "Test reason",
	}

	if score.RuleName != "test_rule" {
		t.Errorf("Expected rule name 'test_rule', got '%s'", score.RuleName)
	}

	if score.Score != 0.5 {
		t.Errorf("Expected score 0.5, got %f", score.Score)
	}

	if score.Reason != "Test reason" {
		t.Errorf("Expected reason 'Test reason', got '%s'", score.Reason)
	}
}

func TestSpamDetectionResult_Structure(t *testing.T) {
	result := SpamDetectionResult{
		PhoneNumber:     "7379037972",
		UserPhoneNumber: "9876543210",
		IsSpam:          false,
		AverageScore:    0.3,
		RuleScores: []SpamScore{
			{
				RuleName: "rule1",
				Score:    0.2,
				Reason:   "Reason 1",
			},
			{
				RuleName: "rule2",
				Score:    0.4,
				Reason:   "Reason 2",
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	if result.PhoneNumber != "7379037972" {
		t.Errorf("Expected phone number '7379037972', got '%s'", result.PhoneNumber)
	}

	if result.UserPhoneNumber != "9876543210" {
		t.Errorf("Expected user phone number '9876543210', got '%s'", result.UserPhoneNumber)
	}

	if result.IsSpam != false {
		t.Errorf("Expected IsSpam to be false, got %v", result.IsSpam)
	}

	if result.AverageScore != 0.3 {
		t.Errorf("Expected average score 0.3, got %f", result.AverageScore)
	}

	if len(result.RuleScores) != 2 {
		t.Errorf("Expected 2 rule scores, got %d", len(result.RuleScores))
	}
}

func TestSpamDetectionRequest_Structure(t *testing.T) {
	req := SpamDetectionRequest{
		PhoneNumber:     "7379037972",
		UserPhoneNumber: "9876543210",
	}

	if req.PhoneNumber != "7379037972" {
		t.Errorf("Expected phone number '7379037972', got '%s'", req.PhoneNumber)
	}

	if req.UserPhoneNumber != "9876543210" {
		t.Errorf("Expected user phone number '9876543210', got '%s'", req.UserPhoneNumber)
	}
}

func TestSpamDetectionRequest_OptionalUserPhone(t *testing.T) {
	req := SpamDetectionRequest{
		PhoneNumber: "7379037972",
		// UserPhoneNumber is optional
	}

	if req.PhoneNumber != "7379037972" {
		t.Errorf("Expected phone number '7379037972', got '%s'", req.PhoneNumber)
	}

	if req.UserPhoneNumber != "" {
		t.Errorf("Expected empty user phone number, got '%s'", req.UserPhoneNumber)
	}
}

