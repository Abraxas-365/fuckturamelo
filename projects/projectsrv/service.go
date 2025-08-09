package projectsrv

import (
	"context"
	"time"

	"github.com/Abraxas-365/fuckturamelo/projects"
	"github.com/Abraxas-365/fuckturamelo/projects/dto"
	"github.com/Abraxas-365/fuckturamelo/projects/models"
	postgres "github.com/Abraxas-365/fuckturamelo/projects/repository"
	"github.com/google/uuid"
)

// ProjectService defines the interface for project business logic
type ProjectService interface {
	// Basic CRUD operations
	CreateProject(ctx context.Context, req *dto.CreateProjectRequest) (*dto.ProjectResponse, error)
	GetProject(ctx context.Context, id uuid.UUID) (*dto.ProjectResponse, error)
	GetProjectWithProviders(ctx context.Context, id uuid.UUID) (*dto.ProjectWithProvidersResponse, error)
	UpdateProject(ctx context.Context, id uuid.UUID, req *dto.UpdateProjectRequest) (*dto.ProjectResponse, error)
	DeleteProject(ctx context.Context, id uuid.UUID) error

	// Query operations
	ListProjects(ctx context.Context, req *dto.ProjectListRequest) (*dto.ProjectListResponse, error)
	GetProjectsByOrganization(ctx context.Context, orgID uuid.UUID) ([]*dto.ProjectResponse, error)
	SearchProjects(ctx context.Context, query string, orgID uuid.UUID) ([]*dto.ProjectResponse, error)

	// Project activation/deactivation
	ActivateProject(ctx context.Context, id uuid.UUID) (*dto.ProjectResponse, error)
	DeactivateProject(ctx context.Context, id uuid.UUID) (*dto.ProjectResponse, error)
	DuplicateProject(ctx context.Context, id uuid.UUID, newName string) (*dto.ProjectResponse, error)

	// Provider management
	AddProvider(ctx context.Context, projectID uuid.UUID, req *dto.AddProviderRequest) (*dto.ProjectProviderResponse, error)
	RemoveProvider(ctx context.Context, projectID, providerID uuid.UUID) error
	UpdateProviderRole(ctx context.Context, projectID, providerID uuid.UUID, req *dto.UpdateProviderRoleRequest) (*dto.ProjectProviderResponse, error)
	GetProjectProviders(ctx context.Context, projectID uuid.UUID) (*dto.ProjectProvidersListResponse, error)
	AddProvidersBulk(ctx context.Context, projectID uuid.UUID, req *dto.BulkProviderRequest) error
	RemoveProvidersBulk(ctx context.Context, projectID uuid.UUID, req *dto.BulkProviderRequest) error

	// Utility operations
	GetProjectStats(ctx context.Context, orgID uuid.UUID) (*ProjectStats, error)
}

// ProjectStats represents project statistics
type ProjectStats struct {
	TotalProjects    int64 `json:"total_projects"`
	ActiveProjects   int64 `json:"active_projects"`
	InactiveProjects int64 `json:"inactive_projects"`
}

// projectService implements ProjectService
type projectService struct {
	repo postgres.ProjectRepository
}

// NewProjectService creates a new project service
func NewProjectService(repo postgres.ProjectRepository) ProjectService {
	return &projectService{
		repo: repo,
	}
}

// CreateProject creates a new project
func (s *projectService) CreateProject(ctx context.Context, req *dto.CreateProjectRequest) (*dto.ProjectResponse, error) {
	// Validate request
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// Check if project name already exists in organization
	exists, err := s.repo.ExistsByNameAndOrganization(ctx, req.Name, req.OrganizationID, nil)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, projects.ProjectsErrors.New(projects.ErrProjectNameExists).
			WithDetail("name", req.Name).
			WithDetail("organization_id", req.OrganizationID.String())
	}

	// Create project model
	project := &models.Project{
		OrganizationID: req.OrganizationID,
		Name:           req.Name,
		Description:    req.Description,
		IsActive:       true,
		Metadata:       req.Metadata,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if project.Metadata == nil {
		project.Metadata = make(models.Metadata)
	}

	// Create project
	createdProject, err := s.repo.Create(ctx, project)
	if err != nil {
		return nil, err
	}

	// Add providers if specified
	if len(req.ProviderIDs) > 0 {
		err = s.repo.AddProvidersBulk(ctx, createdProject.ID, req.ProviderIDs)
		if err != nil {
			// Log error but don't fail the project creation
			// You might want to implement proper logging here
		}
	}

	return &dto.ProjectResponse{Project: createdProject}, nil
}

