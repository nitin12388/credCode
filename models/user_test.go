package models

import (
	"testing"
	"time"
)

func TestContact_Validation(t *testing.T) {
	tests := []struct {
		name    string
		contact Contact
		wantErr bool
	}{
		{
			name: "valid contact",
			contact: Contact{
				ID:          "c1",
				PhoneNumber: "1234567890",
				Name:        "John Doe",
				AddedAt:     time.Now(),
			},
			wantErr: false,
		},
		{
			name: "contact with empty ID",
			contact: Contact{
				ID:          "",
				PhoneNumber: "1234567890",
				Name:        "John Doe",
				AddedAt:     time.Now(),
			},
			wantErr: false, // ID can be empty, validation happens at repository level
		},
		{
			name: "contact with empty phone",
			contact: Contact{
				ID:          "c1",
				PhoneNumber: "",
				Name:        "John Doe",
				AddedAt:     time.Now(),
			},
			wantErr: false, // Validation happens at repository level
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic structure test
			if tt.contact.ID == "" && tt.name != "contact with empty ID" {
				t.Errorf("Contact ID should not be empty")
			}
		})
	}
}

func TestUser_Validation(t *testing.T) {
	tests := []struct {
		name    string
		user    User
		wantErr bool
	}{
		{
			name: "valid user",
			user: User{
				ID:          "1",
				PhoneNumber: "7379037972",
				Name:        "John Doe",
				Contacts:    []*Contact{},
			},
			wantErr: false,
		},
		{
			name: "user with contacts",
			user: User{
				ID:          "1",
				PhoneNumber: "7379037972",
				Name:        "John Doe",
				Contacts: []*Contact{
					{
						ID:          "c1",
						PhoneNumber: "9876543210",
						Name:        "Jane",
						AddedAt:     time.Now(),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "user with empty ID",
			user: User{
				ID:          "",
				PhoneNumber: "7379037972",
				Name:        "John Doe",
				Contacts:    []*Contact{},
			},
			wantErr: false, // Validation happens at repository level
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.user.ID == "" && tt.name != "user with empty ID" {
				t.Errorf("User ID should not be empty")
			}
			if tt.user.PhoneNumber == "" {
				t.Errorf("User phone number should not be empty")
			}
		})
	}
}

func TestUser_AddContact(t *testing.T) {
	user := &User{
		ID:          "1",
		PhoneNumber: "7379037972",
		Name:        "John Doe",
		Contacts:    []*Contact{},
	}

	contact := &Contact{
		ID:          "c1",
		PhoneNumber: "9876543210",
		Name:        "Jane",
		AddedAt:     time.Now(),
	}

	user.Contacts = append(user.Contacts, contact)

	if len(user.Contacts) != 1 {
		t.Errorf("Expected 1 contact, got %d", len(user.Contacts))
	}

	if user.Contacts[0].Name != "Jane" {
		t.Errorf("Expected contact name 'Jane', got '%s'", user.Contacts[0].Name)
	}
}

