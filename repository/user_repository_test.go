package repository

import (
	"credCode/models"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// TestNewInMemoryUserRepository tests the creation of a new repository
func TestNewInMemoryUserRepository(t *testing.T) {
	repo := NewInMemoryUserRepository()
	if repo == nil {
		t.Fatal("Expected non-nil repository")
	}
	if repo.users == nil {
		t.Error("Expected users map to be initialized")
	}
	if repo.phones == nil {
		t.Error("Expected phones map to be initialized")
	}
}

// TestCreateUser tests user creation
func TestCreateUser(t *testing.T) {
	repo := NewInMemoryUserRepository()

	user := &models.User{
		ID:          "test1",
		PhoneNumber: "1111111111",
		Name:        "Test User",
		Contacts:    []*models.Contact{},
	}

	err := repo.CreateUser(user)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify user was added
	retrieved, err := repo.GetUserByID("test1")
	if err != nil {
		t.Fatalf("Expected to retrieve user, got error: %v", err)
	}
	if retrieved.Name != "Test User" {
		t.Errorf("Expected name 'Test User', got '%s'", retrieved.Name)
	}
}

// TestCreateUser_DuplicateID tests duplicate user ID error
func TestCreateUser_DuplicateID(t *testing.T) {
	repo := NewInMemoryUserRepository()

	user1 := &models.User{
		ID:          "test1",
		PhoneNumber: "1111111111",
		Name:        "Test User 1",
		Contacts:    []*models.Contact{},
	}

	user2 := &models.User{
		ID:          "test1",
		PhoneNumber: "2222222222",
		Name:        "Test User 2",
		Contacts:    []*models.Contact{},
	}

	err := repo.CreateUser(user1)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	err = repo.CreateUser(user2)
	if err != ErrUserExists {
		t.Errorf("Expected ErrUserExists, got: %v", err)
	}
}

// TestCreateUser_DuplicatePhone tests duplicate phone number error
func TestCreateUser_DuplicatePhone(t *testing.T) {
	repo := NewInMemoryUserRepository()

	user1 := &models.User{
		ID:          "test1",
		PhoneNumber: "1111111111",
		Name:        "Test User 1",
		Contacts:    []*models.Contact{},
	}

	user2 := &models.User{
		ID:          "test2",
		PhoneNumber: "1111111111",
		Name:        "Test User 2",
		Contacts:    []*models.Contact{},
	}

	err := repo.CreateUser(user1)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	err = repo.CreateUser(user2)
	if err != ErrUserExists {
		t.Errorf("Expected ErrUserExists, got: %v", err)
	}
}

// TestGetUserByID tests retrieving user by ID
func TestGetUserByID(t *testing.T) {
	repo := NewInMemoryUserRepository()

	user := &models.User{
		ID:          "test1",
		PhoneNumber: "1111111111",
		Name:        "Test User",
		Contacts:    []*models.Contact{},
	}

	repo.CreateUser(user)

	retrieved, err := repo.GetUserByID("test1")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if retrieved.ID != "test1" {
		t.Errorf("Expected ID 'test1', got '%s'", retrieved.ID)
	}
}

// TestGetUserByID_NotFound tests user not found error
func TestGetUserByID_NotFound(t *testing.T) {
	repo := NewInMemoryUserRepository()

	_, err := repo.GetUserByID("nonexistent")
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got: %v", err)
	}
}

// TestGetUserByPhoneNumber tests retrieving user by phone number
func TestGetUserByPhoneNumber(t *testing.T) {
	repo := NewInMemoryUserRepository()

	user := &models.User{
		ID:          "test1",
		PhoneNumber: "1111111111",
		Name:        "Test User",
		Contacts:    []*models.Contact{},
	}

	repo.CreateUser(user)

	retrieved, err := repo.GetUserByPhoneNumber("1111111111")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if retrieved.PhoneNumber != "1111111111" {
		t.Errorf("Expected phone '1111111111', got '%s'", retrieved.PhoneNumber)
	}
}

// TestGetUserByPhoneNumber_NotFound tests phone not found error
func TestGetUserByPhoneNumber_NotFound(t *testing.T) {
	repo := NewInMemoryUserRepository()

	_, err := repo.GetUserByPhoneNumber("9999999999")
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got: %v", err)
	}
}

