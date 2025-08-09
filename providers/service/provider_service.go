package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/Abraxas-365/fuckturamelo/providers"
	"github.com/Abraxas-365/fuckturamelo/providers/dto"
	"github.com/Abraxas-365/fuckturamelo/providers/models"
	postgres "github.com/Abraxas-365/fuckturamelo/providers/repository"
)

// ProviderService defines the interface for provider business logic
type ProviderService interface {
	// Basic CRUD operations
	CreateProvider(ctx context.Context, req *dto.CreateProviderRequest) (*dto.ProviderResponse, error)
	GetProvider(ctx context.Context, id uuid.UUID) (*dto.ProviderResponse, error)
	UpdateProvider(ctx context.Context, id uuid.UUID, req *dto.UpdateProviderRequest) (*dto.ProviderResponse, error)
	DeleteProvider(ctx context.Context, id uuid.UUID) error

	// Query operations
	ListProviders(ctx context.Context, req *dto.ProviderListRequest) (*dto.ProviderListResponse, error)
	GetProvidersByOrganization(ctx context.Context, orgID uuid.UUID) ([]dto.ProviderResponse, error)
	SearchProviders(ctx context.Context, query string, orgID uuid.UUID) ([]dto.ProviderResponse, error)

	// Business operations
	ActivateProvider(ctx context.Context, id uuid.UUID) (*dto.ProviderResponse, error)
	DeactivateProvider(ctx context.Context, id uuid.UUID) (*dto.ProviderResponse, error)
	DuplicateProvider(ctx context.Context, id uuid.UUID, newName string) (*dto.ProviderResponse, error)
}

// providerService implements ProviderService
type providerService struct {
	repo postgres.ProviderRepository
}

// NewProviderService creates a new provider service
func NewProviderService(repo postgres.ProviderRepository) ProviderService {
	return &providerService{
		repo: repo,
	}
}

// CreateProvider creates a new provider
func (s *providerService) CreateProvider(ctx context.Context, req *dto.CreateProviderRequest) (*dto.ProviderResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, providers.ProvidersErrors.New(providers.ErrProviderValidationFailed).
			WithCause(err)
	}

	// Check if provider with same name exists in organization
	existing, err := s.repo.GetByNameAndOrganization(ctx, req.Name, req.OrganizationID)
	if err != nil && !providers.IsProviderNotFound(err) {
		return nil, err
	}
	if existing != nil {
		return nil, providers.ProvidersErrors.New(providers.ErrProviderNameExists).
			WithDetail("name", req.Name).
			WithDetail("organization_id", req.OrganizationID.String())
	}

	// Create provider model
	provider := &models.Provider{
		ID:             uuid.New(),
		UserID:         req.UserID,
		OrganizationID: req.OrganizationID,
		Name:           req.Name,
		ProviderCode:   req.ProviderCode,
		IsActive:       true, // Default to active
		Metadata:       req.Metadata,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Override IsActive if explicitly set
	if req.IsActive != nil {
		provider.IsActive = *req.IsActive
	}

	// Initialize metadata if nil
	if provider.Metadata == nil {
		provider.Metadata = make(map[string]interface{})
	}

	// Create provider
	created, err := s.repo.Create(ctx, provider)
	if err != nil {
		return nil, err
	}

	return s.modelToResponse(created), nil
}

// GetProvider retrieves a provider by ID
func (s *providerService) GetProvider(ctx context.Context, id uuid.UUID) (*dto.ProviderResponse, error) {
	provider, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.modelToResponse(provider), nil
}

