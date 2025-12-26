# Spam Detection Service

## Overview

A call interceptor service that detects spam calls by analyzing graph data using multiple rules. Each rule evaluates different patterns and returns a spam score. The service averages all rule scores to determine if a caller is spam.

## Architecture

```
API Layer (HTTP)
    ↓
Service Layer (SpamDetectionService)
    ↓
Rules Layer (SpamRule implementations)
    ↓
Repository Layer (GraphRepository)
    ↓
Cayley Graph Database
```

## Components

### 1. SpamRule Interface

Defines the contract for all spam detection rules:

```go
type SpamRule interface {
    Name() string
    Evaluate(phoneNumber string, graphRepo GraphRepository) (*SpamScore, error)
}
```

### 2. Spam Detection Service

Orchestrates multiple rules and calculates average spam score:

- Runs all registered rules
- Collects scores from each rule
- Calculates average score
- Determines if spam based on threshold (default: 0.5)

### 3. Rules

#### Contact Count Rule
- **Purpose**: Evaluates trust based on how many users have saved the number
- **Logic**: 
  - 0 contacts = high spam score (0.7)
  - < threshold (3) = moderate spam score
  - >= threshold = low spam score
- **Query**: Counts users who have saved this phone number

#### Call Pattern Rule
- **Purpose**: Detects suspicious call patterns
- **Logic**: 
  - Looks for answered calls with duration <= 30 seconds in last 60 minutes
  - More suspicious calls = higher spam score
- **Query**: Filters calls by `is_answered=true`, `duration<=30s`, `time_range=last_60_min`

### 4. API Endpoints

#### POST `/api/v1/spam/detect`
Detects spam for a phone number.

**Request:**
```json
{
  "phone_number": "7379037972"
}
```

**Response:**
```json
{
  "phone_number": "7379037972",
  "is_spam": false,
  "average_score": 0.15,
  "rule_scores": [
    {
      "rule_name": "contact_count_rule",
      "score": 0.1,
      "reason": "Phone number saved by 12 users (trusted)"
    },
    {
      "rule_name": "call_pattern_rule",
      "score": 0.2,
      "reason": "Found 1 suspicious calls (answered but <=30s) out of 2 total calls in last 1h0m0s"
    }
  ],
  "timestamp": "2025-12-26T15:00:00Z"
}
```

#### GET `/api/v1/spam/rules`
Returns all registered rules.

**Response:**
```json
{
  "rules": ["contact_count_rule", "call_pattern_rule"],
  "count": 2
}
```

#### GET `/health`
Health check endpoint.

**Response:**
```json
{
  "status": "healthy"
}
```

## Usage

### Starting the Server

```bash
go run cmd/server/main.go
```

Server starts on port 8080.

### Making API Calls

```bash
# Detect spam
curl -X POST http://localhost:8080/api/v1/spam/detect \
  -H "Content-Type: application/json" \
  -d '{"phone_number": "7379037972"}'

# Get rules
curl http://localhost:8080/api/v1/spam/rules
```

## Adding New Rules

### Step 1: Create Rule Implementation

```go
package rules

import (
    "credCode/models"
    "credCode/repository"
    "credCode/service"
)

type MyCustomRule struct {
    // Rule configuration
}

func (r *MyCustomRule) Name() string {
    return "my_custom_rule"
}

func (r *MyCustomRule) Evaluate(phoneNumber string, graphRepo repository.GraphRepository) (*models.SpamScore, error) {
    // Your rule logic here
    // Query graph repository
    // Calculate score (0.0 to 1.0)
    // Return SpamScore
}

func NewMyCustomRule() service.SpamRule {
    return &MyCustomRule{}
}
```

### Step 2: Register Rule

In `cmd/server/main.go`:

```go
customRule := rules.NewMyCustomRule()
spamService.RegisterRule(customRule)
```

## Rule Scoring Guidelines

- **0.0 - 0.3**: Not spam (trusted)
- **0.3 - 0.6**: Suspicious (moderate risk)
- **0.6 - 1.0**: Spam (high risk)

Each rule should return a score in this range based on its analysis.

## Configuration

### Service Threshold

Default threshold is 0.5. Average score >= threshold = spam.

```go
spamService := service.NewSpamDetectionService(graphRepo, 0.5)
```

### Rule Configuration

Rules can be configured with different parameters:

```go
// Contact Count Rule: threshold=3, maxScore=0.7
contactRule := rules.NewContactCountRule(3, 0.7)

// Call Pattern Rule: duration=30s, window=60min, weight=0.6
callRule := rules.NewCallPatternRule(30, 60*time.Minute, 0.6)
```

## Example Rule Ideas

1. **Frequency Rule**: Too many calls in short time
2. **Time Pattern Rule**: Calls at unusual hours
3. **Duration Pattern Rule**: All calls very short or very long
4. **Reciprocity Rule**: User never calls back
5. **Network Analysis Rule**: Connected to known spam numbers

## Testing

Run the test script:

```bash
./test_spam_api.sh
```

Or test manually:

```bash
# Start server
go run cmd/server/main.go

# In another terminal
curl -X POST http://localhost:8080/api/v1/spam/detect \
  -H "Content-Type: application/json" \
  -d '{"phone_number": "7379037972"}' | jq .
```

## File Structure

```
credCode/
├── api/
│   ├── handler.go      # HTTP handlers
│   └── server.go       # HTTP server
├── service/
│   ├── spam_rule.go              # Rule interface and registry
│   ├── spam_detection_service.go # Main service
│   └── rules/
│       ├── contact_count_rule.go # Contact count rule
│       └── call_pattern_rule.go  # Call pattern rule
├── models/
│   └── spam.go         # Spam detection models
└── cmd/
    └── server/
        └── main.go     # Server entry point
```

## Benefits

1. **Extensible**: Easy to add new rules
2. **Modular**: Each rule is independent
3. **Testable**: Rules can be tested individually
4. **Configurable**: Rules and thresholds are configurable
5. **Scalable**: Can handle many rules efficiently

## Future Enhancements

1. **Rule Weights**: Weight rules differently in average calculation
2. **Rule Dependencies**: Some rules depend on others
3. **Caching**: Cache rule results for performance
4. **Machine Learning**: Use ML models as rules
5. **Real-time Updates**: Update scores as new data arrives