// TestGetAllUsers tests retrieving all users
func TestGetAllUsers(t *testing.T) {
	repo := NewInMemoryUserRepository()

	users := []*models.User{
		{ID: "1", PhoneNumber: "1111111111", Name: "User 1", Contacts: []*models.Contact{}},
		{ID: "2", PhoneNumber: "2222222222", Name: "User 2", Contacts: []*models.Contact{}},
		{ID: "3", PhoneNumber: "3333333333", Name: "User 3", Contacts: []*models.Contact{}},
	}

	for _, user := range users {
		repo.CreateUser(user)
	}

	allUsers, err := repo.GetAllUsers()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(allUsers) != 3 {
		t.Errorf("Expected 3 users, got %d", len(allUsers))
	}
}

// TestGetAllUsers_Empty tests retrieving users from empty repository
func TestGetAllUsers_Empty(t *testing.T) {
	repo := NewInMemoryUserRepository()

	allUsers, err := repo.GetAllUsers()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(allUsers) != 0 {
		t.Errorf("Expected 0 users, got %d", len(allUsers))
	}
}

// TestUpdateUser tests user update
func TestUpdateUser(t *testing.T) {
	repo := NewInMemoryUserRepository()

	user := &models.User{
		ID:          "test1",
		PhoneNumber: "1111111111",
		Name:        "Test User",
		Contacts:    []*models.Contact{},
	}

	repo.CreateUser(user)

	// Update user
	updatedUser := &models.User{
		ID:          "test1",
		PhoneNumber: "1111111111",
		Name:        "Updated User",
		Contacts:    []*models.Contact{},
	}

	err := repo.UpdateUser(updatedUser)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify update
	retrieved, _ := repo.GetUserByID("test1")
	if retrieved.Name != "Updated User" {
		t.Errorf("Expected name 'Updated User', got '%s'", retrieved.Name)
	}
}

// TestUpdateUser_PhoneNumberChange tests updating user phone number
func TestUpdateUser_PhoneNumberChange(t *testing.T) {
	repo := NewInMemoryUserRepository()

	user := &models.User{
		ID:          "test1",
		PhoneNumber: "1111111111",
		Name:        "Test User",
		Contacts:    []*models.Contact{},
	}

	repo.CreateUser(user)

	// Update phone number
	updatedUser := &models.User{
		ID:          "test1",
		PhoneNumber: "9999999999",
		Name:        "Test User",
		Contacts:    []*models.Contact{},
	}

	err := repo.UpdateUser(updatedUser)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify new phone number works
	retrieved, err := repo.GetUserByPhoneNumber("9999999999")
	if err != nil {
		t.Fatalf("Expected to find user by new phone, got error: %v", err)
	}
	if retrieved.ID != "test1" {
		t.Errorf("Expected ID 'test1', got '%s'", retrieved.ID)
	}

	// Verify old phone number doesn't work
	_, err = repo.GetUserByPhoneNumber("1111111111")
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound for old phone, got: %v", err)
	}
}

// TestUpdateUser_NotFound tests update non-existent user
func TestUpdateUser_NotFound(t *testing.T) {
	repo := NewInMemoryUserRepository()

	user := &models.User{
		ID:          "nonexistent",
		PhoneNumber: "1111111111",
		Name:        "Test User",
		Contacts:    []*models.Contact{},
	}

	err := repo.UpdateUser(user)
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got: %v", err)
	}
}

// TestDeleteUser tests user deletion
func TestDeleteUser(t *testing.T) {
	repo := NewInMemoryUserRepository()

	user := &models.User{
		ID:          "test1",
		PhoneNumber: "1111111111",
		Name:        "Test User",
		Contacts:    []*models.Contact{},
	}

	repo.CreateUser(user)

	err := repo.DeleteUser("test1")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify deletion
	_, err = repo.GetUserByID("test1")
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got: %v", err)
	}

	// Verify phone mapping is also deleted
	_, err = repo.GetUserByPhoneNumber("1111111111")
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound for phone, got: %v", err)
	}
}

