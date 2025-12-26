package scoring

import "credCode/models"

// AverageScorer calculates spam score using average of all rule scores
type AverageScorer struct{}

// NewAverageScorer creates a new average-based scorer
func NewAverageScorer() Scorer {
	return &AverageScorer{}
}

// CalculateScore calculates the average score from all rule scores
// Returns: (averageScore, isSpam)
func (s *AverageScorer) CalculateScore(scores []models.SpamScore, threshold float64) (float64, bool) {
	if len(scores) == 0 {
		return 0.0, false
	}

	var totalScore float64
	for _, score := range scores {
		totalScore += score.Score
	}

	averageScore := totalScore / float64(len(scores))
	isSpam := averageScore >= threshold

	return averageScore, isSpam
}
