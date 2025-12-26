package service

import (
	"context"

	"credCode/models"
	"credCode/repository"
)

// SpamRule defines the interface for spam detection rules
// Each rule analyzes the graph data and returns a spam score
type SpamRule interface {
	// Name returns the name of the rule
	Name() string

	// Evaluate evaluates the rule and returns a spam score
	// Score: 0.0 = not spam, 1.0 = definitely spam
	// ctx: Context for cancellation and timeouts
	// phoneNumber: The caller's phone number
	// userPhoneNumber: The user's phone number (optional, empty string if not provided)
	// graphRepo: The graph repository to query
	Evaluate(ctx context.Context, phoneNumber string, userPhoneNumber string, graphRepo repository.GraphRepository) (*models.SpamScore, error)
}

// SpamRuleRegistry manages all registered spam rules
type SpamRuleRegistry struct {
	rules []SpamRule
}

// NewSpamRuleRegistry creates a new rule registry
func NewSpamRuleRegistry() *SpamRuleRegistry {
	return &SpamRuleRegistry{
		rules: make([]SpamRule, 0),
	}
}

// Register adds a rule to the registry
func (r *SpamRuleRegistry) Register(rule SpamRule) {
	r.rules = append(r.rules, rule)
}

// GetAllRules returns all registered rules
func (r *SpamRuleRegistry) GetAllRules() []SpamRule {
	return r.rules
}