// TestDeleteUser_NotFound tests deleting non-existent user
func TestDeleteUser_NotFound(t *testing.T) {
	repo := NewInMemoryUserRepository()

	err := repo.DeleteUser("nonexistent")
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got: %v", err)
	}
}

// TestAddContact tests adding a contact
func TestAddContact(t *testing.T) {
	repo := NewInMemoryUserRepository()

	user := &models.User{
		ID:          "test1",
		PhoneNumber: "1111111111",
		Name:        "Test User",
		Contacts:    []*models.Contact{},
	}

	repo.CreateUser(user)

	contact := &models.Contact{
		ID:          "c1",
		PhoneNumber: "2222222222",
		Name:        "Contact 1",
		AddedAt:     time.Now(),
	}

	err := repo.AddContact("test1", contact)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify contact was added
	contacts, _ := repo.GetUserContacts("test1")
	if len(contacts) != 1 {
		t.Errorf("Expected 1 contact, got %d", len(contacts))
	}
}

// TestAddContact_UserNotFound tests adding contact to non-existent user
func TestAddContact_UserNotFound(t *testing.T) {
	repo := NewInMemoryUserRepository()

	contact := &models.Contact{
		ID:          "c1",
		PhoneNumber: "2222222222",
		Name:        "Contact 1",
		AddedAt:     time.Now(),
	}

	err := repo.AddContact("nonexistent", contact)
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got: %v", err)
	}
}

// TestAddContact_Duplicate tests adding duplicate contact
func TestAddContact_Duplicate(t *testing.T) {
	repo := NewInMemoryUserRepository()

	user := &models.User{
		ID:          "test1",
		PhoneNumber: "1111111111",
		Name:        "Test User",
		Contacts:    []*models.Contact{},
	}

	repo.CreateUser(user)

	contact := &models.Contact{
		ID:          "c1",
		PhoneNumber: "2222222222",
		Name:        "Contact 1",
		AddedAt:     time.Now(),
	}

	repo.AddContact("test1", contact)

	// Try to add same contact again
	err := repo.AddContact("test1", contact)
	if err != ErrContactExists {
		t.Errorf("Expected ErrContactExists, got: %v", err)
	}
}

// TestGetContact tests retrieving a specific contact
func TestGetContact(t *testing.T) {
	repo := NewInMemoryUserRepository()

	user := &models.User{
		ID:          "test1",
		PhoneNumber: "1111111111",
		Name:        "Test User",
		Contacts:    []*models.Contact{},
	}

	repo.CreateUser(user)

	contact := &models.Contact{
		ID:          "c1",
		PhoneNumber: "2222222222",
		Name:        "Contact 1",
		AddedAt:     time.Now(),
	}

	repo.AddContact("test1", contact)

	retrieved, err := repo.GetContact("test1", "c1")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if retrieved.Name != "Contact 1" {
		t.Errorf("Expected name 'Contact 1', got '%s'", retrieved.Name)
	}
}

// TestGetContact_UserNotFound tests getting contact for non-existent user
func TestGetContact_UserNotFound(t *testing.T) {
	repo := NewInMemoryUserRepository()

	_, err := repo.GetContact("nonexistent", "c1")
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got: %v", err)
	}
}

// TestGetContact_NotFound tests getting non-existent contact
func TestGetContact_NotFound(t *testing.T) {
	repo := NewInMemoryUserRepository()

	user := &models.User{
		ID:          "test1",
		PhoneNumber: "1111111111",
		Name:        "Test User",
		Contacts:    []*models.Contact{},
	}

	repo.CreateUser(user)

	_, err := repo.GetContact("test1", "nonexistent")
	if err != ErrContactNotFound {
		t.Errorf("Expected ErrContactNotFound, got: %v", err)
	}
}

// TestGetUserContacts tests retrieving all contacts for a user
func TestGetUserContacts(t *testing.T) {
	repo := NewInMemoryUserRepository()

	user := &models.User{
		ID:          "test1",
		PhoneNumber: "1111111111",
		Name:        "Test User",
		Contacts:    []*models.Contact{},
	}

	repo.CreateUser(user)

	contacts := []*models.Contact{
		{ID: "c1", PhoneNumber: "2222222222", Name: "Contact 1", AddedAt: time.Now()},
		{ID: "c2", PhoneNumber: "3333333333", Name: "Contact 2", AddedAt: time.Now()},
		{ID: "c3", PhoneNumber: "4444444444", Name: "Contact 3", AddedAt: time.Now()},
	}

	for _, contact := range contacts {
		repo.AddContact("test1", contact)
	}

	retrieved, err := repo.GetUserContacts("test1")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(retrieved) != 3 {
		t.Errorf("Expected 3 contacts, got %d", len(retrieved))
	}
}

