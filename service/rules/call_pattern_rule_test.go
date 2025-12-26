package rules

import (
	"context"
	"testing"
	"time"

	"credCode/models"
	"credCode/repository"
)

// mockGraphRepositoryForCalls is a mock implementation for call pattern rule tests
type mockGraphRepositoryForCalls struct {
	callsWithFiltersFunc func(ctx context.Context, phoneNumber string, filters repository.CallFilters, direction string) ([]*models.Edge, int)
}

func (m *mockGraphRepositoryForCalls) GetCallsWithFilters(ctx context.Context, phoneNumber string, filters repository.CallFilters, direction string) ([]*models.Edge, int) {
	if m.callsWithFiltersFunc != nil {
		return m.callsWithFiltersFunc(ctx, phoneNumber, filters, direction)
	}
	return nil, 0
}

// Implement other required methods with no-ops
func (m *mockGraphRepositoryForCalls) AddNode(ctx context.Context, phoneNumber string) error {
	return nil
}
func (m *mockGraphRepositoryForCalls) AddNodeWithName(ctx context.Context, phoneNumber, name string) error {
	return nil
}
func (m *mockGraphRepositoryForCalls) GetNode(ctx context.Context, phoneNumber string) (*models.Node, error) {
	return nil, nil
}
func (m *mockGraphRepositoryForCalls) NodeExists(ctx context.Context, phoneNumber string) bool {
	return false
}
func (m *mockGraphRepositoryForCalls) GetAllNodes(ctx context.Context) ([]*models.Node, error) {
	return nil, nil
}
func (m *mockGraphRepositoryForCalls) DeleteNode(ctx context.Context, phoneNumber string) error {
	return nil
}
func (m *mockGraphRepositoryForCalls) AddEdgeWithMetadata(ctx context.Context, from, to string, metadata models.EdgeMetadata) (*models.Edge, error) {
	return nil, nil
}
func (m *mockGraphRepositoryForCalls) GetEdge(ctx context.Context, edgeID string) (*models.Edge, error) {
	return nil, nil
}
func (m *mockGraphRepositoryForCalls) GetEdgeWithMetadata(ctx context.Context, edgeID string) (*models.Edge, models.EdgeMetadata, error) {
	return nil, nil, nil
}
func (m *mockGraphRepositoryForCalls) DeleteEdge(ctx context.Context, edgeID string) error {
	return nil
}
func (m *mockGraphRepositoryForCalls) GetUsersWithContact(ctx context.Context, phoneNumber string) ([]string, int) {
	return nil, 0
}
func (m *mockGraphRepositoryForCalls) GetOutgoingEdges(ctx context.Context, phoneNumber string, edgeType models.EdgeType) []*models.Edge {
	return nil
}
func (m *mockGraphRepositoryForCalls) GetIncomingEdges(ctx context.Context, phoneNumber string, edgeType models.EdgeType) []*models.Edge {
	return nil
}
func (m *mockGraphRepositoryForCalls) IsDirectContact(ctx context.Context, userPhone, callerPhone string) bool {
	return false
}
func (m *mockGraphRepositoryForCalls) GetSecondLevelContactCount(ctx context.Context, userPhone, callerPhone string) int {
	return 0
}
func (m *mockGraphRepositoryForCalls) LoadSeedData(ctx context.Context, filePath string) error {
	return nil
}

// TestCallPatternRule_NoSuspiciousCalls tests when there are no suspicious calls
func TestCallPatternRule_NoSuspiciousCalls(t *testing.T) {
	rule := NewCallPatternRule(30, 60*time.Minute, 0.6).(*CallPatternRule)

	mockRepo := &mockGraphRepositoryForCalls{
		callsWithFiltersFunc: func(ctx context.Context, phoneNumber string, filters repository.CallFilters, direction string) ([]*models.Edge, int) {
			// No suspicious calls
			return []*models.Edge{}, 0
		},
	}

	score, err := rule.Evaluate(context.Background(), "1234567890", "", mockRepo)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if score.Score != 0.0 {
		t.Errorf("Expected score 0.0 for no suspicious calls, got: %f", score.Score)
	}

	if score.RuleName != "call_pattern_rule" {
		t.Errorf("Expected rule name 'call_pattern_rule', got: %s", score.RuleName)
	}
}

// TestCallPatternRule_OneSuspiciousCall tests when there is one suspicious call
func TestCallPatternRule_OneSuspiciousCall(t *testing.T) {
	rule := NewCallPatternRule(30, 60*time.Minute, 0.6).(*CallPatternRule)

	mockRepo := &mockGraphRepositoryForCalls{
		callsWithFiltersFunc: func(ctx context.Context, phoneNumber string, filters repository.CallFilters, direction string) ([]*models.Edge, int) {
			if filters.IsAnswered != nil && *filters.IsAnswered && filters.MaxDuration != nil {
				// Suspicious calls query
				return []*models.Edge{{ID: "call1"}}, 1
			}
			// Total calls query
			return []*models.Edge{{ID: "call1"}, {ID: "call2"}}, 2
		},
	}

	score, err := rule.Evaluate(context.Background(), "1234567890", "", mockRepo)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// With 1 suspicious call, score = 0.6 * (1 - 1/(1+1)) = 0.6 * 0.5 = 0.3
	expectedScore := 0.6 * (1.0 - 1.0/(1.0+1.0))
	if score.Score != expectedScore {
		t.Errorf("Expected score %f for 1 suspicious call, got: %f", expectedScore, score.Score)
	}

	if score.Score < 0 || score.Score > 0.6 {
		t.Errorf("Score %f should be in range [0, 0.6]", score.Score)
	}
}

