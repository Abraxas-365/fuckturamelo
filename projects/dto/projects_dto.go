package dto

import (
	"github.com/Abraxas-365/fuckturamelo/projects/models"
	"github.com/google/uuid"
)

// CreateProjectRequest represents the request payload for creating a project
type CreateProjectRequest struct {
	OrganizationID uuid.UUID       `json:"organization_id" validate:"required"`
	Name           string          `json:"name" validate:"required,min=1,max=255"`
	Description    *string         `json:"description" validate:"omitempty,max=1000"`
	Metadata       models.Metadata `json:"metadata"`
	ProviderIDs    []uuid.UUID     `json:"provider_ids,omitempty"`
}

// UpdateProjectRequest represents the request payload for updating a project
type UpdateProjectRequest struct {
	Name        *string         `json:"name" validate:"omitempty,min=1,max=255"`
	Description *string         `json:"description" validate:"omitempty,max=1000"`
	IsActive    *bool           `json:"is_active"`
	Metadata    models.Metadata `json:"metadata"`
}

// ProjectListRequest represents query parameters for listing projects
type ProjectListRequest struct {
	OrganizationID *uuid.UUID `query:"organization_id"`
	IsActive       *bool      `query:"is_active"`
	Search         *string    `query:"search"`
	Page           int        `query:"page" validate:"min=1"`
	PageSize       int        `query:"page_size" validate:"min=1,max=100"`
	SortBy         string     `query:"sort_by"`
	SortOrder      string     `query:"sort_order" validate:"omitempty,oneof=asc desc"`
}

// AddProviderRequest represents the request to add a provider to a project
type AddProviderRequest struct {
	ProviderID uuid.UUID `json:"provider_id" validate:"required"`
	Role       *string   `json:"role" validate:"omitempty,max=100"`
}

// UpdateProviderRoleRequest represents the request to update a provider's role
type UpdateProviderRoleRequest struct {
	Role *string `json:"role" validate:"omitempty,max=100"`
}

// BulkProviderRequest represents bulk operations on project providers
type BulkProviderRequest struct {
	ProviderIDs []uuid.UUID `json:"provider_ids" validate:"required,min=1"`
}

// ProjectResponse represents the response for a single project
type ProjectResponse struct {
	*models.Project `json:",inline"`
}

// ProjectWithProvidersResponse represents a project with its providers
type ProjectWithProvidersResponse struct {
	*models.ProjectWithProviders `json:",inline"`
}

// ProjectListResponse represents the response for listing projects
type ProjectListResponse struct {
	Projects    []*models.Project `json:"projects"`
	Total       int64             `json:"total"`
	Page        int               `json:"page"`
	PageSize    int               `json:"page_size"`
	TotalPages  int               `json:"total_pages"`
	HasNext     bool              `json:"has_next"`
	HasPrevious bool              `json:"has_previous"`
}

// ProjectProviderResponse represents the response for project-provider operations
type ProjectProviderResponse struct {
	*models.ProjectProvider `json:",inline"`
}

// ProjectProvidersListResponse represents the response for listing project providers
type ProjectProvidersListResponse struct {
	Providers []*models.ProjectProviderDetails `json:"providers"`
	Total     int                              `json:"total"`
}
