package rules

import (
	"context"
	"testing"

	"credCode/models"
	"credCode/repository"
)

// mockGraphRepositoryForSecondLevel is a mock implementation for second-level contact rule tests
type mockGraphRepositoryForSecondLevel struct {
	isDirectContactFunc          func(ctx context.Context, userPhone, callerPhone string) bool
	secondLevelContactCountFunc  func(ctx context.Context, userPhone, callerPhone string) int
}

func (m *mockGraphRepositoryForSecondLevel) IsDirectContact(ctx context.Context, userPhone, callerPhone string) bool {
	if m.isDirectContactFunc != nil {
		return m.isDirectContactFunc(ctx, userPhone, callerPhone)
	}
	return false
}

func (m *mockGraphRepositoryForSecondLevel) GetSecondLevelContactCount(ctx context.Context, userPhone, callerPhone string) int {
	if m.secondLevelContactCountFunc != nil {
		return m.secondLevelContactCountFunc(ctx, userPhone, callerPhone)
	}
	return 0
}

// Implement other required methods with no-ops
func (m *mockGraphRepositoryForSecondLevel) AddNode(ctx context.Context, phoneNumber string) error { return nil }
func (m *mockGraphRepositoryForSecondLevel) AddNodeWithName(ctx context.Context, phoneNumber, name string) error { return nil }
func (m *mockGraphRepositoryForSecondLevel) GetNode(ctx context.Context, phoneNumber string) (*models.Node, error) { return nil, nil }
func (m *mockGraphRepositoryForSecondLevel) NodeExists(ctx context.Context, phoneNumber string) bool { return false }
func (m *mockGraphRepositoryForSecondLevel) GetAllNodes(ctx context.Context) ([]*models.Node, error) { return nil, nil }
func (m *mockGraphRepositoryForSecondLevel) DeleteNode(ctx context.Context, phoneNumber string) error { return nil }
func (m *mockGraphRepositoryForSecondLevel) AddEdgeWithMetadata(ctx context.Context, from, to string, metadata models.EdgeMetadata) (*models.Edge, error) { return nil, nil }
func (m *mockGraphRepositoryForSecondLevel) GetEdge(ctx context.Context, edgeID string) (*models.Edge, error) { return nil, nil }
func (m *mockGraphRepositoryForSecondLevel) GetEdgeWithMetadata(ctx context.Context, edgeID string) (*models.Edge, models.EdgeMetadata, error) { return nil, nil, nil }
func (m *mockGraphRepositoryForSecondLevel) DeleteEdge(ctx context.Context, edgeID string) error { return nil }
func (m *mockGraphRepositoryForSecondLevel) GetUsersWithContact(ctx context.Context, phoneNumber string) ([]string, int) { return nil, 0 }
func (m *mockGraphRepositoryForSecondLevel) GetOutgoingEdges(ctx context.Context, phoneNumber string, edgeType models.EdgeType) []*models.Edge { return nil }
func (m *mockGraphRepositoryForSecondLevel) GetIncomingEdges(ctx context.Context, phoneNumber string, edgeType models.EdgeType) []*models.Edge { return nil }
func (m *mockGraphRepositoryForSecondLevel) GetCallsWithFilters(ctx context.Context, phoneNumber string, filters repository.CallFilters, direction string) ([]*models.Edge, int) { return nil, 0 }
func (m *mockGraphRepositoryForSecondLevel) LoadSeedData(ctx context.Context, filePath string) error { return nil }

// TestSecondLevelContactRule_NoUserPhone tests when user phone number is not provided
func TestSecondLevelContactRule_NoUserPhone(t *testing.T) {
	rule := NewSecondLevelContactRule(2, 0.5).(*SecondLevelContactRule)

	mockRepo := &mockGraphRepositoryForSecondLevel{}

	score, err := rule.Evaluate(context.Background(), "1234567890", "", mockRepo)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if score.Score != 0.0 {
		t.Errorf("Expected score 0.0 when user phone not provided, got: %f", score.Score)
	}

	if score.RuleName != "second_level_contact_rule" {
		t.Errorf("Expected rule name 'second_level_contact_rule', got: %s", score.RuleName)
	}
}

