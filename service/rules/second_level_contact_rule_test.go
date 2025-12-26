package rules

import (
	"context"
	"testing"
	"time"

	"credCode/models"
	"credCode/repository"
)

func TestNewSecondLevelContactRule(t *testing.T) {
	rule := NewSecondLevelContactRule(2, 0.5)

	if rule == nil {
		t.Fatal("Expected rule to be created, got nil")
	}

	// Test that rule has correct name
	if rule.Name() != "second_level_contact_rule" {
		t.Errorf("Expected name 'second_level_contact_rule', got '%s'", rule.Name())
	}
}

func TestSecondLevelContactRule_Name(t *testing.T) {
	rule := NewSecondLevelContactRule(2, 0.5)

	if rule.Name() != "second_level_contact_rule" {
		t.Errorf("Expected name 'second_level_contact_rule', got '%s'", rule.Name())
	}
}

func TestSecondLevelContactRule_Evaluate_NoUserPhone(t *testing.T) {
	graphRepo := repository.NewInMemoryGraphRepository()
	rule := NewSecondLevelContactRule(2, 0.5)

	ctx := context.Background()
	graphRepo.AddNode(ctx, "7379037972")

	// Test with empty user phone number
	score, err := rule.Evaluate(ctx, "7379037972", "", graphRepo)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should skip and return 0.0
	if score.Score != 0.0 {
		t.Errorf("Expected score 0.0 when user phone not provided, got %f", score.Score)
	}
}

func TestSecondLevelContactRule_Evaluate_DirectContact(t *testing.T) {
	graphRepo := repository.NewInMemoryGraphRepository()
	rule := NewSecondLevelContactRule(2, 0.5)

	ctx := context.Background()
	graphRepo.AddNode(ctx, "7379037972")
	graphRepo.AddNode(ctx, "9876543210")

	// Create direct contact
	meta := &models.ContactMetadata{Name: "Contact", AddedAt: time.Now()}
	graphRepo.AddEdgeWithMetadata(ctx, "9876543210", "7379037972", meta)

	// Test with direct contact
	score, err := rule.Evaluate(ctx, "7379037972", "9876543210", graphRepo)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should return 0.0 for direct contact
	if score.Score != 0.0 {
		t.Errorf("Expected score 0.0 for direct contact, got %f", score.Score)
	}
}

func TestSecondLevelContactRule_Evaluate_SecondLevelContact(t *testing.T) {
	graphRepo := repository.NewInMemoryGraphRepository()
	rule := NewSecondLevelContactRule(2, 0.5) // Min 2 level-2 matches

	ctx := context.Background()
	phones := []string{"7379037972", "9876543210", "1234567890", "5555555555"}
	for _, phone := range phones {
		graphRepo.AddNode(ctx, phone)
	}

	// Level 1: 7379037972 -> 9876543210, 1234567890
	meta1 := &models.ContactMetadata{Name: "Contact1", AddedAt: time.Now()}
	graphRepo.AddEdgeWithMetadata(ctx, "7379037972", "9876543210", meta1)
	graphRepo.AddEdgeWithMetadata(ctx, "7379037972", "1234567890", meta1)

	// Level 2: Both 9876543210 and 1234567890 have 5555555555 as contact
	meta2 := &models.ContactMetadata{Name: "Contact2", AddedAt: time.Now()}
	graphRepo.AddEdgeWithMetadata(ctx, "9876543210", "5555555555", meta2)
	graphRepo.AddEdgeWithMetadata(ctx, "1234567890", "5555555555", meta2)

	// Test: caller is 5555555555, user is 7379037972
	score, err := rule.Evaluate(ctx, "5555555555", "7379037972", graphRepo)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should return low score (<= 0.1) as we have 2 level-2 matches (>= threshold)
	// The formula returns a small positive value, not exactly 0.0
	if score.Score > 0.1 {
		t.Errorf("Expected low score (<= 0.1) for sufficient level-2 matches, got %f", score.Score)
	}
}

func TestSecondLevelContactRule_Evaluate_InsufficientLevel2Matches(t *testing.T) {
	graphRepo := repository.NewInMemoryGraphRepository()
	rule := NewSecondLevelContactRule(2, 0.5) // Min 2 level-2 matches

	ctx := context.Background()
	phones := []string{"7379037972", "9876543210", "5555555555"}
	for _, phone := range phones {
		graphRepo.AddNode(ctx, phone)
	}

	// Level 1: 7379037972 -> 9876543210
	meta1 := &models.ContactMetadata{Name: "Contact1", AddedAt: time.Now()}
	graphRepo.AddEdgeWithMetadata(ctx, "7379037972", "9876543210", meta1)

	// Level 2: Only 9876543210 has 5555555555 as contact (only 1 match)
	meta2 := &models.ContactMetadata{Name: "Contact2", AddedAt: time.Now()}
	graphRepo.AddEdgeWithMetadata(ctx, "9876543210", "5555555555", meta2)

	// Test: caller is 5555555555, user is 7379037972
	score, err := rule.Evaluate(ctx, "5555555555", "7379037972", graphRepo)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should return moderate score as we have only 1 level-2 match (< threshold)
	// Formula: maxScore * (1.0 - count/threshold) = 0.5 * (1.0 - 1/2) = 0.25
	if score.Score < 0.2 || score.Score > 0.3 {
		t.Errorf("Expected score around 0.25 for insufficient level-2 matches, got %f", score.Score)
	}
}

func TestSecondLevelContactRule_Evaluate_NoLevel2Matches(t *testing.T) {
	graphRepo := repository.NewInMemoryGraphRepository()
	rule := NewSecondLevelContactRule(2, 0.5)

	ctx := context.Background()
	phones := []string{"7379037972", "9876543210", "5555555555"}
	for _, phone := range phones {
		graphRepo.AddNode(ctx, phone)
	}

	// Level 1: 7379037972 -> 9876543210
	meta1 := &models.ContactMetadata{Name: "Contact1", AddedAt: time.Now()}
	graphRepo.AddEdgeWithMetadata(ctx, "7379037972", "9876543210", meta1)

	// No level 2 contacts for 5555555555

	// Test: caller is 5555555555, user is 7379037972
	score, err := rule.Evaluate(ctx, "5555555555", "7379037972", graphRepo)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should return maxScore as no level-2 matches
	if score.Score != 0.5 {
		t.Errorf("Expected score 0.5 for no level-2 matches, got %f", score.Score)
	}
}
