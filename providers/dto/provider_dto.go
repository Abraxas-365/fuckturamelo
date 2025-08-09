package dto

import (
	"time"

	"github.com/Abraxas-365/craftable/validatex"
	"github.com/google/uuid"
)

// CreateProviderRequest represents the request to create a new provider
type CreateProviderRequest struct {
	UserID         *uuid.UUID     `json:"user_id,omitempty" validate:"omitempty,uuid"`
	OrganizationID uuid.UUID      `json:"organization_id" validate:"required,uuid"`
	Name           string         `json:"name" validate:"required,min=1,max=255"`
	ProviderCode   *string        `json:"provider_code,omitempty" validate:"omitempty,max=50"`
	IsActive       *bool          `json:"is_active,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
}

// UpdateProviderRequest represents the request to update an existing provider
type UpdateProviderRequest struct {
	UserID       *uuid.UUID     `json:"user_id,omitempty" validate:"omitempty,uuid"`
	Name         *string        `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	ProviderCode *string        `json:"provider_code,omitempty" validate:"omitempty,max=50"`
	IsActive     *bool          `json:"is_active,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

// ProviderResponse represents the response containing provider data
type ProviderResponse struct {
	ID             uuid.UUID      `json:"id"`
	UserID         *uuid.UUID     `json:"user_id,omitempty"`
	OrganizationID uuid.UUID      `json:"organization_id"`
	Name           string         `json:"name"`
	ProviderCode   *string        `json:"provider_code,omitempty"`
	IsActive       bool           `json:"is_active"`
	Metadata       map[string]any `json:"metadata"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// ProviderListRequest represents the request for listing providers with filters
type ProviderListRequest struct {
	OrganizationID *uuid.UUID `json:"organization_id,omitempty" validate:"omitempty,uuid"`
	IsActive       *bool      `json:"is_active,omitempty"`
	Search         *string    `json:"search,omitempty" validate:"omitempty,max=255"`
	Page           int        `json:"page,omitempty" validate:"omitempty,min=1"`
	PageSize       int        `json:"page_size,omitempty" validate:"omitempty,min=1,max=100"`
	OrderBy        *string    `json:"order_by,omitempty" validate:"omitempty,oneof=name created_at updated_at"`
	Desc           bool       `json:"desc,omitempty"`
}

// ProviderListResponse represents the paginated response for providers
type ProviderListResponse struct {
	Data       []ProviderResponse `json:"data"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	Total      int                `json:"total"`
	TotalPages int                `json:"total_pages"`
	HasNext    bool               `json:"has_next"`
	HasPrev    bool               `json:"has_prev"`
}

// Validate validates the CreateProviderRequest
func (r *CreateProviderRequest) Validate() error {
	return validatex.Validate(r)
}

// Validate validates the UpdateProviderRequest
func (r *UpdateProviderRequest) Validate() error {
	return validatex.Validate(r)
}

// Validate validates the ProviderListRequest
func (r *ProviderListRequest) Validate() error {
	return validatex.Validate(r)
}
