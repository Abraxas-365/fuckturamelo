package organization

import (
	"context"
)

// Primary ports
type OrganizationService interface {
	CreateOrganization(ctx context.Context, creatorUserID, name, description string) (*Organization, error)
	GetOrganization(ctx context.Context, id string) (*Organization, error)
	GetUserOrganizations(ctx context.Context, userID string) ([]*Organization, error)
	UpdateOrganization(ctx context.Context, organizationID, name, description string) (*Organization, error)

	// Membership management
	InviteUser(ctx context.Context, organizationID, inviterUserID, email string, role *Role) (*Invitation, error)
	AcceptInvitation(ctx context.Context, token string, userID string) (*Membership, error)
	GetOrganizationMembers(ctx context.Context, organizationID string) ([]*Membership, error)
	UpdateMemberRole(ctx context.Context, organizationID, userID string, role *Role) (*Membership, error)
	RemoveMember(ctx context.Context, organizationID, userID string) error
}

// Secondary ports
type OrganizationRepository interface {
	Create(ctx context.Context, org *Organization) error
	GetByID(ctx context.Context, id string) (*Organization, error)
	Update(ctx context.Context, org *Organization) error
	GetByUserID(ctx context.Context, userID string) ([]*Organization, error)
}

type MembershipRepository interface {
	Create(ctx context.Context, membership *Membership) error
	GetByUserAndOrganization(ctx context.Context, userID, organizationID string) (*Membership, error)
	GetByOrganization(ctx context.Context, organizationID string) ([]*Membership, error)
	GetByUser(ctx context.Context, userID string) ([]*Membership, error)
	Update(ctx context.Context, membership *Membership) error
	Delete(ctx context.Context, userID, organizationID string) error
}

type InvitationRepository interface {
	Create(ctx context.Context, invitation *Invitation) error
	GetByToken(ctx context.Context, token string) (*Invitation, error)
	GetByOrganization(ctx context.Context, organizationID string) ([]*Invitation, error)
	Update(ctx context.Context, invitation *Invitation) error
}

type IDGenerator interface {
	Generate() string
}
