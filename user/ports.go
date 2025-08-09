package user

import (
	"context"
)

// Primary ports (driving/input)
type UserService interface {
	CreateUser(ctx context.Context, email, name string) (*User, error)
	GetUser(ctx context.Context, id string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUser(ctx context.Context, id, name string) (*User, error)
	DeactivateUser(ctx context.Context, id string) error
	ActivateUser(ctx context.Context, id string) error

	CreateUserProfile(ctx context.Context, userID, firstName, lastName string) (*UserProfile, error)
	GetUserProfile(ctx context.Context, userID string) (*UserProfile, error)
	UpdateUserProfile(ctx context.Context, userID, firstName, lastName, bio string) (*UserProfile, error)
	UpdateUserProfilePicture(ctx context.Context, userID, pictureURL string) (*UserProfile, error)
	SetUserPhone(ctx context.Context, userID, phone string) (*UserProfile, error)
}

// Secondary ports (driven/output)
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	GetAll(ctx context.Context, limit, offset int) ([]*User, error)
}

type UserProfileRepository interface {
	Create(ctx context.Context, profile *UserProfile) error
	GetByUserID(ctx context.Context, userID string) (*UserProfile, error)
	Update(ctx context.Context, profile *UserProfile) error
	Delete(ctx context.Context, userID string) error
}

