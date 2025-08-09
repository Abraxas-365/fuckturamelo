package usersrv

import (
	"context"
	"fmt"

	"github.com/Abraxas-365/craftable/errx"
	"github.com/Abraxas-365/fuckturamelo/user"
)

var (
	userErrors = errx.NewRegistry("USER")

	ErrUserNotFound    = userErrors.Register("USER_NOT_FOUND", errx.TypeNotFound, 404, "User not found")
	ErrUserExists      = userErrors.Register("USER_EXISTS", errx.TypeBadRequest, 400, "User already exists")
	ErrProfileNotFound = userErrors.Register("PROFILE_NOT_FOUND", errx.TypeNotFound, 404, "User profile not found")
	ErrInvalidEmail    = userErrors.Register("INVALID_EMAIL", errx.TypeBadRequest, 400, "Invalid email format")
	ErrInvalidUserData = userErrors.Register("INVALID_USER_DATA", errx.TypeBadRequest, 400, "Invalid user data")
	ErrInvalidUUID     = userErrors.Register("INVALID_UUID", errx.TypeBadRequest, 400, "Invalid UUID format")
)

type userService struct {
	userRepo    user.UserRepository
	profileRepo user.UserProfileRepository
}

func NewUserService(
	userRepo user.UserRepository,
	profileRepo user.UserProfileRepository,
) user.UserService {
	return &userService{
		userRepo:    userRepo,
		profileRepo: profileRepo,
	}
}

