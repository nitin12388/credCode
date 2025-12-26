package repository

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"

	"credCode/models"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrContactNotFound = errors.New("contact not found")
	ErrUserExists      = errors.New("user already exists")
	ErrContactExists   = errors.New("contact already exists")
)

// UserRepository defines the interface for user operations
type UserRepository interface {
	// User CRUD operations
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	GetUserByPhoneNumber(ctx context.Context, phone string) (*models.User, error)
	GetAllUsers(ctx context.Context) ([]*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id string) error

	// Contact CRUD operations
	AddContact(ctx context.Context, userID string, contact *models.Contact) error
	GetContact(ctx context.Context, userID, contactID string) (*models.Contact, error)
	GetUserContacts(ctx context.Context, userID string) ([]*models.Contact, error)
	UpdateContact(ctx context.Context, userID string, contact *models.Contact) error
	DeleteContact(ctx context.Context, userID, contactID string) error

	// Seed data operations
	LoadSeedData(ctx context.Context, filePath string) error
}

// InMemoryUserRepository implements UserRepository with in-memory storage
type InMemoryUserRepository struct {
	users  map[string]*models.User // key: user ID
	phones map[string]string       // key: phone number, value: user ID
	mu     sync.RWMutex
}

// NewInMemoryUserRepository creates a new in-memory user repository
func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users:  make(map[string]*models.User),
		phones: make(map[string]string),
	}
}

// CreateUser creates a new user
func (r *InMemoryUserRepository) CreateUser(ctx context.Context, user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.ID]; exists {
		return ErrUserExists
	}

	if _, exists := r.phones[user.PhoneNumber]; exists {
		return ErrUserExists
	}

	r.users[user.ID] = user
	r.phones[user.PhoneNumber] = user.ID
	return nil
}

// GetUserByID retrieves a user by ID
func (r *InMemoryUserRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// GetUserByPhoneNumber retrieves a user by phone number
func (r *InMemoryUserRepository) GetUserByPhoneNumber(ctx context.Context, phone string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	userID, exists := r.phones[phone]
	if !exists {
		return nil, ErrUserNotFound
	}

	user, exists := r.users[userID]
	if !exists {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// GetAllUsers retrieves all users
func (r *InMemoryUserRepository) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]*models.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}
	return users, nil
}

// UpdateUser updates an existing user
func (r *InMemoryUserRepository) UpdateUser(ctx context.Context, user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existingUser, exists := r.users[user.ID]
	if !exists {
		return ErrUserNotFound
	}

	// If phone number is being changed, update the phone map
	if existingUser.PhoneNumber != user.PhoneNumber {
		delete(r.phones, existingUser.PhoneNumber)
		r.phones[user.PhoneNumber] = user.ID
	}

	r.users[user.ID] = user
	return nil
}

// DeleteUser deletes a user
func (r *InMemoryUserRepository) DeleteUser(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[id]
	if !exists {
		return ErrUserNotFound
	}

	delete(r.users, id)
	delete(r.phones, user.PhoneNumber)
	return nil
}

// AddContact adds a contact to a user's contact list
func (r *InMemoryUserRepository) AddContact(ctx context.Context, userID string, contact *models.Contact) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[userID]
	if !exists {
		return ErrUserNotFound
	}

	// Check if contact already exists
	for _, c := range user.Contacts {
		if c.ID == contact.ID {
			return ErrContactExists
		}
	}

	user.Contacts = append(user.Contacts, contact)
	return nil
}

// GetContact retrieves a specific contact from a user's contact list
func (r *InMemoryUserRepository) GetContact(ctx context.Context, userID, contactID string) (*models.Contact, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[userID]
	if !exists {
		return nil, ErrUserNotFound
	}

	for _, contact := range user.Contacts {
		if contact.ID == contactID {
			return contact, nil
		}
	}

	return nil, ErrContactNotFound
}

// GetUserContacts retrieves all contacts for a user
func (r *InMemoryUserRepository) GetUserContacts(ctx context.Context, userID string) ([]*models.Contact, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[userID]
	if !exists {
		return nil, ErrUserNotFound
	}

	return user.Contacts, nil
}

// UpdateContact updates a contact in a user's contact list
func (r *InMemoryUserRepository) UpdateContact(ctx context.Context, userID string, contact *models.Contact) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[userID]
	if !exists {
		return ErrUserNotFound
	}

	for i, c := range user.Contacts {
		if c.ID == contact.ID {
			user.Contacts[i] = contact
			return nil
		}
	}

	return ErrContactNotFound
}

// DeleteContact removes a contact from a user's contact list
func (r *InMemoryUserRepository) DeleteContact(ctx context.Context, userID, contactID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[userID]
	if !exists {
		return ErrUserNotFound
	}

	for i, contact := range user.Contacts {
		if contact.ID == contactID {
			user.Contacts = append(user.Contacts[:i], user.Contacts[i+1:]...)
			return nil
		}
	}

	return ErrContactNotFound
}

// LoadSeedData loads seed data from a JSON file
func (r *InMemoryUserRepository) LoadSeedData(ctx context.Context, filePath string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Read the JSON file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Parse the JSON
	var seedData struct {
		Users []*models.User `json:"users"`
	}

	if err := json.Unmarshal(data, &seedData); err != nil {
		return err
	}

	// Load users into the repository
	for _, user := range seedData.Users {
		r.users[user.ID] = user
		r.phones[user.PhoneNumber] = user.ID
	}

	return nil
}