// TestGetUserContacts_UserNotFound tests getting contacts for non-existent user
func TestGetUserContacts_UserNotFound(t *testing.T) {
	repo := NewInMemoryUserRepository()

	_, err := repo.GetUserContacts("nonexistent")
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got: %v", err)
	}
}

// TestGetUserContacts_Empty tests getting contacts for user with no contacts
func TestGetUserContacts_Empty(t *testing.T) {
	repo := NewInMemoryUserRepository()

	user := &models.User{
		ID:          "test1",
		PhoneNumber: "1111111111",
		Name:        "Test User",
		Contacts:    []*models.Contact{},
	}

	repo.CreateUser(user)

	contacts, err := repo.GetUserContacts("test1")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(contacts) != 0 {
		t.Errorf("Expected 0 contacts, got %d", len(contacts))
	}
}

// TestUpdateContact tests updating a contact
func TestUpdateContact(t *testing.T) {
	repo := NewInMemoryUserRepository()

	user := &models.User{
		ID:          "test1",
		PhoneNumber: "1111111111",
		Name:        "Test User",
		Contacts:    []*models.Contact{},
	}

	repo.CreateUser(user)

	contact := &models.Contact{
		ID:          "c1",
		PhoneNumber: "2222222222",
		Name:        "Contact 1",
		AddedAt:     time.Now(),
	}

	repo.AddContact("test1", contact)

	// Update contact
	updatedContact := &models.Contact{
		ID:          "c1",
		PhoneNumber: "9999999999",
		Name:        "Updated Contact",
		AddedAt:     time.Now(),
	}

	err := repo.UpdateContact("test1", updatedContact)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify update
	retrieved, _ := repo.GetContact("test1", "c1")
	if retrieved.Name != "Updated Contact" {
		t.Errorf("Expected name 'Updated Contact', got '%s'", retrieved.Name)
	}
	if retrieved.PhoneNumber != "9999999999" {
		t.Errorf("Expected phone '9999999999', got '%s'", retrieved.PhoneNumber)
	}
}

// TestUpdateContact_UserNotFound tests updating contact for non-existent user
func TestUpdateContact_UserNotFound(t *testing.T) {
	repo := NewInMemoryUserRepository()

	contact := &models.Contact{
		ID:          "c1",
		PhoneNumber: "2222222222",
		Name:        "Contact 1",
		AddedAt:     time.Now(),
	}

	err := repo.UpdateContact("nonexistent", contact)
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got: %v", err)
	}
}

// TestUpdateContact_NotFound tests updating non-existent contact
func TestUpdateContact_NotFound(t *testing.T) {
	repo := NewInMemoryUserRepository()

	user := &models.User{
		ID:          "test1",
		PhoneNumber: "1111111111",
		Name:        "Test User",
		Contacts:    []*models.Contact{},
	}

	repo.CreateUser(user)

	contact := &models.Contact{
		ID:          "nonexistent",
		PhoneNumber: "2222222222",
		Name:        "Contact 1",
		AddedAt:     time.Now(),
	}

	err := repo.UpdateContact("test1", contact)
	if err != ErrContactNotFound {
		t.Errorf("Expected ErrContactNotFound, got: %v", err)
	}
}

// TestDeleteContact tests deleting a contact
func TestDeleteContact(t *testing.T) {
	repo := NewInMemoryUserRepository()

	user := &models.User{
		ID:          "test1",
		PhoneNumber: "1111111111",
		Name:        "Test User",
		Contacts:    []*models.Contact{},
	}

	repo.CreateUser(user)

	contact := &models.Contact{
		ID:          "c1",
		PhoneNumber: "2222222222",
		Name:        "Contact 1",
		AddedAt:     time.Now(),
	}

	repo.AddContact("test1", contact)

	err := repo.DeleteContact("test1", "c1")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify deletion
	_, err = repo.GetContact("test1", "c1")
	if err != ErrContactNotFound {
		t.Errorf("Expected ErrContactNotFound, got: %v", err)
	}
}

