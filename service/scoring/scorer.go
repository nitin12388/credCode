package scoring

import "credCode/models"

// Scorer defines the interface for calculating spam scores from rule scores
type Scorer interface {
	// CalculateScore calculates the final spam score and determines if it's spam
	// Returns: (averageScore, isSpam)
	CalculateScore(scores []models.SpamScore, threshold float64) (float64, bool)
}