// TestCallPatternRule_MultipleSuspiciousCalls tests when there are multiple suspicious calls
func TestCallPatternRule_MultipleSuspiciousCalls(t *testing.T) {
	rule := NewCallPatternRule(30, 60*time.Minute, 0.6).(*CallPatternRule)

	mockRepo := &mockGraphRepositoryForCalls{
		callsWithFiltersFunc: func(ctx context.Context, phoneNumber string, filters repository.CallFilters, direction string) ([]*models.Edge, int) {
			if filters.IsAnswered != nil && *filters.IsAnswered && filters.MaxDuration != nil {
				// Suspicious calls query - 5 suspicious calls
				return make([]*models.Edge, 5), 5
			}
			// Total calls query
			return make([]*models.Edge, 10), 10
		},
	}

	score, err := rule.Evaluate(context.Background(), "1234567890", "", mockRepo)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// With 5 suspicious calls, score should be higher
	expectedScore := 0.6 * (1.0 - 1.0/(1.0+5.0))
	if score.Score != expectedScore {
		t.Errorf("Expected score %f for 5 suspicious calls, got: %f", expectedScore, score.Score)
	}

	if score.Score > 0.6 {
		t.Errorf("Score %f should not exceed suspiciousWeight 0.6", score.Score)
	}
}

// TestCallPatternRule_ManySuspiciousCalls tests when there are many suspicious calls
func TestCallPatternRule_ManySuspiciousCalls(t *testing.T) {
	rule := NewCallPatternRule(30, 60*time.Minute, 0.6).(*CallPatternRule)

	mockRepo := &mockGraphRepositoryForCalls{
		callsWithFiltersFunc: func(ctx context.Context, phoneNumber string, filters repository.CallFilters, direction string) ([]*models.Edge, int) {
			if filters.IsAnswered != nil && *filters.IsAnswered && filters.MaxDuration != nil {
				// Many suspicious calls
				return make([]*models.Edge, 100), 100
			}
			// Total calls
			return make([]*models.Edge, 150), 150
		},
	}

	score, err := rule.Evaluate(context.Background(), "1234567890", "", mockRepo)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// With many suspicious calls, score should approach suspiciousWeight but not exceed it
	if score.Score > 0.6 {
		t.Errorf("Score %f should not exceed suspiciousWeight 0.6", score.Score)
	}

	if score.Score < 0 {
		t.Error("Score should not be negative")
	}
}

// TestCallPatternRule_FiltersApplied tests that filters are correctly applied
func TestCallPatternRule_FiltersApplied(t *testing.T) {
	rule := NewCallPatternRule(30, 60*time.Minute, 0.6).(*CallPatternRule)

	var capturedFilters repository.CallFilters
	mockRepo := &mockGraphRepositoryForCalls{
		callsWithFiltersFunc: func(ctx context.Context, phoneNumber string, filters repository.CallFilters, direction string) ([]*models.Edge, int) {
			if filters.IsAnswered != nil && *filters.IsAnswered && filters.MaxDuration != nil {
				capturedFilters = filters
				return []*models.Edge{}, 0
			}
			return []*models.Edge{}, 0
		},
	}

	_, err := rule.Evaluate(context.Background(), "1234567890", "", mockRepo)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify filters were set correctly
	if capturedFilters.IsAnswered == nil || !*capturedFilters.IsAnswered {
		t.Error("Expected IsAnswered filter to be true")
	}

	if capturedFilters.MaxDuration == nil || *capturedFilters.MaxDuration != 30 {
		t.Errorf("Expected MaxDuration filter to be 30, got: %v", capturedFilters.MaxDuration)
	}

	if capturedFilters.TimeRangeStart == nil {
		t.Error("Expected TimeRangeStart filter to be set")
	}
}

// TestCallPatternRule_Name tests the rule name
func TestCallPatternRule_Name(t *testing.T) {
	rule := NewCallPatternRule(30, 60*time.Minute, 0.6).(*CallPatternRule)

	if rule.Name() != "call_pattern_rule" {
		t.Errorf("Expected rule name 'call_pattern_rule', got: %s", rule.Name())
	}
}

// TestCallPatternRule_ScoreRange tests that scores are in valid range
func TestCallPatternRule_ScoreRange(t *testing.T) {
	rule := NewCallPatternRule(30, 60*time.Minute, 0.6).(*CallPatternRule)

	testCases := []struct {
		name            string
		suspiciousCount int
		totalCount      int
	}{
		{"no_suspicious", 0, 0},
		{"one_suspicious", 1, 5},
		{"few_suspicious", 3, 10},
		{"many_suspicious", 10, 20},
		{"very_many_suspicious", 50, 100},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := &mockGraphRepositoryForCalls{
				callsWithFiltersFunc: func(ctx context.Context, phoneNumber string, filters repository.CallFilters, direction string) ([]*models.Edge, int) {
					if filters.IsAnswered != nil && *filters.IsAnswered && filters.MaxDuration != nil {
						return make([]*models.Edge, tc.suspiciousCount), tc.suspiciousCount
					}
					return make([]*models.Edge, tc.totalCount), tc.totalCount
				},
			}

			score, err := rule.Evaluate(context.Background(), "1234567890", "", mockRepo)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if score.Score < 0 || score.Score > 1.0 {
				t.Errorf("Score %f is out of valid range [0, 1] for suspicious=%d, total=%d", score.Score, tc.suspiciousCount, tc.totalCount)
			}

			if score.Score > 0.6 {
				t.Errorf("Score %f should not exceed suspiciousWeight 0.6", score.Score)
			}
		})
	}
}