// TestDeleteContact_UserNotFound tests deleting contact for non-existent user
func TestDeleteContact_UserNotFound(t *testing.T) {
	repo := NewInMemoryUserRepository()

	err := repo.DeleteContact("nonexistent", "c1")
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got: %v", err)
	}
}

// TestDeleteContact_NotFound tests deleting non-existent contact
func TestDeleteContact_NotFound(t *testing.T) {
	repo := NewInMemoryUserRepository()

	user := &models.User{
		ID:          "test1",
		PhoneNumber: "1111111111",
		Name:        "Test User",
		Contacts:    []*models.Contact{},
	}

	repo.CreateUser(user)

	err := repo.DeleteContact("test1", "nonexistent")
	if err != ErrContactNotFound {
		t.Errorf("Expected ErrContactNotFound, got: %v", err)
	}
}

// TestLoadSeedData tests loading seed data from JSON file
func TestLoadSeedData(t *testing.T) {
	repo := NewInMemoryUserRepository()

	// Create a temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_seed.json")

	jsonData := `{
		"users": [
			{
				"id": "1",
				"phone_number": "1111111111",
				"name": "User 1",
				"contacts": [
					{
						"id": "c1",
						"phone_number": "2222222222",
						"name": "Contact 1",
						"added_at": "2024-01-15T10:30:00Z"
					}
				]
			},
			{
				"id": "2",
				"phone_number": "3333333333",
				"name": "User 2",
				"contacts": []
			}
		]
	}`

	err := os.WriteFile(testFile, []byte(jsonData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err = repo.LoadSeedData(testFile)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify data was loaded
	users, _ := repo.GetAllUsers()
	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}

	// Verify user 1 with contact
	user1, err := repo.GetUserByID("1")
	if err != nil {
		t.Fatalf("Expected to find user 1, got error: %v", err)
	}
	if user1.Name != "User 1" {
		t.Errorf("Expected name 'User 1', got '%s'", user1.Name)
	}
	if len(user1.Contacts) != 1 {
		t.Errorf("Expected 1 contact, got %d", len(user1.Contacts))
	}

	// Verify user can be found by phone
	user2, err := repo.GetUserByPhoneNumber("3333333333")
	if err != nil {
		t.Fatalf("Expected to find user by phone, got error: %v", err)
	}
	if user2.ID != "2" {
		t.Errorf("Expected ID '2', got '%s'", user2.ID)
	}
}

// TestLoadSeedData_FileNotFound tests loading from non-existent file
func TestLoadSeedData_FileNotFound(t *testing.T) {
	repo := NewInMemoryUserRepository()

	err := repo.LoadSeedData("nonexistent_file.json")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

// TestLoadSeedData_InvalidJSON tests loading invalid JSON
func TestLoadSeedData_InvalidJSON(t *testing.T) {
	repo := NewInMemoryUserRepository()

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "invalid.json")

	invalidJSON := `{invalid json}`
	err := os.WriteFile(testFile, []byte(invalidJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err = repo.LoadSeedData(testFile)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

// TestConcurrentUserOperations tests thread safety of user operations
func TestConcurrentUserOperations(t *testing.T) {
	repo := NewInMemoryUserRepository()
	var wg sync.WaitGroup

	// Concurrent creates
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			user := &models.User{
				ID:          string(rune('0' + id)),
				PhoneNumber: string(rune('0' + id)) + "000000000",
				Name:        "User " + string(rune('0' + id)),
				Contacts:    []*models.Contact{},
			}
			repo.CreateUser(user)
		}(i)
	}

	wg.Wait()

	// Verify all users were created
	users, _ := repo.GetAllUsers()
	if len(users) != 10 {
		t.Errorf("Expected 10 users after concurrent creates, got %d", len(users))
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			repo.GetUserByID(string(rune('0' + id)))
		}(i)
	}

	wg.Wait()
}

// TestConcurrentContactOperations tests thread safety of contact operations
func TestConcurrentContactOperations(t *testing.T) {
	repo := NewInMemoryUserRepository()

	user := &models.User{
		ID:          "test1",
		PhoneNumber: "1111111111",
		Name:        "Test User",
		Contacts:    []*models.Contact{},
	}

	repo.CreateUser(user)

	var wg sync.WaitGroup

	// Concurrent contact additions
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			contact := &models.Contact{
				ID:          "c" + string(rune('0' + id)),
				PhoneNumber: string(rune('0' + id)) + "000000000",
				Name:        "Contact " + string(rune('0' + id)),
				AddedAt:     time.Now(),
			}
			repo.AddContact("test1", contact)
		}(i)
	}

	wg.Wait()

	// Verify all contacts were added
	contacts, _ := repo.GetUserContacts("test1")
	if len(contacts) != 10 {
		t.Errorf("Expected 10 contacts after concurrent adds, got %d", len(contacts))
	}
}

