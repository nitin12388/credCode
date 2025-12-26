package rules

import (
	"context"
	"testing"
	"time"

	"credCode/models"
	"credCode/repository"
)

func TestNewCallPatternRule(t *testing.T) {
	rule := NewCallPatternRule(30, 60*time.Minute, 0.6)

	if rule == nil {
		t.Fatal("Expected rule to be created, got nil")
	}

	// Test that rule has correct name
	if rule.Name() != "call_pattern_rule" {
		t.Errorf("Expected name 'call_pattern_rule', got '%s'", rule.Name())
	}
}

func TestCallPatternRule_Name(t *testing.T) {
	rule := NewCallPatternRule(30, 60*time.Minute, 0.6)

	if rule.Name() != "call_pattern_rule" {
		t.Errorf("Expected name 'call_pattern_rule', got '%s'", rule.Name())
	}
}

func TestCallPatternRule_Evaluate_NoCalls(t *testing.T) {
	graphRepo := repository.NewInMemoryGraphRepository()
	rule := NewCallPatternRule(30, 60*time.Minute, 0.6)

	ctx := context.Background()
	graphRepo.AddNode(ctx, "9999999999") // Phone with no calls

	score, err := rule.Evaluate(ctx, "9999999999", "", graphRepo)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should have low spam score for no calls
	if score.Score > 0.1 {
		t.Errorf("Expected low score (< 0.1) for no calls, got %f", score.Score)
	}
}

func TestCallPatternRule_Evaluate_SuspiciousPattern(t *testing.T) {
	graphRepo := repository.NewInMemoryGraphRepository()
	rule := NewCallPatternRule(30, 60*time.Minute, 0.6) // Max duration 30s, window 60m

	ctx := context.Background()
	graphRepo.AddNode(ctx, "7379037972")
	graphRepo.AddNode(ctx, "9876543210")

	now := time.Now()

	// Add suspicious calls: answered but very short duration
	call1 := &models.CallMetadata{
		IsAnswered:        true,
		DurationInSeconds: 5, // Very short
		Timestamp:         now.Add(-30 * time.Minute),
	}
	call2 := &models.CallMetadata{
		IsAnswered:        true,
		DurationInSeconds: 10, // Very short
		Timestamp:         now.Add(-20 * time.Minute),
	}

	graphRepo.AddEdgeWithMetadata(ctx, "7379037972", "9876543210", call1)
	graphRepo.AddEdgeWithMetadata(ctx, "7379037972", "9876543210", call2)

	score, err := rule.Evaluate(ctx, "7379037972", "", graphRepo)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should have higher spam score for suspicious pattern
	if score.Score < 0.3 {
		t.Errorf("Expected higher score (>= 0.3) for suspicious pattern, got %f", score.Score)
	}
}

func TestCallPatternRule_Evaluate_NormalPattern(t *testing.T) {
	graphRepo := repository.NewInMemoryGraphRepository()
	rule := NewCallPatternRule(30, 60*time.Minute, 0.6)

	ctx := context.Background()
	graphRepo.AddNode(ctx, "7379037972")
	graphRepo.AddNode(ctx, "9876543210")

	now := time.Now()

	// Add normal calls: longer duration
	call1 := &models.CallMetadata{
		IsAnswered:        true,
		DurationInSeconds: 120, // Normal duration
		Timestamp:         now.Add(-30 * time.Minute),
	}

	graphRepo.AddEdgeWithMetadata(ctx, "7379037972", "9876543210", call1)

	score, err := rule.Evaluate(ctx, "7379037972", "", graphRepo)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should have lower spam score for normal pattern
	if score.Score > 0.2 {
		t.Errorf("Expected lower score (< 0.2) for normal pattern, got %f", score.Score)
	}
}

func TestCallPatternRule_Evaluate_OutsideTimeWindow(t *testing.T) {
	graphRepo := repository.NewInMemoryGraphRepository()
	rule := NewCallPatternRule(30, 60*time.Minute, 0.6) // Window: 60 minutes

	ctx := context.Background()
	graphRepo.AddNode(ctx, "7379037972")
	graphRepo.AddNode(ctx, "9876543210")

	now := time.Now()

	// Add call outside time window (2 hours ago)
	call1 := &models.CallMetadata{
		IsAnswered:        true,
		DurationInSeconds: 5, // Short duration
		Timestamp:         now.Add(-2 * time.Hour), // Outside window
	}

	graphRepo.AddEdgeWithMetadata(ctx, "7379037972", "9876543210", call1)

	score, err := rule.Evaluate(ctx, "7379037972", "", graphRepo)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should have low score as call is outside time window
	if score.Score > 0.1 {
		t.Errorf("Expected low score (< 0.1) for calls outside window, got %f", score.Score)
	}
}