func (s *userService) CreateUser(ctx context.Context, email, name string) (*user.User, error) {
	// Validate input
	if email == "" || name == "" {
		return nil, userErrors.New(ErrInvalidUserData).
			WithDetail("message", "Email and name are required")
	}

	// Check if user exists
	existingUser, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil && !errx.IsCode(err, ErrUserNotFound) {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	if existingUser != nil {
		return nil, userErrors.New(ErrUserExists).WithDetail("email", email)
	}

	user := user.NewUser(email, name)

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *userService) GetUser(ctx context.Context, id string) (*user.User, error) {
	if id == "" {
		return nil, userErrors.New(ErrInvalidUserData).
			WithDetail("message", "User ID is required")
	}

	if !user.IsValidUUID(id) {
		return nil, userErrors.New(ErrInvalidUUID).
			WithDetail("user_id", id)
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errx.IsCode(err, ErrUserNotFound) {
			return nil, userErrors.New(ErrUserNotFound).WithDetail("user_id", id)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
	if email == "" {
		return nil, userErrors.New(ErrInvalidEmail).
			WithDetail("message", "Email is required")
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errx.IsCode(err, ErrUserNotFound) {
			return nil, userErrors.New(ErrUserNotFound).WithDetail("email", email)
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return user, nil
}

func (s *userService) UpdateUser(ctx context.Context, id, name string) (*user.User, error) {
	if id == "" || name == "" {
		return nil, userErrors.New(ErrInvalidUserData).
			WithDetail("message", "User ID and name are required")
	}

	if !user.IsValidUUID(id) {
		return nil, userErrors.New(ErrInvalidUUID).
			WithDetail("user_id", id)
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errx.IsCode(err, ErrUserNotFound) {
			return nil, userErrors.New(ErrUserNotFound).WithDetail("user_id", id)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user.UpdateName(name)

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

func (s *userService) DeactivateUser(ctx context.Context, id string) error {
	if id == "" {
		return userErrors.New(ErrInvalidUserData).
			WithDetail("message", "User ID is required")
	}

	if !user.IsValidUUID(id) {
		return userErrors.New(ErrInvalidUUID).
			WithDetail("user_id", id)
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errx.IsCode(err, ErrUserNotFound) {
			return userErrors.New(ErrUserNotFound).WithDetail("user_id", id)
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	user.Deactivate()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to deactivate user: %w", err)
	}

	return nil
}

func (s *userService) ActivateUser(ctx context.Context, id string) error {
	if id == "" {
		return userErrors.New(ErrInvalidUserData).
			WithDetail("message", "User ID is required")
	}

	if !user.IsValidUUID(id) {
		return userErrors.New(ErrInvalidUUID).
			WithDetail("user_id", id)
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errx.IsCode(err, ErrUserNotFound) {
			return userErrors.New(ErrUserNotFound).WithDetail("user_id", id)
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	user.Activate()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to activate user: %w", err)
	}

	return nil
}

func (s *userService) CreateUserProfile(ctx context.Context, userID, firstName, lastName string) (*user.UserProfile, error) {
	if userID == "" || firstName == "" || lastName == "" {
		return nil, userErrors.New(ErrInvalidUserData).
			WithDetail("message", "User ID, first name, and last name are required")
	}

	if !user.IsValidUUID(userID) {
		return nil, userErrors.New(ErrInvalidUUID).
			WithDetail("user_id", userID)
	}

	// Verify user exists
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errx.IsCode(err, ErrUserNotFound) {
			return nil, userErrors.New(ErrUserNotFound).WithDetail("user_id", userID)
		}
		return nil, fmt.Errorf("failed to verify user exists: %w", err)
	}

	// Check if profile already exists
	existingProfile, err := s.profileRepo.GetByUserID(ctx, userID)
	if err != nil && !errx.IsCode(err, ErrProfileNotFound) {
		return nil, fmt.Errorf("failed to check existing profile: %w", err)
	}

	if existingProfile != nil {
		return nil, userErrors.New(ErrUserExists).
			WithDetail("message", "User profile already exists").
			WithDetail("user_id", userID)
	}

	profile := user.NewUserProfile(userID, firstName, lastName)

	if err := s.profileRepo.Create(ctx, profile); err != nil {
		return nil, fmt.Errorf("failed to create profile: %w", err)
	}

	return profile, nil
}

func (s *userService) GetUserProfile(ctx context.Context, userID string) (*user.UserProfile, error) {
	if userID == "" {
		return nil, userErrors.New(ErrInvalidUserData).
			WithDetail("message", "User ID is required")
	}

	if !user.IsValidUUID(userID) {
		return nil, userErrors.New(ErrInvalidUUID).
			WithDetail("user_id", userID)
	}

	profile, err := s.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errx.IsCode(err, ErrProfileNotFound) {
			return nil, userErrors.New(ErrProfileNotFound).WithDetail("user_id", userID)
		}
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}
	return profile, nil
}

func (s *userService) UpdateUserProfile(ctx context.Context, userID, firstName, lastName, bio string) (*user.UserProfile, error) {
	if userID == "" {
		return nil, userErrors.New(ErrInvalidUserData).
			WithDetail("message", "User ID is required")
	}

	if !user.IsValidUUID(userID) {
		return nil, userErrors.New(ErrInvalidUUID).
			WithDetail("user_id", userID)
	}

	profile, err := s.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errx.IsCode(err, ErrProfileNotFound) {
			return nil, userErrors.New(ErrProfileNotFound).WithDetail("user_id", userID)
		}
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	if firstName != "" && lastName != "" {
		profile.UpdateName(firstName, lastName)
	}

	if bio != "" {
		profile.UpdateBio(bio)
	}

	if err := s.profileRepo.Update(ctx, profile); err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return profile, nil
}

func (s *userService) UpdateUserProfilePicture(ctx context.Context, userID, pictureURL string) (*user.UserProfile, error) {
	if userID == "" || pictureURL == "" {
		return nil, userErrors.New(ErrInvalidUserData).
			WithDetail("message", "User ID and picture URL are required")
	}

	if !user.IsValidUUID(userID) {
		return nil, userErrors.New(ErrInvalidUUID).
			WithDetail("user_id", userID)
	}

	profile, err := s.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errx.IsCode(err, ErrProfileNotFound) {
			return nil, userErrors.New(ErrProfileNotFound).WithDetail("user_id", userID)
		}
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	profile.SetProfilePicture(pictureURL)

	if err := s.profileRepo.Update(ctx, profile); err != nil {
		return nil, fmt.Errorf("failed to update profile picture: %w", err)
	}

	return profile, nil
}

func (s *userService) SetUserPhone(ctx context.Context, userID, phone string) (*user.UserProfile, error) {
	if userID == "" || phone == "" {
		return nil, userErrors.New(ErrInvalidUserData).
			WithDetail("message", "User ID and phone are required")
	}

	if !user.IsValidUUID(userID) {
		return nil, userErrors.New(ErrInvalidUUID).
			WithDetail("user_id", userID)
	}

	profile, err := s.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errx.IsCode(err, ErrProfileNotFound) {
			return nil, userErrors.New(ErrProfileNotFound).WithDetail("user_id", userID)
		}
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	profile.SetPhone(phone)

	if err := s.profileRepo.Update(ctx, profile); err != nil {
		return nil, fmt.Errorf("failed to set phone: %w", err)
	}

	return profile, nil
}

