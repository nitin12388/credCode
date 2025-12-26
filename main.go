package main

import (
	"fmt"
	"log"
	"time"

	"credCode/models"
	"credCode/repository"
)

func main() {
	// Initialize the in-memory repository
	repo := repository.NewInMemoryUserRepository()

	// Load seed data
	fmt.Println("Loading seed data...")
	if err := repo.LoadSeedData("repository/seed_data.json"); err != nil {
		log.Fatalf("Failed to load seed data: %v", err)
	}
	fmt.Println("Seed data loaded successfully!\n")

	// Demonstrate CRUD operations
	demonstrateUserOperations(repo)
	demonstrateContactOperations(repo)
}

func demonstrateUserOperations(repo repository.UserRepository) {
	fmt.Println("=== USER OPERATIONS ===")

	// Get all users
	users, err := repo.GetAllUsers()
	if err != nil {
		log.Printf("Error getting users: %v", err)
		return
	}
	fmt.Printf("Total users loaded: %d\n", len(users))

	// Get user by ID
	user, err := repo.GetUserByID("1")
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
	user, err = repo.GetUserByPhoneNumber("9876543210")
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
	if err := repo.CreateUser(newUser); err != nil {
		log.Printf("Error creating user: %v", err)
	} else {
		fmt.Printf("\nCreated new user: %s (Phone: %s)\n", newUser.Name, newUser.PhoneNumber)
	}

	// Update user
	user.Name = "Priya Kumar (Updated)"
	if err := repo.UpdateUser(user); err != nil {
		log.Printf("Error updating user: %v", err)
	} else {
		fmt.Printf("Updated user: %s\n", user.Name)
	}

	fmt.Println()
}

func demonstrateContactOperations(repo repository.UserRepository) {
	fmt.Println("=== CONTACT OPERATIONS ===")

	// Get user contacts
	contacts, err := repo.GetUserContacts("1")
	if err != nil {
		log.Printf("Error getting contacts: %v", err)
		return
	}
	fmt.Printf("\nContacts for User 1:\n")
	for _, contact := range contacts {
		fmt.Printf("  - %s (%s) added at %s\n",
			contact.Name,
			contact.PhoneNumber,
			contact.AddedAt.Format("2006-01-02"))
	}

	// Add a new contact
	newContact := &models.Contact{
		ID:          "c10",
		PhoneNumber: "1111222233",
		Name:        "New Contact",
		AddedAt:     time.Now(),
	}
	if err := repo.AddContact("1", newContact); err != nil {
		log.Printf("Error adding contact: %v", err)
	} else {
		fmt.Printf("\nAdded new contact: %s to User 1\n", newContact.Name)
	}

	// Get specific contact
	contact, err := repo.GetContact("1", "c1")
	if err != nil {
		log.Printf("Error getting contact: %v", err)
	} else {
		fmt.Printf("\nContact c1 for User 1:\n")
		fmt.Printf("  Name: %s\n", contact.Name)
		fmt.Printf("  Phone: %s\n", contact.PhoneNumber)
	}

	// Update contact
	contact.Name = "Arjun (Updated)"
	if err := repo.UpdateContact("1", contact); err != nil {
		log.Printf("Error updating contact: %v", err)
	} else {
		fmt.Printf("Updated contact: %s\n", contact.Name)
	}

	// Delete contact
	if err := repo.DeleteContact("1", "c2"); err != nil {
		log.Printf("Error deleting contact: %v", err)
	} else {
		fmt.Println("\nDeleted contact c2 from User 1")
	}

	// Verify deletion
	remainingContacts, _ := repo.GetUserContacts("1")
	fmt.Printf("Remaining contacts for User 1: %d\n", len(remainingContacts))
}