// TestSecondLevelContactRule_DirectContact tests when caller is in user's direct contacts
func TestSecondLevelContactRule_DirectContact(t *testing.T) {
	rule := NewSecondLevelContactRule(2, 0.5).(*SecondLevelContactRule)

	mockRepo := &mockGraphRepositoryForSecondLevel{
		isDirectContactFunc: func(ctx context.Context, userPhone, callerPhone string) bool {
			return true // Caller is in user's direct contacts
		},
	}

	score, err := rule.Evaluate(context.Background(), "1234567890", "9876543210", mockRepo)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if score.Score != 0.0 {
		t.Errorf("Expected score 0.0 for direct contact, got: %f", score.Score)
	}

	if score.Reason != "Caller is in user's direct contact list (level 1)" {
		t.Errorf("Expected reason about direct contact, got: %s", score.Reason)
	}
}

// TestSecondLevelContactRule_ZeroLevel2Contacts tests when there are zero level-2 contacts
func TestSecondLevelContactRule_ZeroLevel2Contacts(t *testing.T) {
	rule := NewSecondLevelContactRule(2, 0.5).(*SecondLevelContactRule)

	mockRepo := &mockGraphRepositoryForSecondLevel{
		isDirectContactFunc: func(ctx context.Context, userPhone, callerPhone string) bool {
			return false // Not a direct contact
		},
		secondLevelContactCountFunc: func(ctx context.Context, userPhone, callerPhone string) int {
			return 0 // No level-2 contacts
		},
	}

	score, err := rule.Evaluate(context.Background(), "1234567890", "9876543210", mockRepo)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if score.Score != 0.5 {
		t.Errorf("Expected score 0.5 (maxScore) for zero level-2 contacts, got: %f", score.Score)
	}
}

// TestSecondLevelContactRule_BelowThreshold tests when level-2 count is below threshold
func TestSecondLevelContactRule_BelowThreshold(t *testing.T) {
	rule := NewSecondLevelContactRule(2, 0.5).(*SecondLevelContactRule)

	mockRepo := &mockGraphRepositoryForSecondLevel{
		isDirectContactFunc: func(ctx context.Context, userPhone, callerPhone string) bool {
			return false
		},
		secondLevelContactCountFunc: func(ctx context.Context, userPhone, callerPhone string) int {
			return 1 // Below threshold of 2
		},
	}

	score, err := rule.Evaluate(context.Background(), "1234567890", "9876543210", mockRepo)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// With threshold=2, count=1, expected score = 0.5 * (1 - 1/2) = 0.5 * 0.5 = 0.25
	expectedScore := 0.5 * (1.0 - 1.0/2.0)
	if score.Score != expectedScore {
		t.Errorf("Expected score %f for 1 level-2 contact (below threshold 2), got: %f", expectedScore, score.Score)
	}
}

// TestSecondLevelContactRule_AtThreshold tests when level-2 count equals threshold
func TestSecondLevelContactRule_AtThreshold(t *testing.T) {
	rule := NewSecondLevelContactRule(2, 0.5).(*SecondLevelContactRule)

	mockRepo := &mockGraphRepositoryForSecondLevel{
		isDirectContactFunc: func(ctx context.Context, userPhone, callerPhone string) bool {
			return false
		},
		secondLevelContactCountFunc: func(ctx context.Context, userPhone, callerPhone string) int {
			return 2 // At threshold
		},
	}

	score, err := rule.Evaluate(context.Background(), "1234567890", "9876543210", mockRepo)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// At threshold, score should be low
	expectedScore := 0.1 * (1.0 - float64(2-2)/float64(2+2))
	if score.Score != expectedScore {
		t.Errorf("Expected score %f for 2 level-2 contacts (at threshold), got: %f", expectedScore, score.Score)
	}

	if score.Score < 0 {
		t.Error("Score should not be negative")
	}
}

// TestSecondLevelContactRule_AboveThreshold tests when level-2 count is above threshold
func TestSecondLevelContactRule_AboveThreshold(t *testing.T) {
	rule := NewSecondLevelContactRule(2, 0.5).(*SecondLevelContactRule)

	mockRepo := &mockGraphRepositoryForSecondLevel{
		isDirectContactFunc: func(ctx context.Context, userPhone, callerPhone string) bool {
			return false
		},
		secondLevelContactCountFunc: func(ctx context.Context, userPhone, callerPhone string) int {
			return 5 // Above threshold of 2
		},
	}

	score, err := rule.Evaluate(context.Background(), "1234567890", "9876543210", mockRepo)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Above threshold, score should be very low
	expectedScore := 0.1 * (1.0 - float64(5-2)/float64(5+2))
	if score.Score != expectedScore {
		t.Errorf("Expected score %f for 5 level-2 contacts (above threshold 2), got: %f", expectedScore, score.Score)
	}

	if score.Score < 0 {
		t.Error("Score should not be negative")
	}

	if score.Score >= 0.5 {
		t.Error("Score should be low for trusted numbers via level-2")
	}
}

