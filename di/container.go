package di

import (
	"context"
	"log"
	"time"

	"credCode/api"
	"credCode/config"
	"credCode/repository"
	"credCode/service"
	"credCode/service/rules"
)

// Container holds all application dependencies
type Container struct {
	config       *config.Config
	userRepo     repository.UserRepository
	graphRepo    repository.GraphRepository
	graphBuilder service.GraphBuilder
	spamService  *service.SpamDetectionService
	server       *api.Server
}

// NewContainer creates a new dependency injection container
func NewContainer(cfg *config.Config) (*Container, error) {
	container := &Container{
		config: cfg,
	}

	// Initialize repositories
	container.userRepo = repository.NewInMemoryUserRepository()
	container.graphRepo = repository.NewInMemoryGraphRepository()

	// Initialize graph builder
	container.graphBuilder = service.NewGraphBuilder()

	// Load seed data
	if err := container.loadSeedData(); err != nil {
		return nil, err
	}

	// Build graph from user data
	if err := container.buildGraph(); err != nil {
		return nil, err
	}

	// Initialize spam detection service
	if err := container.initializeSpamService(); err != nil {
		return nil, err
	}

	// Initialize server
	container.server = api.NewServer(container.spamService, cfg.ServerPort)

	return container, nil
}

// loadSeedData loads user and call data into repositories
func (c *Container) loadSeedData() error {
	ctx := context.Background()

	log.Println("Loading user seed data...")
	if err := c.userRepo.LoadSeedData(ctx, c.config.UserSeedDataPath); err != nil {
		return err
	}
	log.Println("✓ User seed data loaded successfully")

	log.Println("Loading call data...")
	if err := c.graphRepo.LoadSeedData(ctx, c.config.CallDataPath); err != nil {
		return err
	}
	log.Println("✓ Call data loaded successfully")

	return nil
}

// buildGraph builds the graph from user data
func (c *Container) buildGraph() error {
	log.Println("Building graph from user data...")
	if err := c.graphBuilder.BuildFromUsers(c.userRepo, c.graphRepo); err != nil {
		return err
	}
	log.Println("✓ Graph constructed successfully")
	return nil
}

// initializeSpamService creates and configures the spam detection service
func (c *Container) initializeSpamService() error {
	// Create spam detection service with threshold from config
	spamService := service.NewSpamDetectionService(c.graphRepo, c.config.SpamThreshold)

	// Register default rules with config values
	contactRule := rules.NewContactCountRule(
		c.config.ContactCountThreshold,
		c.config.ContactCountMaxScore,
	)
	spamService.RegisterRule(contactRule)

	// Parse time window duration
	timeWindow, err := time.ParseDuration(c.config.CallPatternTimeWindow)
	if err != nil {
		// Default to 60 minutes if parsing fails
		timeWindow = 60 * time.Minute
	}

	callPatternRule := rules.NewCallPatternRule(
		c.config.CallPatternDurationThreshold,
		timeWindow,
		c.config.CallPatternSuspiciousWeight,
	)
	spamService.RegisterRule(callPatternRule)

	secondLevelRule := rules.NewSecondLevelContactRule(
		c.config.SecondLevelThreshold,
		c.config.SecondLevelMaxScore,
	)
	spamService.RegisterRule(secondLevelRule)

	c.spamService = spamService
	return nil
}

// GetServer returns the HTTP server
func (c *Container) GetServer() *api.Server {
	return c.server
}

// GetUserRepo returns the user repository (for testing)
func (c *Container) GetUserRepo() repository.UserRepository {
	return c.userRepo
}

// GetGraphRepo returns the graph repository (for testing)
func (c *Container) GetGraphRepo() repository.GraphRepository {
	return c.graphRepo
}

// GetSpamService returns the spam detection service (for testing)
func (c *Container) GetSpamService() *service.SpamDetectionService {
	return c.spamService
}
