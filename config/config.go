package config

// Config holds all application configuration
type Config struct {
	// Data paths
	UserSeedDataPath string
	CallDataPath     string

	// Server configuration
	ServerPort string

	// Spam detection configuration
	SpamThreshold float64

	// Rule configurations
	ContactCountThreshold        int
	ContactCountMaxScore         float64
	CallPatternDurationThreshold int
	CallPatternTimeWindow        string // Duration string like "60m"
	CallPatternSuspiciousWeight  float64
	SecondLevelThreshold         int
	SecondLevelMaxScore          float64
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	return &Config{
		UserSeedDataPath:             "contacts_generated.json",
		CallDataPath:                 "call_data.json",
		ServerPort:                   "8080",
		SpamThreshold:                0.5,
		ContactCountThreshold:        3,
		ContactCountMaxScore:         0.7,
		CallPatternDurationThreshold: 30,
		CallPatternTimeWindow:        "60m",
		CallPatternSuspiciousWeight:  0.6,
		SecondLevelThreshold:         2,
		SecondLevelMaxScore:          0.5,
	}
}
