package models

// SpamScore represents the result of a spam detection rule
type SpamScore struct {
	RuleName string  `json:"rule_name"`
	Score    float64 `json:"score"`  // Score between 0.0 (not spam) and 1.0 (spam)
	Reason   string  `json:"reason"` // Human-readable reason for the score
}

// SpamDetectionResult represents the final spam detection result
type SpamDetectionResult struct {
	PhoneNumber     string      `json:"phone_number"`
	UserPhoneNumber string      `json:"user_phone_number,omitempty"` // User's phone number if provided
	IsSpam          bool        `json:"is_spam"`
	AverageScore    float64     `json:"average_score"`
	RuleScores      []SpamScore `json:"rule_scores"`
	Timestamp       string      `json:"timestamp"`
}

// SpamDetectionRequest represents the API request
type SpamDetectionRequest struct {
	PhoneNumber     string `json:"phone_number"`                // Caller phone number
	UserPhoneNumber string `json:"user_phone_number,omitempty"` // User's phone number (optional, for context-aware rules)
}