// TestSecondLevelContactRule_ManyLevel2Contacts tests when there are many level-2 contacts
func TestSecondLevelContactRule_ManyLevel2Contacts(t *testing.T) {
	rule := NewSecondLevelContactRule(2, 0.5).(*SecondLevelContactRule)

	mockRepo := &mockGraphRepositoryForSecondLevel{
		isDirectContactFunc: func(ctx context.Context, userPhone, callerPhone string) bool {
			return false
		},
		secondLevelContactCountFunc: func(ctx context.Context, userPhone, callerPhone string) int {
			return 50 // Many level-2 contacts
		},
	}

	score, err := rule.Evaluate(context.Background(), "1234567890", "9876543210", mockRepo)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// With many level-2 contacts, score should approach 0
	if score.Score < 0 {
		t.Error("Score should not be negative")
	}

	if score.Score > 0.1 {
		t.Errorf("Score should be very low for many level-2 contacts, got: %f", score.Score)
	}
}

// TestSecondLevelContactRule_Name tests the rule name
func TestSecondLevelContactRule_Name(t *testing.T) {
	rule := NewSecondLevelContactRule(2, 0.5).(*SecondLevelContactRule)

	if rule.Name() != "second_level_contact_rule" {
		t.Errorf("Expected rule name 'second_level_contact_rule', got: %s", rule.Name())
	}
}

// TestSecondLevelContactRule_ScoreRange tests that scores are in valid range [0, 1]
func TestSecondLevelContactRule_ScoreRange(t *testing.T) {
	rule := NewSecondLevelContactRule(2, 0.5).(*SecondLevelContactRule)

	testCases := []struct {
		name            string
		isDirect        bool
		level2Count     int
		userPhoneNumber string
	}{
		{"no_user_phone", false, 0, ""},
		{"direct_contact", true, 0, "9876543210"},
		{"zero_level2", false, 0, "9876543210"},
		{"one_level2", false, 1, "9876543210"},
		{"two_level2", false, 2, "9876543210"},
		{"five_level2", false, 5, "9876543210"},
		{"many_level2", false, 50, "9876543210"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := &mockGraphRepositoryForSecondLevel{
				isDirectContactFunc: func(ctx context.Context, userPhone, callerPhone string) bool {
					return tc.isDirect
				},
				secondLevelContactCountFunc: func(ctx context.Context, userPhone, callerPhone string) int {
					return tc.level2Count
				},
			}

			score, err := rule.Evaluate(context.Background(), "1234567890", tc.userPhoneNumber, mockRepo)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if score.Score < 0 || score.Score > 1.0 {
				t.Errorf("Score %f is out of valid range [0, 1] for isDirect=%v, level2Count=%d", score.Score, tc.isDirect, tc.level2Count)
			}
		})
	}
}

// TestSecondLevelContactRule_DirectContactSkipsLevel2 tests that direct contact check happens first
func TestSecondLevelContactRule_DirectContactSkipsLevel2(t *testing.T) {
	rule := NewSecondLevelContactRule(2, 0.5).(*SecondLevelContactRule)

	level2Called := false
	mockRepo := &mockGraphRepositoryForSecondLevel{
		isDirectContactFunc: func(ctx context.Context, userPhone, callerPhone string) bool {
			return true // Direct contact, should skip level-2 check
		},
		secondLevelContactCountFunc: func(ctx context.Context, userPhone, callerPhone string) int {
			level2Called = true // This should not be called
			return 0
		},
	}

	score, err := rule.Evaluate(context.Background(), "1234567890", "9876543210", mockRepo)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if level2Called {
		t.Error("GetSecondLevelContactCount should not be called when IsDirectContact returns true")
	}

	if score.Score != 0.0 {
		t.Errorf("Expected score 0.0 for direct contact, got: %f", score.Score)
	}
}

