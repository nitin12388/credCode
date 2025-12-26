package models

import "time"

// Contact represents a contact in a user's contact list
type Contact struct {
	ID          string    `json:"id"`
	PhoneNumber string    `json:"phone_number"`
	Name        string    `json:"name"`
	AddedAt     time.Time `json:"added_at"`
}

// User represents a user in the system
type User struct {
	ID          string     `json:"id"`
	PhoneNumber string     `json:"phone_number"`
	Name        string     `json:"name"`
	Contacts    []*Contact `json:"contacts"`
}

