package postgres

import (
	"context"

	"github.com/Abraxas-365/fuckturamelo/projects/dto"
	"github.com/Abraxas-365/fuckturamelo/projects/models"
	"github.com/google/uuid"
)

// ProjectRepository defines the interface for project repository operations
type ProjectRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, project *models.Project) (*models.Project, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Project, error)
	GetByIDWithProviders(ctx context.Context, id uuid.UUID) (*models.ProjectWithProviders, error)
	Update(ctx context.Context, id uuid.UUID, project *models.Project) (*models.Project, error)
	Delete(ctx context.Context, id uuid.UUID) error

	// Query operations
	List(ctx context.Context, req *dto.ProjectListRequest) (*dto.ProjectListResponse, error)
	GetByOrganization(ctx context.Context, orgID uuid.UUID) ([]*models.Project, error)
	GetByNameAndOrganization(ctx context.Context, name string, orgID uuid.UUID) (*models.Project, error)
	Search(ctx context.Context, query string, orgID uuid.UUID) ([]*models.Project, error)

	// Provider relationship operations
	AddProvider(ctx context.Context, projectProvider *models.ProjectProvider) (*models.ProjectProvider, error)
	RemoveProvider(ctx context.Context, projectID, providerID uuid.UUID) error
	UpdateProviderRole(ctx context.Context, projectID, providerID uuid.UUID, role *string) (*models.ProjectProvider, error)
	GetProviders(ctx context.Context, projectID uuid.UUID) ([]*models.ProjectProviderDetails, error)
	GetProjectsByProvider(ctx context.Context, providerID uuid.UUID) ([]*models.Project, error)

	// Bulk operations
	CreateBulk(ctx context.Context, projects []*models.Project) ([]*models.Project, error)
	UpdateBulk(ctx context.Context, projects []*models.Project) error
	DeleteBulk(ctx context.Context, ids []uuid.UUID) error
	AddProvidersBulk(ctx context.Context, projectID uuid.UUID, providerIDs []uuid.UUID) error
	RemoveProvidersBulk(ctx context.Context, projectID uuid.UUID, providerIDs []uuid.UUID) error

	// Utility operations
	ExistsByNameAndOrganization(ctx context.Context, name string, orgID uuid.UUID, excludeID *uuid.UUID) (bool, error)
	CountByOrganization(ctx context.Context, orgID uuid.UUID) (int64, error)
}
