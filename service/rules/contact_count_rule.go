package rules

import (
	"context"
	"fmt"

	"credCode/models"
	"credCode/repository"
	"credCode/service"
)

// ContactCountRule evaluates spam score based on how many users have saved this number
// More users saving = lower spam score (more trusted)
type ContactCountRule struct {
	threshold int     // Minimum contacts to be considered trusted
	maxScore  float64 // Maximum spam score if below threshold
}

// NewContactCountRule creates a new contact count rule
func NewContactCountRule(threshold int, maxScore float64) service.SpamRule {
	return &ContactCountRule{
		threshold: threshold,
		maxScore:  maxScore,
	}
}

// Name returns the rule name
func (r *ContactCountRule) Name() string {
	return "contact_count_rule"
}

// Evaluate evaluates the contact count rule
func (r *ContactCountRule) Evaluate(ctx context.Context, phoneNumber string, userPhoneNumber string, graphRepo repository.GraphRepository) (*models.SpamScore, error) {
	// Query: How many users have saved this phone number?
	_, count := graphRepo.GetUsersWithContact(ctx, phoneNumber)

	// Calculate score: fewer contacts = higher spam score
	var score float64
	var reason string

	if count == 0 {
		// No one has saved this number - high spam probability
		score = r.maxScore
		reason = fmt.Sprintf("Phone number not saved by any user (0 contacts)")
	} else if count < r.threshold {
		// Below threshold - moderate spam probability
		// Score decreases as count increases
		score = r.maxScore * (1.0 - float64(count)/float64(r.threshold))
		reason = fmt.Sprintf("Phone number saved by only %d user(s) (below threshold of %d)", count, r.threshold)
	} else {
		// Above threshold - low spam probability
		// Score decreases further as count increases
		score = 0.1 * (1.0 - float64(count-r.threshold)/float64(count+r.threshold))
		if score < 0 {
			score = 0
		}
		reason = fmt.Sprintf("Phone number saved by %d users (trusted)", count)
	}

	return &models.SpamScore{
		RuleName: r.Name(),
		Score:    score,
		Reason:   reason,
	}, nil
}

// Ensure ContactCountRule implements SpamRule interface
var _ service.SpamRule = (*ContactCountRule)(nil)