// GetProject retrieves a project by ID
func (s *projectService) GetProject(ctx context.Context, id uuid.UUID) (*dto.ProjectResponse, error) {
	project, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &dto.ProjectResponse{Project: project}, nil
}

// GetProjectWithProviders retrieves a project with its providers
func (s *projectService) GetProjectWithProviders(ctx context.Context, id uuid.UUID) (*dto.ProjectWithProvidersResponse, error) {
	project, err := s.repo.GetByIDWithProviders(ctx, id)
	if err != nil {
		return nil, err
	}

	return &dto.ProjectWithProvidersResponse{ProjectWithProviders: project}, nil
}

// UpdateProject updates an existing project
func (s *projectService) UpdateProject(ctx context.Context, id uuid.UUID, req *dto.UpdateProjectRequest) (*dto.ProjectResponse, error) {
	// Get existing project
	existingProject, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check name uniqueness if name is being changed
	if req.Name != nil && *req.Name != existingProject.Name {
		exists, err := s.repo.ExistsByNameAndOrganization(ctx, *req.Name, existingProject.OrganizationID, &id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, projects.ProjectsErrors.New(projects.ErrProjectNameExists).
				WithDetail("name", *req.Name).
				WithDetail("organization_id", existingProject.OrganizationID.String())
		}
	}

	// Apply updates
	updatedProject := *existingProject
	if req.Name != nil {
		updatedProject.Name = *req.Name
	}
	if req.Description != nil {
		updatedProject.Description = req.Description
	}
	if req.IsActive != nil {
		updatedProject.IsActive = *req.IsActive
	}
	if req.Metadata != nil {
		updatedProject.Metadata = req.Metadata
	}
	updatedProject.UpdatedAt = time.Now()

	// Update project
	result, err := s.repo.Update(ctx, id, &updatedProject)
	if err != nil {
		return nil, err
	}

	return &dto.ProjectResponse{Project: result}, nil
}

// DeleteProject deletes a project
func (s *projectService) DeleteProject(ctx context.Context, id uuid.UUID) error {
	// Check if project exists
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// TODO: Check for dependencies (invoices, etc.)
	// This is where you would add business logic to prevent deletion
	// if the project has active invoices or other dependencies

	return s.repo.Delete(ctx, id)
}

// ListProjects lists projects with pagination and filtering
func (s *projectService) ListProjects(ctx context.Context, req *dto.ProjectListRequest) (*dto.ProjectListResponse, error) {
	// Set defaults
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}

	return s.repo.List(ctx, req)
}

// GetProjectsByOrganization gets all projects for an organization
func (s *projectService) GetProjectsByOrganization(ctx context.Context, orgID uuid.UUID) ([]*dto.ProjectResponse, error) {
	projects, err := s.repo.GetByOrganization(ctx, orgID)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.ProjectResponse, len(projects))
	for i, project := range projects {
		responses[i] = &dto.ProjectResponse{Project: project}
	}

	return responses, nil
}

// SearchProjects searches projects
func (s *projectService) SearchProjects(ctx context.Context, query string, orgID uuid.UUID) ([]*dto.ProjectResponse, error) {
	projects, err := s.repo.Search(ctx, query, orgID)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.ProjectResponse, len(projects))
	for i, project := range projects {
		responses[i] = &dto.ProjectResponse{Project: project}
	}

	return responses, nil
}

// ActivateProject activates a project
func (s *projectService) ActivateProject(ctx context.Context, id uuid.UUID) (*dto.ProjectResponse, error) {
	req := &dto.UpdateProjectRequest{
		IsActive: &[]bool{true}[0],
	}
	return s.UpdateProject(ctx, id, req)
}

// DeactivateProject deactivates a project
func (s *projectService) DeactivateProject(ctx context.Context, id uuid.UUID) (*dto.ProjectResponse, error) {
	req := &dto.UpdateProjectRequest{
		IsActive: &[]bool{false}[0],
	}
	return s.UpdateProject(ctx, id, req)
}

// DuplicateProject duplicates an existing project
func (s *projectService) DuplicateProject(ctx context.Context, id uuid.UUID, newName string) (*dto.ProjectResponse, error) {
	// Get original project
	original, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Create duplicate request
	req := &dto.CreateProjectRequest{
		OrganizationID: original.OrganizationID,
		Name:           newName,
		Description:    original.Description,
		Metadata:       original.Metadata,
	}

	// Get providers from original project
	providers, err := s.repo.GetProviders(ctx, id)
	if err == nil && len(providers) > 0 {
		req.ProviderIDs = make([]uuid.UUID, len(providers))
		for i, provider := range providers {
			req.ProviderIDs[i] = provider.ProviderID
		}
	}

	return s.CreateProject(ctx, req)
}

