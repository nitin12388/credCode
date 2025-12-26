package rules

import (
	"context"
	"fmt"

	"credCode/models"
	"credCode/repository"
	"credCode/service"
)

// SecondLevelContactRule evaluates spam score based on second-level contacts
// Checks if caller is in user's contacts (level 1) or in contacts of user's contacts (level 2)
// More level-2 contacts having the caller = lower spam score (more trusted)
type SecondLevelContactRule struct {
	threshold int     // Minimum level-2 contacts to be considered trusted
	maxScore  float64 // Maximum spam score if below threshold
}

// NewSecondLevelContactRule creates a new second-level contact rule
func NewSecondLevelContactRule(threshold int, maxScore float64) service.SpamRule {
	return &SecondLevelContactRule{
		threshold: threshold,
		maxScore:  maxScore,
	}
}

// Name returns the rule name
func (r *SecondLevelContactRule) Name() string {
	return "second_level_contact_rule"
}

// Evaluate evaluates the second-level contact rule
func (r *SecondLevelContactRule) Evaluate(ctx context.Context, phoneNumber string, userPhoneNumber string, graphRepo repository.GraphRepository) (*models.SpamScore, error) {
	// If user phone number is not provided, skip this rule
	if userPhoneNumber == "" {
		return &models.SpamScore{
			RuleName: r.Name(),
			Score:    0.0,
			Reason:   "User phone number not provided, skipping second-level contact check",
		}, nil
	}

	// Step 1: Check if caller is directly in user's contacts (level 1) using Cayley query
	isDirectContact := graphRepo.IsDirectContact(ctx, userPhoneNumber, phoneNumber)

	if isDirectContact {
		// Caller is in user's direct contacts - very low spam score
		return &models.SpamScore{
			RuleName: r.Name(),
			Score:    0.0,
			Reason:   "Caller is in user's direct contact list (level 1)",
		}, nil
	}

	// Step 2: Check contacts of user's contacts (level 2) using Cayley graph query
	// This uses Cayley's path traversal: userPhone -has_contact-> ? -has_contact-> callerPhone
	level2Count := graphRepo.GetSecondLevelContactCount(ctx, userPhoneNumber, phoneNumber)

	// Calculate score based on level-2 count
	var score float64
	var reason string

	if level2Count == 0 {
		// No level-2 contacts have the caller - high spam probability
		score = r.maxScore
		reason = fmt.Sprintf("Caller not in user's contacts or contacts of user's contacts (0 level-2 matches)")
	} else if level2Count < r.threshold {
		// Below threshold - moderate spam probability
		// Score decreases as count increases
		score = r.maxScore * (1.0 - float64(level2Count)/float64(r.threshold))
		reason = fmt.Sprintf("Caller found in %d of user's contact's contact lists (below threshold of %d)", level2Count, r.threshold)
	} else {
		// Above threshold - low spam probability
		// Score decreases further as count increases
		score = 0.1 * (1.0 - float64(level2Count-r.threshold)/float64(level2Count+r.threshold))
		if score < 0 {
			score = 0
		}
		reason = fmt.Sprintf("Caller found in %d of user's contact's contact lists (trusted via level-2)", level2Count)
	}

	return &models.SpamScore{
		RuleName: r.Name(),
		Score:    score,
		Reason:   reason,
	}, nil
}

// Ensure SecondLevelContactRule implements SpamRule interface
var _ service.SpamRule = (*SecondLevelContactRule)(nil)