// UpdateProvider updates an existing provider
func (s *providerService) UpdateProvider(ctx context.Context, id uuid.UUID, req *dto.UpdateProviderRequest) (*dto.ProviderResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, providers.ProvidersErrors.New(providers.ErrProviderValidationFailed).
			WithCause(err)
	}

	// Get existing provider
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if name is being changed and conflicts
	if req.Name != nil && *req.Name != existing.Name {
		nameConflict, err := s.repo.GetByNameAndOrganization(ctx, *req.Name, existing.OrganizationID)
		if err != nil && !providers.IsProviderNotFound(err) {
			return nil, err
		}
		if nameConflict != nil {
			return nil, providers.ProvidersErrors.New(providers.ErrProviderNameExists).
				WithDetail("name", *req.Name).
				WithDetail("organization_id", existing.OrganizationID.String())
		}
	}

	// Apply updates
	updated := *existing

	if req.UserID != nil {
		updated.UserID = req.UserID
	}
	if req.Name != nil {
		updated.Name = *req.Name
	}
	if req.ProviderCode != nil {
		updated.ProviderCode = req.ProviderCode
	}
	if req.IsActive != nil {
		updated.IsActive = *req.IsActive
	}
	if req.Metadata != nil {
		updated.Metadata = req.Metadata
	}

	updated.UpdatedAt = time.Now()

	// Update provider
	result, err := s.repo.Update(ctx, id, &updated)
	if err != nil {
		return nil, err
	}

	return s.modelToResponse(result), nil
}

// DeleteProvider deletes a provider
func (s *providerService) DeleteProvider(ctx context.Context, id uuid.UUID) error {
	// Check if provider exists
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// TODO: Add business logic to check if provider can be deleted
	// e.g., check if provider is referenced by any invoices

	return s.repo.Delete(ctx, id)
}

// ListProviders lists providers with filtering and pagination
func (s *providerService) ListProviders(ctx context.Context, req *dto.ProviderListRequest) (*dto.ProviderListResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, providers.ProvidersErrors.New(providers.ErrProviderValidationFailed).
			WithCause(err)
	}

	return s.repo.List(ctx, req)
}

// GetProvidersByOrganization retrieves all providers for an organization
func (s *providerService) GetProvidersByOrganization(ctx context.Context, orgID uuid.UUID) ([]dto.ProviderResponse, error) {
	providersData, err := s.repo.GetByOrganization(ctx, orgID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.ProviderResponse, len(providersData))
	for i, p := range providersData {
		responses[i] = *s.modelToResponse(p)
	}

	return responses, nil
}

// SearchProviders performs text search on providers
func (s *providerService) SearchProviders(ctx context.Context, query string, orgID uuid.UUID) ([]dto.ProviderResponse, error) {
	if query == "" {
		return []dto.ProviderResponse{}, nil
	}

	providersData, err := s.repo.Search(ctx, query, orgID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.ProviderResponse, len(providersData))
	for i, p := range providersData {
		responses[i] = *s.modelToResponse(p)
	}

	return responses, nil
}

// ActivateProvider activates a provider
func (s *providerService) ActivateProvider(ctx context.Context, id uuid.UUID) (*dto.ProviderResponse, error) {
	req := &dto.UpdateProviderRequest{
		IsActive: &[]bool{true}[0],
	}

	return s.UpdateProvider(ctx, id, req)
}

// DeactivateProvider deactivates a provider
func (s *providerService) DeactivateProvider(ctx context.Context, id uuid.UUID) (*dto.ProviderResponse, error) {
	req := &dto.UpdateProviderRequest{
		IsActive: &[]bool{false}[0],
	}

	return s.UpdateProvider(ctx, id, req)
}

// DuplicateProvider creates a copy of an existing provider with a new name
func (s *providerService) DuplicateProvider(ctx context.Context, id uuid.UUID, newName string) (*dto.ProviderResponse, error) {
	// Get original provider
	original, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Create duplicate request
	req := &dto.CreateProviderRequest{
		UserID:         original.UserID,
		OrganizationID: original.OrganizationID,
		Name:           newName,
		ProviderCode:   original.ProviderCode,
		IsActive:       &original.IsActive,
		Metadata:       original.Metadata,
	}

	return s.CreateProvider(ctx, req)
}

// Helper methods

func (s *providerService) modelToResponse(provider *models.Provider) *dto.ProviderResponse {
	return &dto.ProviderResponse{
		ID:             provider.ID,
		UserID:         provider.UserID,
		OrganizationID: provider.OrganizationID,
		Name:           provider.Name,
		ProviderCode:   provider.ProviderCode,
		IsActive:       provider.IsActive,
		Metadata:       provider.Metadata,
		CreatedAt:      provider.CreatedAt,
		UpdatedAt:      provider.UpdatedAt,
	}
}
