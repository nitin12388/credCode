package rules

import (
	"context"
	"testing"
	"time"

	"credCode/models"
	"credCode/repository"
)

func TestNewContactCountRule(t *testing.T) {
	rule := NewContactCountRule(3, 0.7)

	if rule == nil {
		t.Fatal("Expected rule to be created, got nil")
	}

	// Test that rule has correct name
	if rule.Name() != "contact_count_rule" {
		t.Errorf("Expected name 'contact_count_rule', got '%s'", rule.Name())
	}
}

func TestContactCountRule_Name(t *testing.T) {
	rule := NewContactCountRule(3, 0.7)

	if rule.Name() != "contact_count_rule" {
		t.Errorf("Expected name 'contact_count_rule', got '%s'", rule.Name())
	}
}

func TestContactCountRule_Evaluate_NoContacts(t *testing.T) {
	graphRepo := repository.NewInMemoryGraphRepository()
	rule := NewContactCountRule(3, 0.7)

	ctx := context.Background()
	graphRepo.AddNode(ctx, "9999999999") // Phone with no contacts

	score, err := rule.Evaluate(ctx, "9999999999", "", graphRepo)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should have high spam score (maxScore) for no contacts
	if score.Score != 0.7 {
		t.Errorf("Expected score 0.7 for no contacts, got %f", score.Score)
	}
}

func TestContactCountRule_Evaluate_BelowThreshold(t *testing.T) {
	graphRepo := repository.NewInMemoryGraphRepository()
	rule := NewContactCountRule(3, 0.7) // Threshold: 3

	ctx := context.Background()
	graphRepo.AddNode(ctx, "7379037972")
	graphRepo.AddNode(ctx, "9876543210")

	// Add 2 contacts (below threshold of 3)
	meta := &models.ContactMetadata{Name: "Contact", AddedAt: time.Now()}
	graphRepo.AddEdgeWithMetadata(ctx, "9876543210", "7379037972", meta)
	graphRepo.AddEdgeWithMetadata(ctx, "1234567890", "7379037972", meta)

	score, err := rule.Evaluate(ctx, "7379037972", "", graphRepo)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should have moderate spam score (between 0 and maxScore)
	if score.Score <= 0 || score.Score >= 0.7 {
		t.Errorf("Expected score between 0 and 0.7, got %f", score.Score)
	}
}

func TestContactCountRule_Evaluate_AboveThreshold(t *testing.T) {
	graphRepo := repository.NewInMemoryGraphRepository()
	rule := NewContactCountRule(3, 0.7) // Threshold: 3

	ctx := context.Background()
	phones := []string{"7379037972", "9876543210", "1234567890", "5555555555", "6666666666"}
	for _, phone := range phones {
		graphRepo.AddNode(ctx, phone)
	}

	// Add 4 contacts (above threshold of 3)
	meta := &models.ContactMetadata{Name: "Contact", AddedAt: time.Now()}
	graphRepo.AddEdgeWithMetadata(ctx, "9876543210", "7379037972", meta)
	graphRepo.AddEdgeWithMetadata(ctx, "1234567890", "7379037972", meta)
	graphRepo.AddEdgeWithMetadata(ctx, "5555555555", "7379037972", meta)
	graphRepo.AddEdgeWithMetadata(ctx, "6666666666", "7379037972", meta)

	score, err := rule.Evaluate(ctx, "7379037972", "", graphRepo)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should have low spam score
	if score.Score > 0.2 {
		t.Errorf("Expected low score (< 0.2), got %f", score.Score)
	}
}
