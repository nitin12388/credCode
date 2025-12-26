package repository

import (
	"context"
	"testing"
	"time"

	"credCode/models"
)

func TestNewInMemoryUserRepository(t *testing.T) {
	repo := NewInMemoryUserRepository()

	if repo == nil {
		t.Fatal("Expected repository to be created, got nil")
	}

	if repo.users == nil {
		t.Error("Expected users map to be initialized")
	}

	if repo.phones == nil {
		t.Error("Expected phones map to be initialized")
	}
}

func TestInMemoryUserRepository_CreateUser(t *testing.T) {
	repo := NewInMemoryUserRepository()
	ctx := context.Background()

	user := &models.User{
		ID:          "1",
		PhoneNumber: "7379037972",
		Name:        "John Doe",
		Contacts:    []*models.Contact{},
	}

	// Test successful creation
	err := repo.CreateUser(ctx, user)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Test duplicate ID
	err = repo.CreateUser(ctx, user)
	if err != ErrUserExists {
		t.Errorf("Expected ErrUserExists, got %v", err)
	}

	// Test duplicate phone number
	user2 := &models.User{
		ID:          "2",
		PhoneNumber: "7379037972", // Same phone
		Name:        "Jane Doe",
		Contacts:    []*models.Contact{},
	}
	err = repo.CreateUser(ctx, user2)
	if err != ErrUserExists {
		t.Errorf("Expected ErrUserExists for duplicate phone, got %v", err)
	}
}

