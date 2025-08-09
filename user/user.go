package user

import (
	"time"

	"github.com/google/uuid"
)

// User aggregate root
type User struct {
	id        string
	email     string
	name      string
	isActive  bool
	createdAt time.Time
	updatedAt time.Time
}

func NewUser(email, name string) *User {
	now := time.Now()
	return &User{
		id:        uuid.New().String(), // Generate UUID as string
		email:     email,
		name:      name,
		isActive:  true,
		createdAt: now,
		updatedAt: now,
	}
}

// NewUserWithID creates a user with a specific ID (for reconstruction from DB)
func NewUserWithID(id, email, name string) *User {
	now := time.Now()
	return &User{
		id:        id,
		email:     email,
		name:      name,
		isActive:  true,
		createdAt: now,
		updatedAt: now,
	}
}

// Domain getters
func (u *User) ID() string           { return u.id }
func (u *User) Email() string        { return u.email }
func (u *User) Name() string         { return u.name }
func (u *User) IsActive() bool       { return u.isActive }
func (u *User) CreatedAt() time.Time { return u.createdAt }
func (u *User) UpdatedAt() time.Time { return u.updatedAt }

// Implement auth.User interface
func (u *User) GetID() string    { return u.id }
func (u *User) GetEmail() string { return u.email }

// Business methods
func (u *User) UpdateName(name string) {
	u.name = name
	u.updatedAt = time.Now()
}

func (u *User) UpdateEmail(email string) {
	u.email = email
	u.updatedAt = time.Now()
}

func (u *User) Deactivate() {
	u.isActive = false
	u.updatedAt = time.Now()
}

func (u *User) Activate() {
	u.isActive = true
	u.updatedAt = time.Now()
}

// UserProfile value object
type UserProfile struct {
	userID         string
	firstName      string
	lastName       string
	profilePicture *string
	bio            string
	phone          *string
	createdAt      time.Time
	updatedAt      time.Time
}

func NewUserProfile(userID, firstName, lastName string) *UserProfile {
	now := time.Now()
	return &UserProfile{
		userID:    userID,
		firstName: firstName,
		lastName:  lastName,
		createdAt: now,
		updatedAt: now,
	}
}

func (up *UserProfile) UserID() string          { return up.userID }
func (up *UserProfile) FirstName() string       { return up.firstName }
func (up *UserProfile) LastName() string        { return up.lastName }
func (up *UserProfile) ProfilePicture() *string { return up.profilePicture }
func (up *UserProfile) Bio() string             { return up.bio }
func (up *UserProfile) Phone() *string          { return up.phone }
func (up *UserProfile) CreatedAt() time.Time    { return up.createdAt }
func (up *UserProfile) UpdatedAt() time.Time    { return up.updatedAt }

func (up *UserProfile) UpdateName(firstName, lastName string) {
	up.firstName = firstName
	up.lastName = lastName
	up.updatedAt = time.Now()
}

func (up *UserProfile) SetProfilePicture(url string) {
	up.profilePicture = &url
	up.updatedAt = time.Now()
}

func (up *UserProfile) UpdateBio(bio string) {
	up.bio = bio
	up.updatedAt = time.Now()
}

func (up *UserProfile) SetPhone(phone string) {
	up.phone = &phone
	up.updatedAt = time.Now()
}

// Helper functions for UUID validation
func IsValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

func GenerateUUID() string {
	return uuid.New().String()
}
