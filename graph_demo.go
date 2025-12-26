package main

import (
	"fmt"
	"log"
	"time"

	"credCode/models"
	"credCode/repository"
)

func main() {
	// Initialize the graph repository
	graphRepo := repository.NewInMemoryGraphRepository()

	fmt.Println("=== GRAPH DATABASE DEMO ===\n")

	// Demo 1: Add nodes and contact edges
	demonstrateContactGraph(graphRepo)

	// Demo 2: Add call edges and query with filters
	demonstrateCallGraph(graphRepo)
}

func demonstrateContactGraph(repo repository.GraphRepository) {
	fmt.Println("--- Contact Graph Operations ---\n")

	// Add nodes
	phones := []string{"7379037972", "9876543210", "1234567890", "5555555555"}
	for _, phone := range phones {
		if err := repo.AddNode(phone); err != nil {
			log.Printf("Error adding node %s: %v", phone, err)
		}
	}
	fmt.Printf("Added %d nodes\n", len(phones))

	// Add contact relationships (bidirectional)
	// 9876543210 has contact 7379037972
	repo.AddContactEdge("9876543210", "7379037972")
	
	// 1234567890 has contact 7379037972
	repo.AddContactEdge("1234567890", "7379037972")
	
	// 5555555555 has contact 7379037972
	repo.AddContactEdge("5555555555", "7379037972")

	fmt.Println("Added contact relationships\n")

	// Query 1: Who has saved phone number 7379037972 in their contacts?
	fmt.Println("Query 1: Who has saved 7379037972 in their contacts?")
	users, count := repo.GetUsersWithContact("7379037972")
	fmt.Printf("Count: %d\n", count)
	fmt.Printf("Users: %v\n\n", users)

	// Query for another number
	fmt.Println("Query 1: Who has saved 9876543210 in their contacts?")
	users, count = repo.GetUsersWithContact("9876543210")
	fmt.Printf("Count: %d\n", count)
	fmt.Printf("Users: %v\n\n", users)
}

func demonstrateCallGraph(repo repository.GraphRepository) {
	fmt.Println("--- Call Graph Operations ---\n")

	now := time.Now()
	
	// Phone 7379037972 makes multiple calls
	phone := "7379037972"

	// Call 1: Answered, 120 seconds, 2 hours ago
	repo.AddCallEdge(phone, "9876543210", true, 120, now.Add(-2*time.Hour))
	
	// Call 2: Not answered, 15 seconds, 1 hour ago
	repo.AddCallEdge(phone, "1234567890", false, 15, now.Add(-1*time.Hour))
	
	// Call 3: Answered, 300 seconds, 30 minutes ago
	repo.AddCallEdge(phone, "5555555555", true, 300, now.Add(-30*time.Minute))
	
	// Call 4: Not answered, 10 seconds, 10 minutes ago
	repo.AddCallEdge(phone, "9876543210", false, 10, now.Add(-10*time.Minute))
	
	// Call 5: Answered, 5 seconds, 5 minutes ago
	repo.AddCallEdge(phone, "1234567890", true, 5, now.Add(-5*time.Minute))

	fmt.Printf("Added 5 call edges from %s\n\n", phone)

	// Query 2: Get all outgoing calls
	fmt.Println("Query 2: All outgoing calls from 7379037972")
	calls, count := repo.GetCallsWithFilters(phone, repository.CallFilters{}, "outgoing")
	fmt.Printf("Total calls: %d\n", count)
	printCallEdges(calls)

	// Query 2.1: Calls with is_answered = true
	fmt.Println("Query 2.1: Answered calls from 7379037972")
	answered := true
	calls, count = repo.GetCallsWithFilters(phone, repository.CallFilters{
		IsAnswered: &answered,
	}, "outgoing")
	fmt.Printf("Answered calls: %d\n", count)
	printCallEdges(calls)

	// Query 2.2: Calls with duration < 20 seconds
	fmt.Println("Query 2.2: Calls with duration < 20 seconds from 7379037972")
	maxDuration := 20
	calls, count = repo.GetCallsWithFilters(phone, repository.CallFilters{
		MaxDuration: &maxDuration,
	}, "outgoing")
	fmt.Printf("Short calls (< 20s): %d\n", count)
	printCallEdges(calls)

	// Query 2.3: Unanswered calls with duration < 20 seconds
	fmt.Println("Query 2.3: Unanswered calls with duration < 20s from 7379037972")
	notAnswered := false
	calls, count = repo.GetCallsWithFilters(phone, repository.CallFilters{
		IsAnswered:  &notAnswered,
		MaxDuration: &maxDuration,
	}, "outgoing")
	fmt.Printf("Unanswered short calls: %d\n", count)
	printCallEdges(calls)

	// Query 2.4: Calls in the last 1 hour
	fmt.Println("Query 2.4: Calls in the last 1 hour from 7379037972")
	oneHourAgo := now.Add(-1 * time.Hour)
	calls, count = repo.GetCallsWithFilters(phone, repository.CallFilters{
		TimeRangeStart: &oneHourAgo,
	}, "outgoing")
	fmt.Printf("Calls in last hour: %d\n", count)
	printCallEdges(calls)

	// Query 2.5: Complex filter - Answered calls with duration > 100 seconds in last 3 hours
	fmt.Println("Query 2.5: Answered calls > 100s in last 3 hours from 7379037972")
	threeHoursAgo := now.Add(-3 * time.Hour)
	minDuration := 100
	calls, count = repo.GetCallsWithFilters(phone, repository.CallFilters{
		IsAnswered:     &answered,
		MinDuration:    &minDuration,
		TimeRangeStart: &threeHoursAgo,
	}, "outgoing")
	fmt.Printf("Long answered calls: %d\n", count)
	printCallEdges(calls)
}

func printCallEdges(edges []*models.Edge) {
	for i, edge := range edges {
		props := models.ParseCallProperties(edge.Properties)
		fmt.Printf("  %d. %s -> %s | Answered: %v | Duration: %ds | Time: %s\n",
			i+1,
			edge.From,
			edge.To,
			props.IsAnswered,
			props.DurationInSeconds,
			edge.CreatedAt.Format("15:04:05"))
	}
	fmt.Println()
}