// TestConcurrentMixedOperations tests thread safety with mixed operations
func TestConcurrentMixedOperations(t *testing.T) {
	repo := NewInMemoryUserRepository()

	// Create initial users
	for i := 0; i < 5; i++ {
		user := &models.User{
			ID:          string(rune('0' + i)),
			PhoneNumber: string(rune('0' + i)) + "000000000",
			Name:        "User " + string(rune('0' + i)),
			Contacts:    []*models.Contact{},
		}
		repo.CreateUser(user)
	}

	var wg sync.WaitGroup

	// Concurrent mixed operations
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			switch id % 4 {
			case 0: // Read
				repo.GetUserByID(string(rune('0' + (id % 5))))
			case 1: // Update
				user := &models.User{
					ID:          string(rune('0' + (id % 5))),
					PhoneNumber: string(rune('0' + (id % 5))) + "000000000",
					Name:        "Updated User",
					Contacts:    []*models.Contact{},
				}
				repo.UpdateUser(user)
			case 2: // GetAllUsers
				repo.GetAllUsers()
			case 3: // GetUserByPhoneNumber
				repo.GetUserByPhoneNumber(string(rune('0' + (id % 5))) + "000000000")
			}
		}(i)
	}

	wg.Wait()

	// Verify repository is still consistent
	users, err := repo.GetAllUsers()
	if err != nil {
		t.Fatalf("Expected no error after concurrent operations, got: %v", err)
	}
	if len(users) != 5 {
		t.Errorf("Expected 5 users after concurrent operations, got %d", len(users))
	}
}

// TestDeleteContact_MultipleContacts tests deleting one contact among many
func TestDeleteContact_MultipleContacts(t *testing.T) {
	repo := NewInMemoryUserRepository()

	user := &models.User{
		ID:          "test1",
		PhoneNumber: "1111111111",
		Name:        "Test User",
		Contacts:    []*models.Contact{},
	}

	repo.CreateUser(user)

	// Add multiple contacts
	contacts := []*models.Contact{
		{ID: "c1", PhoneNumber: "2222222222", Name: "Contact 1", AddedAt: time.Now()},
		{ID: "c2", PhoneNumber: "3333333333", Name: "Contact 2", AddedAt: time.Now()},
		{ID: "c3", PhoneNumber: "4444444444", Name: "Contact 3", AddedAt: time.Now()},
	}

	for _, contact := range contacts {
		repo.AddContact("test1", contact)
	}

	// Delete middle contact
	err := repo.DeleteContact("test1", "c2")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify deletion
	remaining, _ := repo.GetUserContacts("test1")
	if len(remaining) != 2 {
		t.Errorf("Expected 2 contacts remaining, got %d", len(remaining))
	}

	// Verify correct contacts remain
	found1, found3 := false, false
	for _, c := range remaining {
		if c.ID == "c1" {
			found1 = true
		}
		if c.ID == "c3" {
			found3 = true
		}
		if c.ID == "c2" {
			t.Error("Found deleted contact c2")
		}
	}

	if !found1 || !found3 {
		t.Error("Expected contacts c1 and c3 to remain")
	}
}

// TestUserRepositoryInterface verifies that InMemoryUserRepository implements UserRepository
func TestUserRepositoryInterface(t *testing.T) {
	var _ UserRepository = (*InMemoryUserRepository)(nil)
}