// Provider management operations

// AddProvider adds a provider to a project
func (s *projectService) AddProvider(ctx context.Context, projectID uuid.UUID, req *dto.AddProviderRequest) (*dto.ProjectProviderResponse, error) {
	// Verify project exists
	project, err := s.repo.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Create project provider relationship
	projectProvider := &models.ProjectProvider{
		ProjectID:      projectID,
		ProviderID:     req.ProviderID,
		OrganizationID: project.OrganizationID,
		Role:           req.Role,
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	result, err := s.repo.AddProvider(ctx, projectProvider)
	if err != nil {
		return nil, err
	}

	return &dto.ProjectProviderResponse{ProjectProvider: result}, nil
}

// RemoveProvider removes a provider from a project
func (s *projectService) RemoveProvider(ctx context.Context, projectID, providerID uuid.UUID) error {
	return s.repo.RemoveProvider(ctx, projectID, providerID)
}

// UpdateProviderRole updates a provider's role in a project
func (s *projectService) UpdateProviderRole(ctx context.Context, projectID, providerID uuid.UUID, req *dto.UpdateProviderRoleRequest) (*dto.ProjectProviderResponse, error) {
	result, err := s.repo.UpdateProviderRole(ctx, projectID, providerID, req.Role)
	if err != nil {
		return nil, err
	}

	return &dto.ProjectProviderResponse{ProjectProvider: result}, nil
}

// GetProjectProviders gets all providers for a project
func (s *projectService) GetProjectProviders(ctx context.Context, projectID uuid.UUID) (*dto.ProjectProvidersListResponse, error) {
	providers, err := s.repo.GetProviders(ctx, projectID)
	if err != nil {
		return nil, err
	}

	return &dto.ProjectProvidersListResponse{
		Providers: providers,
		Total:     len(providers),
	}, nil
}

// AddProvidersBulk adds multiple providers to a project
func (s *projectService) AddProvidersBulk(ctx context.Context, projectID uuid.UUID, req *dto.BulkProviderRequest) error {
	// Verify project exists
	_, err := s.repo.GetByID(ctx, projectID)
	if err != nil {
		return err
	}

	return s.repo.AddProvidersBulk(ctx, projectID, req.ProviderIDs)
}

// RemoveProvidersBulk removes multiple providers from a project
func (s *projectService) RemoveProvidersBulk(ctx context.Context, projectID uuid.UUID, req *dto.BulkProviderRequest) error {
	return s.repo.RemoveProvidersBulk(ctx, projectID, req.ProviderIDs)
}

// GetProjectStats gets project statistics for an organization
func (s *projectService) GetProjectStats(ctx context.Context, orgID uuid.UUID) (*ProjectStats, error) {
	// Get total count
	total, err := s.repo.CountByOrganization(ctx, orgID)
	if err != nil {
		return nil, err
	}

	// For now, implement with separate queries
	// In production, you might want to optimize this with a single query
	req := &dto.ProjectListRequest{
		OrganizationID: &orgID,
		IsActive:       &[]bool{true}[0],
		Page:           1,
		PageSize:       1,
	}

	activeResult, err := s.repo.List(ctx, req)
	if err != nil {
		return nil, err
	}

	active := activeResult.Total
	inactive := total - active

	return &ProjectStats{
		TotalProjects:    total,
		ActiveProjects:   active,
		InactiveProjects: inactive,
	}, nil
}

// Validation helpers

func (s *projectService) validateCreateRequest(req *dto.CreateProjectRequest) error {
	if req.OrganizationID == uuid.Nil {
		return projects.ProjectsErrors.New(projects.ErrProjectValidationFailed).
			WithDetail("field", "organization_id").
			WithDetail("reason", "required")
	}

	if req.Name == "" {
		return projects.ProjectsErrors.New(projects.ErrProjectValidationFailed).
			WithDetail("field", "name").
			WithDetail("reason", "required")
	}

	if len(req.Name) > 255 {
		return projects.ProjectsErrors.New(projects.ErrProjectValidationFailed).
			WithDetail("field", "name").
			WithDetail("reason", "too_long").
			WithDetail("max_length", "255")
	}

	if req.Description != nil && len(*req.Description) > 1000 {
		return projects.ProjectsErrors.New(projects.ErrProjectValidationFailed).
			WithDetail("field", "description").
			WithDetail("reason", "too_long").
			WithDetail("max_length", "1000")
	}

	return nil
}
