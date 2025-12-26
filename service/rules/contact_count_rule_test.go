package rules

import (
	"context"
	"testing"

	"credCode/models"
	"credCode/repository"
)

// mockGraphRepository is a mock implementation of GraphRepository for testing
type mockGraphRepository struct {
	usersWithContactFunc func(ctx context.Context, phoneNumber string) ([]string, int)
}

func (m *mockGraphRepository) GetUsersWithContact(ctx context.Context, phoneNumber string) ([]string, int) {
	if m.usersWithContactFunc != nil {
		return m.usersWithContactFunc(ctx, phoneNumber)
	}
	return nil, 0
}

// Implement other required methods with no-ops for this test
func (m *mockGraphRepository) AddNode(ctx context.Context, phoneNumber string) error { return nil }
func (m *mockGraphRepository) AddNodeWithName(ctx context.Context, phoneNumber, name string) error {
	return nil
}
func (m *mockGraphRepository) GetNode(ctx context.Context, phoneNumber string) (*models.Node, error) {
	return nil, nil
}
func (m *mockGraphRepository) NodeExists(ctx context.Context, phoneNumber string) bool { return false }
func (m *mockGraphRepository) GetAllNodes(ctx context.Context) ([]*models.Node, error) {
	return nil, nil
}
func (m *mockGraphRepository) DeleteNode(ctx context.Context, phoneNumber string) error { return nil }
func (m *mockGraphRepository) AddEdgeWithMetadata(ctx context.Context, from, to string, metadata models.EdgeMetadata) (*models.Edge, error) {
	return nil, nil
}
func (m *mockGraphRepository) GetEdge(ctx context.Context, edgeID string) (*models.Edge, error) {
	return nil, nil
}
func (m *mockGraphRepository) GetEdgeWithMetadata(ctx context.Context, edgeID string) (*models.Edge, models.EdgeMetadata, error) {
	return nil, nil, nil
}
func (m *mockGraphRepository) DeleteEdge(ctx context.Context, edgeID string) error { return nil }
func (m *mockGraphRepository) GetOutgoingEdges(ctx context.Context, phoneNumber string, edgeType models.EdgeType) []*models.Edge {
	return nil
}
func (m *mockGraphRepository) GetIncomingEdges(ctx context.Context, phoneNumber string, edgeType models.EdgeType) []*models.Edge {
	return nil
}
func (m *mockGraphRepository) GetCallsWithFilters(ctx context.Context, phoneNumber string, filters repository.CallFilters, direction string) ([]*models.Edge, int) {
	return nil, 0
}
func (m *mockGraphRepository) IsDirectContact(ctx context.Context, userPhone, callerPhone string) bool {
	return false
}
func (m *mockGraphRepository) GetSecondLevelContactCount(ctx context.Context, userPhone, callerPhone string) int {
	return 0
}
func (m *mockGraphRepository) LoadSeedData(ctx context.Context, filePath string) error { return nil }

// TestContactCountRule_ZeroContacts tests when phone number has zero contacts
func TestContactCountRule_ZeroContacts(t *testing.T) {
	rule := NewContactCountRule(3, 0.7).(*ContactCountRule)

	mockRepo := &mockGraphRepository{
		usersWithContactFunc: func(ctx context.Context, phoneNumber string) ([]string, int) {
			return []string{}, 0
		},
	}

	score, err := rule.Evaluate(context.Background(), "1234567890", "", mockRepo)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if score.Score != 0.7 {
		t.Errorf("Expected score 0.7 for zero contacts, got: %f", score.Score)
	}

	if score.RuleName != "contact_count_rule" {
		t.Errorf("Expected rule name 'contact_count_rule', got: %s", score.RuleName)
	}
}

// TestContactCountRule_BelowThreshold tests when phone number has contacts below threshold
func TestContactCountRule_BelowThreshold(t *testing.T) {
	rule := NewContactCountRule(3, 0.7).(*ContactCountRule)

	mockRepo := &mockGraphRepository{
		usersWithContactFunc: func(ctx context.Context, phoneNumber string) ([]string, int) {
			return []string{"user1", "user2"}, 2
		},
	}

	score, err := rule.Evaluate(context.Background(), "1234567890", "", mockRepo)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// With threshold=3, count=2, expected score = 0.7 * (1 - 2/3) = 0.7 * 0.333 = 0.233
	expectedScore := 0.7 * (1.0 - 2.0/3.0)
	if score.Score != expectedScore {
		t.Errorf("Expected score %f for 2 contacts (below threshold 3), got: %f", expectedScore, score.Score)
	}
}