func TestInMemoryUserRepository_GetUserByID(t *testing.T) {
	repo := NewInMemoryUserRepository()
	ctx := context.Background()

	user := &models.User{
		ID:          "1",
		PhoneNumber: "7379037972",
		Name:        "John Doe",
		Contacts:    []*models.Contact{},
	}

	repo.CreateUser(ctx, user)

	// Test successful retrieval
	retrieved, err := repo.GetUserByID(ctx, "1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if retrieved.ID != "1" {
		t.Errorf("Expected ID '1', got '%s'", retrieved.ID)
	}

	// Test not found
	_, err = repo.GetUserByID(ctx, "999")
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestInMemoryUserRepository_GetUserByPhoneNumber(t *testing.T) {
	repo := NewInMemoryUserRepository()
	ctx := context.Background()

	user := &models.User{
		ID:          "1",
		PhoneNumber: "7379037972",
		Name:        "John Doe",
		Contacts:    []*models.Contact{},
	}

	repo.CreateUser(ctx, user)

	// Test successful retrieval
	retrieved, err := repo.GetUserByPhoneNumber(ctx, "7379037972")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if retrieved.PhoneNumber != "7379037972" {
		t.Errorf("Expected phone '7379037972', got '%s'", retrieved.PhoneNumber)
	}

	// Test not found
	_, err = repo.GetUserByPhoneNumber(ctx, "9999999999")
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestInMemoryUserRepository_GetAllUsers(t *testing.T) {
	repo := NewInMemoryUserRepository()
	ctx := context.Background()

	// Create multiple users
	users := []*models.User{
		{ID: "1", PhoneNumber: "7379037972", Name: "John", Contacts: []*models.Contact{}},
		{ID: "2", PhoneNumber: "9876543210", Name: "Jane", Contacts: []*models.Contact{}},
		{ID: "3", PhoneNumber: "1234567890", Name: "Bob", Contacts: []*models.Contact{}},
	}

	for _, user := range users {
		repo.CreateUser(ctx, user)
	}

	allUsers, err := repo.GetAllUsers(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(allUsers) != 3 {
		t.Errorf("Expected 3 users, got %d", len(allUsers))
	}
}

func TestInMemoryUserRepository_UpdateUser(t *testing.T) {
	repo := NewInMemoryUserRepository()
	ctx := context.Background()

	user := &models.User{
		ID:          "1",
		PhoneNumber: "7379037972",
		Name:        "John Doe",
		Contacts:    []*models.Contact{},
	}

	repo.CreateUser(ctx, user)

	// Update user
	user.Name = "John Updated"
	err := repo.UpdateUser(ctx, user)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify update
	updated, _ := repo.GetUserByID(ctx, "1")
	if updated.Name != "John Updated" {
		t.Errorf("Expected name 'John Updated', got '%s'", updated.Name)
	}

	// Test update non-existent user
	nonExistent := &models.User{
		ID:          "999",
		PhoneNumber: "9999999999",
		Name:        "Non Existent",
		Contacts:    []*models.Contact{},
	}
	err = repo.UpdateUser(ctx, nonExistent)
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestInMemoryUserRepository_DeleteUser(t *testing.T) {
	repo := NewInMemoryUserRepository()
	ctx := context.Background()

	user := &models.User{
		ID:          "1",
		PhoneNumber: "7379037972",
		Name:        "John Doe",
		Contacts:    []*models.Contact{},
	}

	repo.CreateUser(ctx, user)

	// Test successful deletion
	err := repo.DeleteUser(ctx, "1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify deletion
	_, err = repo.GetUserByID(ctx, "1")
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound after deletion, got %v", err)
	}

	// Test delete non-existent user
	err = repo.DeleteUser(ctx, "999")
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestInMemoryUserRepository_AddContact(t *testing.T) {
	repo := NewInMemoryUserRepository()
	ctx := context.Background()

	user := &models.User{
		ID:          "1",
		PhoneNumber: "7379037972",
		Name:        "John Doe",
		Contacts:    []*models.Contact{},
	}

	repo.CreateUser(ctx, user)

	contact := &models.Contact{
		ID:          "c1",
		PhoneNumber: "9876543210",
		Name:        "Jane",
		AddedAt:     time.Now(),
	}

	// Test successful addition
	err := repo.AddContact(ctx, "1", contact)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify contact was added
	contacts, _ := repo.GetUserContacts(ctx, "1")
	if len(contacts) != 1 {
		t.Errorf("Expected 1 contact, got %d", len(contacts))
	}

	// Test duplicate contact
	err = repo.AddContact(ctx, "1", contact)
	if err != ErrContactExists {
		t.Errorf("Expected ErrContactExists, got %v", err)
	}

	// Test add to non-existent user
	err = repo.AddContact(ctx, "999", contact)
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestInMemoryUserRepository_GetContact(t *testing.T) {
	repo := NewInMemoryUserRepository()
	ctx := context.Background()

	user := &models.User{
		ID:          "1",
		PhoneNumber: "7379037972",
		Name:        "John Doe",
		Contacts:    []*models.Contact{},
	}

	repo.CreateUser(ctx, user)

	contact := &models.Contact{
		ID:          "c1",
		PhoneNumber: "9876543210",
		Name:        "Jane",
		AddedAt:     time.Now(),
	}

	repo.AddContact(ctx, "1", contact)

	// Test successful retrieval
	retrieved, err := repo.GetContact(ctx, "1", "c1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if retrieved.ID != "c1" {
		t.Errorf("Expected contact ID 'c1', got '%s'", retrieved.ID)
	}

	// Test not found
	_, err = repo.GetContact(ctx, "1", "c999")
	if err != ErrContactNotFound {
		t.Errorf("Expected ErrContactNotFound, got %v", err)
	}
}

func TestInMemoryUserRepository_GetUserContacts(t *testing.T) {
	repo := NewInMemoryUserRepository()
	ctx := context.Background()

	user := &models.User{
		ID:          "1",
		PhoneNumber: "7379037972",
		Name:        "John Doe",
		Contacts:    []*models.Contact{},
	}

	repo.CreateUser(ctx, user)

	contacts := []*models.Contact{
		{ID: "c1", PhoneNumber: "9876543210", Name: "Jane", AddedAt: time.Now()},
		{ID: "c2", PhoneNumber: "1234567890", Name: "Bob", AddedAt: time.Now()},
	}

	for _, contact := range contacts {
		repo.AddContact(ctx, "1", contact)
	}

	// Test retrieval
	retrieved, err := repo.GetUserContacts(ctx, "1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(retrieved) != 2 {
		t.Errorf("Expected 2 contacts, got %d", len(retrieved))
	}
}

func TestInMemoryUserRepository_UpdateContact(t *testing.T) {
	repo := NewInMemoryUserRepository()
	ctx := context.Background()

	user := &models.User{
		ID:          "1",
		PhoneNumber: "7379037972",
		Name:        "John Doe",
		Contacts:    []*models.Contact{},
	}

	repo.CreateUser(ctx, user)

	contact := &models.Contact{
		ID:          "c1",
		PhoneNumber: "9876543210",
		Name:        "Jane",
		AddedAt:     time.Now(),
	}

	repo.AddContact(ctx, "1", contact)

	// Update contact
	contact.Name = "Jane Updated"
	err := repo.UpdateContact(ctx, "1", contact)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify update
	updated, _ := repo.GetContact(ctx, "1", "c1")
	if updated.Name != "Jane Updated" {
		t.Errorf("Expected name 'Jane Updated', got '%s'", updated.Name)
	}
}

func TestInMemoryUserRepository_DeleteContact(t *testing.T) {
	repo := NewInMemoryUserRepository()
	ctx := context.Background()

	user := &models.User{
		ID:          "1",
		PhoneNumber: "7379037972",
		Name:        "John Doe",
		Contacts:    []*models.Contact{},
	}

	repo.CreateUser(ctx, user)

	contact := &models.Contact{
		ID:          "c1",
		PhoneNumber: "9876543210",
		Name:        "Jane",
		AddedAt:     time.Now(),
	}

	repo.AddContact(ctx, "1", contact)

	// Test successful deletion
	err := repo.DeleteContact(ctx, "1", "c1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify deletion
	_, err = repo.GetContact(ctx, "1", "c1")
	if err != ErrContactNotFound {
		t.Errorf("Expected ErrContactNotFound after deletion, got %v", err)
	}
}
