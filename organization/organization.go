package organization

import (
	"fmt"
	"slices"
	"time"
)

// Role value object
type Role struct {
	name        string
	permissions []Permission
}

func NewRole(name string, permissions ...Permission) *Role {
	return &Role{
		name:        name,
		permissions: permissions,
	}
}

func (r *Role) Name() string              { return r.name }
func (r *Role) Permissions() []Permission { return r.permissions }

func (r *Role) HasPermission(permission Permission) bool {
	return slices.Contains(r.permissions, permission)
}

// Permission enumeration
type Permission string

const (
	PermissionViewOrg        Permission = "org.view"
	PermissionManageOrg      Permission = "org.manage"
	PermissionInviteMembers  Permission = "org.invite_members"
	PermissionRemoveMembers  Permission = "org.remove_members"
	PermissionViewProducts   Permission = "products.view"
	PermissionCreateProducts Permission = "products.create"
	PermissionManageProducts Permission = "products.manage"
	PermissionDeleteProducts Permission = "products.delete"
	PermissionViewMembers    Permission = "members.view"
	PermissionManageMembers  Permission = "members.manage"
)

// Predefined roles
var (
	RoleOrgAdmin = NewRole("org_admin",
		PermissionViewOrg, PermissionManageOrg,
		PermissionInviteMembers, PermissionRemoveMembers,
		PermissionViewProducts, PermissionViewMembers, PermissionManageMembers)

	RoleOrgMember = NewRole("org_member",
		PermissionViewOrg, PermissionViewProducts, PermissionViewMembers)

	RoleProductProvider = NewRole("product_provider",
		PermissionViewOrg, PermissionViewProducts,
		PermissionCreateProducts, PermissionManageProducts, PermissionDeleteProducts)
)

// Organization aggregate root
type Organization struct {
	id          string
	name        string
	description string
	ownerUserID string
	isActive    bool
	createdAt   time.Time
	updatedAt   time.Time
}

func NewOrganization(id, name, description, ownerUserID string) *Organization {
	now := time.Now()
	return &Organization{
		id:          id,
		name:        name,
		description: description,
		ownerUserID: ownerUserID,
		isActive:    true,
		createdAt:   now,
		updatedAt:   now,
	}
}

func (o *Organization) ID() string           { return o.id }
func (o *Organization) Name() string         { return o.name }
func (o *Organization) Description() string  { return o.description }
func (o *Organization) OwnerUserID() string  { return o.ownerUserID }
func (o *Organization) IsActive() bool       { return o.isActive }
func (o *Organization) CreatedAt() time.Time { return o.createdAt }
func (o *Organization) UpdatedAt() time.Time { return o.updatedAt }

func (o *Organization) UpdateName(name string) {
	o.name = name
	o.updatedAt = time.Now()
}

func (o *Organization) UpdateDescription(description string) {
	o.description = description
	o.updatedAt = time.Now()
}

func (o *Organization) TransferOwnership(newOwnerUserID string) {
	o.ownerUserID = newOwnerUserID
	o.updatedAt = time.Now()
}

// Membership entity
type Membership struct {
	id             string
	userID         string
	organizationID string
	role           *Role
	isActive       bool
	joinedAt       time.Time
	updatedAt      time.Time
}

func NewMembership(id, userID, organizationID string, role *Role) *Membership {
	now := time.Now()
	return &Membership{
		id:             id,
		userID:         userID,
		organizationID: organizationID,
		role:           role,
		isActive:       true,
		joinedAt:       now,
		updatedAt:      now,
	}
}

func (m *Membership) ID() string             { return m.id }
func (m *Membership) UserID() string         { return m.userID }
func (m *Membership) OrganizationID() string { return m.organizationID }
func (m *Membership) Role() *Role            { return m.role }
func (m *Membership) IsActive() bool         { return m.isActive }
func (m *Membership) JoinedAt() time.Time    { return m.joinedAt }
func (m *Membership) UpdatedAt() time.Time   { return m.updatedAt }

func (m *Membership) UpdateRole(role *Role) {
	m.role = role
	m.updatedAt = time.Now()
}

func (m *Membership) Deactivate() {
	m.isActive = false
	m.updatedAt = time.Now()
}

func (m *Membership) HasPermission(permission Permission) bool {
	return m.isActive && m.role.HasPermission(permission)
}

// Invitation entity
type Invitation struct {
	id             string
	organizationID string
	inviterUserID  string
	email          string
	role           *Role
	token          string
	isUsed         bool
	expiresAt      time.Time
	createdAt      time.Time
}

func NewInvitation(id, organizationID, inviterUserID, email string, role *Role, token string, expiresAt time.Time) *Invitation {
	return &Invitation{
		id:             id,
		organizationID: organizationID,
		inviterUserID:  inviterUserID,
		email:          email,
		role:           role,
		token:          token,
		isUsed:         false,
		expiresAt:      expiresAt,
		createdAt:      time.Now(),
	}
}

func (i *Invitation) ID() string             { return i.id }
func (i *Invitation) OrganizationID() string { return i.organizationID }
func (i *Invitation) InviterUserID() string  { return i.inviterUserID }
func (i *Invitation) Email() string          { return i.email }
func (i *Invitation) Role() *Role            { return i.role }
func (i *Invitation) Token() string          { return i.token }
func (i *Invitation) IsUsed() bool           { return i.isUsed }
func (i *Invitation) ExpiresAt() time.Time   { return i.expiresAt }
func (i *Invitation) CreatedAt() time.Time   { return i.createdAt }

func (i *Invitation) IsExpired() bool {
	return time.Now().After(i.expiresAt)
}

func (i *Invitation) Use() error {
	if i.isUsed {
		return fmt.Errorf("invitation already used")
	}
	if i.IsExpired() {
		return fmt.Errorf("invitation expired")
	}
	i.isUsed = true
	return nil
}
