package rules

import (
	"context"
	"fmt"
	"time"

	"credCode/models"
	"credCode/repository"
	"credCode/service"
)

// CallPatternRule evaluates spam score based on call patterns
// Looks for suspicious patterns like: answered calls with very short duration
type CallPatternRule struct {
	durationThreshold int           // Duration threshold in seconds (e.g., 30)
	timeWindow        time.Duration // Time window to analyze (e.g., 60 minutes)
	suspiciousWeight  float64       // Weight for suspicious calls
}

// NewCallPatternRule creates a new call pattern rule
func NewCallPatternRule(durationThreshold int, timeWindow time.Duration, suspiciousWeight float64) service.SpamRule {
	return &CallPatternRule{
		durationThreshold: durationThreshold,
		timeWindow:        timeWindow,
		suspiciousWeight:  suspiciousWeight,
	}
}

// Name returns the rule name
func (r *CallPatternRule) Name() string {
	return "call_pattern_rule"
}

// Evaluate evaluates the call pattern rule
func (r *CallPatternRule) Evaluate(ctx context.Context, phoneNumber string, userPhoneNumber string, graphRepo repository.GraphRepository) (*models.SpamScore, error) {
	// Query: Get calls in the last time window
	timeStart := time.Now().Add(-r.timeWindow)

	// Filter: is_answered=true AND duration<=threshold
	answered := true
	maxDuration := r.durationThreshold

	filters := repository.CallFilters{
		IsAnswered:     &answered,
		MaxDuration:    &maxDuration,
		TimeRangeStart: &timeStart,
	}

	// Get both outgoing and incoming calls
	_, count := graphRepo.GetCallsWithFilters(ctx, phoneNumber, filters, "both")

	var score float64
	var reason string

	if count == 0 {
		// No suspicious calls - low spam score
		score = 0.0
		reason = fmt.Sprintf("No suspicious call patterns found in last %v", r.timeWindow)
	} else {
		// Calculate score based on number of suspicious calls
		// More suspicious calls = higher spam score
		// Score increases with count but caps at suspiciousWeight
		score = r.suspiciousWeight * (1.0 - 1.0/(1.0+float64(count)))
		if score > r.suspiciousWeight {
			score = r.suspiciousWeight
		}

		// Get total calls in time window for context
		_, totalCount := graphRepo.GetCallsWithFilters(ctx, phoneNumber, repository.CallFilters{
			TimeRangeStart: &timeStart,
		}, "both")

		reason = fmt.Sprintf("Found %d suspicious calls (answered but <=%ds) out of %d total calls in last %v",
			count, r.durationThreshold, totalCount, r.timeWindow)
	}

	return &models.SpamScore{
		RuleName: r.Name(),
		Score:    score,
		Reason:   reason,
	}, nil
}

// Ensure CallPatternRule implements SpamRule interface
var _ service.SpamRule = (*CallPatternRule)(nil)
