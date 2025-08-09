package authsrv

import (
	"time"
)

// OAuthToken value object
type OAuthToken struct {
	accessToken  string
	refreshToken string
	expiresAt    time.Time
}

func NewOAuthToken(accessToken, refreshToken string, expiresAt time.Time) *OAuthToken {
	return &OAuthToken{
		accessToken:  accessToken,
		refreshToken: refreshToken,
		expiresAt:    expiresAt,
	}
}

func (t *OAuthToken) AccessToken() string  { return t.accessToken }
func (t *OAuthToken) RefreshToken() string { return t.refreshToken }
func (t *OAuthToken) ExpiresAt() time.Time { return t.expiresAt }
func (t *OAuthToken) IsExpired() bool      { return time.Now().After(t.expiresAt) }

// OAuthAccount entity
type OAuthAccount struct {
	id            string
	userID        string
	provider      string
	providerID    string
	providerEmail string
	token         *OAuthToken
	metadata      map[string]any
	createdAt     time.Time
	updatedAt     time.Time
}

func NewOAuthAccount(id, userID, provider, providerID, providerEmail string, token *OAuthToken) *OAuthAccount {
	now := time.Now()
	return &OAuthAccount{
		id:            id,
		userID:        userID,
		provider:      provider,
		providerID:    providerID,
		providerEmail: providerEmail,
		token:         token,
		metadata:      make(map[string]any),
		createdAt:     now,
		updatedAt:     now,
	}
}

func (oa *OAuthAccount) ID() string            { return oa.id }
func (oa *OAuthAccount) UserID() string        { return oa.userID }
func (oa *OAuthAccount) Provider() string      { return oa.provider }
func (oa *OAuthAccount) ProviderID() string    { return oa.providerID }
func (oa *OAuthAccount) ProviderEmail() string { return oa.providerEmail }
func (oa *OAuthAccount) Token() *OAuthToken    { return oa.token }

func (oa *OAuthAccount) UpdateToken(token *OAuthToken) {
	oa.token = token
	oa.updatedAt = time.Now()
}

// Provider user info value object
type ProviderUserInfo struct {
	providerID     string
	email          string
	name           string
	provider       string
	profilePicture *string
	token          *OAuthToken
	rawData        map[string]any
}

func NewProviderUserInfo(providerID, email, name, provider string, token *OAuthToken) *ProviderUserInfo {
	return &ProviderUserInfo{
		providerID: providerID,
		email:      email,
		name:       name,
		provider:   provider,
		token:      token,
		rawData:    make(map[string]any),
	}
}

func (pui *ProviderUserInfo) ProviderID() string { return pui.providerID }
func (pui *ProviderUserInfo) Email() string      { return pui.email }
func (pui *ProviderUserInfo) Name() string       { return pui.name }
func (pui *ProviderUserInfo) Provider() string   { return pui.provider }
func (pui *ProviderUserInfo) Token() *OAuthToken { return pui.token }

// Authentication result
type AuthenticationResult struct {
	userID                  string
	accessToken             string
	expiresIn               int
	defaultOrganizationID   *string
	organizationMemberships []string // membership IDs
}

func NewAuthenticationResult(userID, accessToken string, expiresIn int) *AuthenticationResult {
	return &AuthenticationResult{
		userID:      userID,
		accessToken: accessToken,
		expiresIn:   expiresIn,
	}
}

func (ar *AuthenticationResult) UserID() string                    { return ar.userID }
func (ar *AuthenticationResult) AccessToken() string               { return ar.accessToken }
func (ar *AuthenticationResult) ExpiresIn() int                    { return ar.expiresIn }
func (ar *AuthenticationResult) DefaultOrganizationID() *string    { return ar.defaultOrganizationID }
func (ar *AuthenticationResult) OrganizationMemberships() []string { return ar.organizationMemberships }

func (ar *AuthenticationResult) SetDefaultOrganization(orgID string) {
	ar.defaultOrganizationID = &orgID
}

func (ar *AuthenticationResult) SetOrganizationMemberships(membershipIDs []string) {
	ar.organizationMemberships = membershipIDs
}

// Token claims
type TokenClaims struct {
	UserID                  string    `json:"user_id"`
	Email                   string    `json:"email"`
	DefaultOrganizationID   *string   `json:"default_org_id,omitempty"`
	OrganizationMemberships []string  `json:"org_memberships,omitempty"`
	ExpiresAt               time.Time `json:"exp"`
	IssuedAt                time.Time `json:"iat"`
}
