package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"credCode/models"
	"credCode/repository"
)

func main() {
	fmt.Println("=== TRUECALLER SYSTEM DEMO ===\n")

	// Initialize repositories
	userRepo := repository.NewInMemoryUserRepository()
	graphRepo := repository.NewInMemoryGraphRepository()

	// Load user seed data
	fmt.Println("--- Loading User Data ---")
	if err := userRepo.LoadSeedData(context.Background(), "contacts_generated.json"); err != nil {
		log.Fatalf("Failed to load seed data: %v", err)
	}
	fmt.Println("✓ User seed data loaded successfully!\n")

	// Build graph from user data
	fmt.Println("--- Building Graph from User Data ---")
	buildGraphFromUsers(userRepo, graphRepo)
	fmt.Println("✓ Graph constructed successfully!\n")

	//Demonstrate user operations
	demonstrateUserOperations(userRepo)

	//Demonstrate graph operations
	demonstrateGraphOperations(graphRepo)

	// Demonstrate integration: Query graph based on user data
	demonstrateIntegration(userRepo, graphRepo)
}

// buildGraphFromUsers constructs the graph from user repository data
func buildGraphFromUsers(userRepo repository.UserRepository, graphRepo repository.GraphRepository) {
	// Get all users
	users, err := userRepo.GetAllUsers(context.Background())
	if err != nil {
		log.Printf("Error getting users: %v", err)
		return
	}

	// Add nodes to graph with names
	for _, user := range users {
		if err := graphRepo.AddNodeWithName(context.Background(), user.PhoneNumber, user.Name); err != nil {
			// Node might already exist, continue
			log.Printf("Note: Node %s already exists or error: %v", user.PhoneNumber, err)
		}
	}
	fmt.Printf("  Added %d nodes to graph\n", len(users))

	// Add contact edges from user contacts
	contactCount := 0
	for _, user := range users {
		contacts, err := userRepo.GetUserContacts(context.Background(), user.ID)
		if err != nil {
			continue
		}

		for _, contact := range contacts {
			// Add contact edge with metadata
			contactMeta := &models.ContactMetadata{
				Name:    contact.Name,
				AddedAt: contact.AddedAt,
			}

			// Add bidirectional contact edge
			_, err := graphRepo.AddEdgeWithMetadata(context.Background(), user.PhoneNumber, contact.PhoneNumber, contactMeta)
			if err != nil {
				log.Printf("Error adding contact edge: %v", err)
				continue
			}
			contactCount++
		}
	}
	fmt.Printf("  Added %d contact edges to graph\n", contactCount)
}

func demonstrateUserOperations(repo repository.UserRepository) {
	ctx := context.Background()
	fmt.Println("=== USER OPERATIONS ===")

	// Get all users
	users, err := repo.GetAllUsers(ctx)
	if err != nil {
		log.Printf("Error getting users: %v", err)
		return
	}
	fmt.Printf("Total users: %d\n", len(users))

	// Get user by ID
	user, err := repo.GetUserByID(ctx, "1")
	if err != nil {
		log.Printf("Error getting user: %v", err)
		return
	}
	fmt.Printf("\nUser by ID (1):\n")
	fmt.Printf("  ID: %s\n", user.ID)
	fmt.Printf("  Name: %s\n", user.Name)
	fmt.Printf("  Phone: %s\n", user.PhoneNumber)
	fmt.Printf("  Contacts: %d\n", len(user.Contacts))

	// Get user by phone number
	user, err = repo.GetUserByPhoneNumber(ctx, "9876543210")
	if err != nil {
		log.Printf("Error getting user by phone: %v", err)
		return
	}
	fmt.Printf("\nUser by Phone (9876543210):\n")
	fmt.Printf("  ID: %s\n", user.ID)
	fmt.Printf("  Name: %s\n", user.Name)

	// Create a new user
	newUser := &models.User{
		ID:          "4",
		PhoneNumber: "5555123456",
		Name:        "New User",
		Contacts:    []*models.Contact{},
	}
	if err := repo.CreateUser(ctx, newUser); err != nil {
		log.Printf("Error creating user: %v", err)
	} else {
		fmt.Printf("\nCreated new user: %s (Phone: %s)\n", newUser.Name, newUser.PhoneNumber)
	}

	// Update user
	user.Name = "Priya Kumar (Updated)"
	if err := repo.UpdateUser(ctx, user); err != nil {
		log.Printf("Error updating user: %v", err)
	} else {
		fmt.Printf("Updated user: %s\n", user.Name)
	}

	fmt.Println()
}