// TestContactCountRule_AtThreshold tests when phone number has exactly threshold contacts
func TestContactCountRule_AtThreshold(t *testing.T) {
	rule := NewContactCountRule(3, 0.7).(*ContactCountRule)

	mockRepo := &mockGraphRepository{
		usersWithContactFunc: func(ctx context.Context, phoneNumber string) ([]string, int) {
			return []string{"user1", "user2", "user3"}, 3
		},
	}

	score, err := rule.Evaluate(context.Background(), "1234567890", "", mockRepo)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// At threshold, score should be low (0.1 * (1 - 0/6) = 0.1)
	expectedScore := 0.1 * (1.0 - float64(3-3)/float64(3+3))
	if score.Score != expectedScore {
		t.Errorf("Expected score %f for 3 contacts (at threshold), got: %f", expectedScore, score.Score)
	}

	if score.Score < 0 {
		t.Error("Score should not be negative")
	}
}

// TestContactCountRule_AboveThreshold tests when phone number has more contacts than threshold
func TestContactCountRule_AboveThreshold(t *testing.T) {
	rule := NewContactCountRule(3, 0.7).(*ContactCountRule)

	mockRepo := &mockGraphRepository{
		usersWithContactFunc: func(ctx context.Context, phoneNumber string) ([]string, int) {
			return []string{"user1", "user2", "user3", "user4", "user5"}, 5
		},
	}

	score, err := rule.Evaluate(context.Background(), "1234567890", "", mockRepo)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Above threshold, score should be very low
	expectedScore := 0.1 * (1.0 - float64(5-3)/float64(5+3))
	if score.Score != expectedScore {
		t.Errorf("Expected score %f for 5 contacts (above threshold 3), got: %f", expectedScore, score.Score)
	}

	if score.Score < 0 {
		t.Error("Score should not be negative")
	}

	if score.Score >= 0.7 {
		t.Error("Score should be low for trusted numbers")
	}
}

// TestContactCountRule_ManyContacts tests when phone number has many contacts
func TestContactCountRule_ManyContacts(t *testing.T) {
	rule := NewContactCountRule(3, 0.7).(*ContactCountRule)

	mockRepo := &mockGraphRepository{
		usersWithContactFunc: func(ctx context.Context, phoneNumber string) ([]string, int) {
			return make([]string, 100), 100
		},
	}

	score, err := rule.Evaluate(context.Background(), "1234567890", "", mockRepo)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// With many contacts, score should approach 0
	if score.Score < 0 {
		t.Error("Score should not be negative")
	}

	if score.Score > 0.1 {
		t.Errorf("Score should be very low for many contacts, got: %f", score.Score)
	}
}

// TestContactCountRule_Name tests the rule name
func TestContactCountRule_Name(t *testing.T) {
	rule := NewContactCountRule(3, 0.7).(*ContactCountRule)

	if rule.Name() != "contact_count_rule" {
		t.Errorf("Expected rule name 'contact_count_rule', got: %s", rule.Name())
	}
}

// TestContactCountRule_ScoreRange tests that scores are in valid range [0, 1]
func TestContactCountRule_ScoreRange(t *testing.T) {
	rule := NewContactCountRule(3, 0.7).(*ContactCountRule)

	testCases := []struct {
		name  string
		count int
	}{
		{"zero", 0},
		{"one", 1},
		{"two", 2},
		{"three", 3},
		{"five", 5},
		{"ten", 10},
		{"hundred", 100},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := &mockGraphRepository{
				usersWithContactFunc: func(ctx context.Context, phoneNumber string) ([]string, int) {
					return make([]string, tc.count), tc.count
				},
			}

			score, err := rule.Evaluate(context.Background(), "1234567890", "", mockRepo)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if score.Score < 0 || score.Score > 1.0 {
				t.Errorf("Score %f is out of valid range [0, 1] for count %d", score.Score, tc.count)
			}
		})
	}
}
