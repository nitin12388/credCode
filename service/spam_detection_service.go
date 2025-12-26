package service

import (
	"context"
	"fmt"
	"time"

	"credCode/models"
	"credCode/repository"
	"credCode/service/scoring"
)

// SpamDetectionService orchestrates spam detection using multiple rules
type SpamDetectionService struct {
	graphRepo repository.GraphRepository
	registry  *SpamRuleRegistry
	scorer    scoring.Scorer
	threshold float64 // Average score threshold to consider as spam (e.g., 0.5)
}

// NewSpamDetectionService creates a new spam detection service
func NewSpamDetectionService(graphRepo repository.GraphRepository, threshold float64) *SpamDetectionService {
	service := &SpamDetectionService{
		graphRepo: graphRepo,
		registry:  NewSpamRuleRegistry(),
		scorer:    scoring.NewAverageScorer(),
		threshold: threshold,
	}

	// Register default rules
	service.registerDefaultRules()

	return service
}

// registerDefaultRules registers the default spam detection rules
// This is a no-op - rules should be registered externally to avoid import cycles
func (s *SpamDetectionService) registerDefaultRules() {
	// Rules are registered via RegisterRule() from external packages
	// This avoids import cycles between service and rules packages
}

// RegisterRule allows adding custom rules
func (s *SpamDetectionService) RegisterRule(rule SpamRule) {
	s.registry.Register(rule)
}

// DetectSpam runs all registered rules and returns the spam detection result
func (s *SpamDetectionService) DetectSpam(phoneNumber string, userPhoneNumber string) (*models.SpamDetectionResult, error) {
	rules := s.registry.GetAllRules()
	if len(rules) == 0 {
		return nil, fmt.Errorf("no spam detection rules registered")
	}

	ruleScores := make([]models.SpamScore, 0, len(rules))

	// Evaluate each rule
	ctx := context.Background()
	for _, rule := range rules {
		score, err := rule.Evaluate(ctx, phoneNumber, userPhoneNumber, s.graphRepo)
		if err != nil {
			// Log error but continue with other rules
			fmt.Printf("Error evaluating rule %s: %v\n", rule.Name(), err)
			continue
		}

		ruleScores = append(ruleScores, *score)
	}

	// Calculate score using injected scorer
	averageScore, isSpam := s.scorer.CalculateScore(ruleScores, s.threshold)

	result := &models.SpamDetectionResult{
		PhoneNumber:     phoneNumber,
		UserPhoneNumber: userPhoneNumber,
		IsSpam:          isSpam,
		AverageScore:    averageScore,
		RuleScores:      ruleScores,
		Timestamp:       time.Now().Format(time.RFC3339),
	}

	return result, nil
}

// GetRegisteredRules returns the names of all registered rules
func (s *SpamDetectionService) GetRegisteredRules() []string {
	rules := s.registry.GetAllRules()
	names := make([]string, len(rules))
	for i, rule := range rules {
		names[i] = rule.Name()
	}
	return names
}