func demonstrateGraphOperations(repo repository.GraphRepository) {
	ctx := context.Background()
	fmt.Println("=== GRAPH OPERATIONS ===")

	// Query 1: Who has saved a phone number in their contacts?
	fmt.Println("\n--- Query 1: Contact Count ---")
	phoneNumber := "7379037972"
	users, count := repo.GetUsersWithContact(ctx, phoneNumber)
	fmt.Printf("Phone number %s is saved by %d users:\n", phoneNumber, count)
	for i, user := range users {
		fmt.Printf("  %d. %s\n", i+1, user)
	}

	// Get node with name
	node, err := repo.GetNode(ctx, phoneNumber)
	if err == nil {
		fmt.Printf("\nNode details:\n")
		fmt.Printf("  Phone: %s\n", node.PhoneNumber)
		fmt.Printf("  Name: %s\n", node.Name)
	}

	// Add some call edges for demonstration
	fmt.Println("\n--- Adding Call Edges ---")
	now := time.Now()

	// Call 1: Answered, 120 seconds
	callMeta1 := &models.CallMetadata{
		IsAnswered:        true,
		DurationInSeconds: 120,
		Timestamp:         now.Add(-2 * time.Hour),
	}
	edge1, _ := repo.AddEdgeWithMetadata(ctx, "7379037972", "9876543210", callMeta1)
	fmt.Printf("✓ Added call: %s -> %s (Answered: %v, Duration: %ds)\n",
		edge1.From, edge1.To, callMeta1.IsAnswered, callMeta1.DurationInSeconds)

	// Call 2: Not answered, 15 seconds
	callMeta2 := &models.CallMetadata{
		IsAnswered:        false,
		DurationInSeconds: 15,
		Timestamp:         now.Add(-1 * time.Hour),
	}
	edge2, _ := repo.AddEdgeWithMetadata(ctx, "7379037972", "1234567890", callMeta2)
	fmt.Printf("✓ Added call: %s -> %s (Answered: %v, Duration: %ds)\n",
		edge2.From, edge2.To, callMeta2.IsAnswered, callMeta2.DurationInSeconds)

	// Query 2: Call filtering
	fmt.Println("\n--- Query 2: Call Filtering ---")

	// Get all outgoing calls
	calls, count := repo.GetCallsWithFilters(ctx, "7379037972", repository.CallFilters{}, "outgoing")
	fmt.Printf("Total outgoing calls: %d\n", count)
	for i, call := range calls {
		if cm, ok := call.Metadata.(*models.CallMetadata); ok {
			fmt.Printf("  %d. -> %s | Answered: %v | Duration: %ds | Time: %s\n",
				i+1, call.To, cm.IsAnswered, cm.DurationInSeconds,
				cm.Timestamp.Format("15:04:05"))
		}
	}

	// Filter: Answered calls only
	answered := true
	calls, count = repo.GetCallsWithFilters(ctx, "7379037972", repository.CallFilters{
		IsAnswered: &answered,
	}, "outgoing")
	fmt.Printf("\nAnswered calls: %d\n", count)
	for i, call := range calls {
		if cm, ok := call.Metadata.(*models.CallMetadata); ok {
			fmt.Printf("  %d. -> %s (Duration: %ds)\n",
				i+1, call.To, cm.DurationInSeconds)
		}
	}

	// Filter: Short calls (< 20 seconds)
	maxDuration := 20
	calls, count = repo.GetCallsWithFilters(ctx, "7379037972", repository.CallFilters{
		MaxDuration: &maxDuration,
	}, "outgoing")
	fmt.Printf("\nShort calls (< 20s): %d\n", count)
	for i, call := range calls {
		if cm, ok := call.Metadata.(*models.CallMetadata); ok {
			fmt.Printf("  %d. -> %s (Duration: %ds)\n",
				i+1, call.To, cm.DurationInSeconds)
		}
	}

	// Filter: Unanswered calls in last hour
	notAnswered := false
	oneHourAgo := now.Add(-1 * time.Hour)
	calls, count = repo.GetCallsWithFilters(ctx, "7379037972", repository.CallFilters{
		IsAnswered:     &notAnswered,
		TimeRangeStart: &oneHourAgo,
	}, "outgoing")
	fmt.Printf("\nUnanswered calls in last hour: %d\n", count)
	for i, call := range calls {
		if cm, ok := call.Metadata.(*models.CallMetadata); ok {
			fmt.Printf("  %d. -> %s (Duration: %ds)\n",
				i+1, call.To, cm.DurationInSeconds)
		}
	}

	fmt.Println()
}

func demonstrateIntegration(userRepo repository.UserRepository, graphRepo repository.GraphRepository) {
	fmt.Println("=== INTEGRATION DEMO ===")
	fmt.Println("Showing how user data and graph work together:\n")

	// Get a user
	user, err := userRepo.GetUserByID(context.Background(), "1")
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("User: %s (%s)\n", user.Name, user.PhoneNumber)
	fmt.Printf("Has %d contacts in user repository\n", len(user.Contacts))

	// Check graph for this user's contacts
	contactEdges := graphRepo.GetOutgoingEdges(context.Background(), user.PhoneNumber, models.EdgeTypeContact)
	fmt.Printf("Has %d contact edges in graph\n", len(contactEdges))

	// Show contact details from graph
	if len(contactEdges) > 0 {
		fmt.Println("\nContact details from graph:")
		for i, edge := range contactEdges {
			if cm, ok := edge.Metadata.(*models.ContactMetadata); ok {
				fmt.Printf("  %d. %s (%s) - Added: %s\n",
					i+1, cm.Name, edge.To, cm.AddedAt.Format("2006-01-02"))
			} else {
				fmt.Printf("  %d. %s - No metadata\n", i+1, edge.To)
			}
		}
	}

	// Query: Who has this user saved?
	usersWithContact, count := graphRepo.GetUsersWithContact(context.Background(), user.PhoneNumber)
	fmt.Printf("\n%s is saved by %d users in their contacts:\n", user.Name, count)
	for i, phone := range usersWithContact {
		// Get user name from user repo
		contactUser, _ := userRepo.GetUserByPhoneNumber(context.Background(), phone)
		name := phone
		if contactUser != nil {
			name = contactUser.Name
		}
		fmt.Printf("  %d. %s (%s)\n", i+1, name, phone)
	}

	fmt.Println()
}
