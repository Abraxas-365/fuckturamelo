package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/Abraxas-365/craftable/storex"
	"github.com/Abraxas-365/craftable/storex/providers/storexpostgres"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/Abraxas-365/fuckturamelo/projects"
	"github.com/Abraxas-365/fuckturamelo/projects/dto"
	"github.com/Abraxas-365/fuckturamelo/projects/models"
)

// projectRepository implements ProjectRepository using storex
type projectRepository struct {
	repo         *storexpostgres.PgRepository[models.Project]
	providerRepo *storexpostgres.PgRepository[models.ProjectProvider]
	db           *sqlx.DB
	search       *storexpostgres.PgSearchable[models.Project]
	bulk         *storexpostgres.PgBulkOperator[models.Project]
}

// NewProjectRepository creates a new project repository
func NewProjectRepository(db *sqlx.DB) ProjectRepository {
	repo := storexpostgres.NewPgRepository[models.Project](db, "projects", "id")
	providerRepo := storexpostgres.NewPgRepository[models.ProjectProvider](db, "project_providers", "id")
	search := storexpostgres.NewPgSearchable(repo)
	bulk := storexpostgres.NewPgBulkOperator(repo)

	return &projectRepository{
		repo:         repo,
		providerRepo: providerRepo,
		db:           db,
		search:       search,
		bulk:         bulk,
	}
}

// Create creates a new project
func (r *projectRepository) Create(ctx context.Context, project *models.Project) (*models.Project, error) {
	if project.ID == uuid.Nil {
		project.ID = uuid.New()
	}

	result, err := r.repo.Create(ctx, *project)
	if err != nil {
		if strings.Contains(err.Error(), "projects_name_org_unique") {
			return nil, projects.ProjectsErrors.New(projects.ErrProjectNameExists).
				WithDetail("name", project.Name).
				WithDetail("organization_id", project.OrganizationID.String()).
				WithCause(err)
		}
		return nil, projects.ProjectsErrors.New(projects.ErrProjectCreateFailed).
			WithDetail("project_name", project.Name).
			WithCause(err)
	}

	return &result, nil
}

// GetByID retrieves a project by ID
func (r *projectRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Project, error) {
	result, err := r.repo.FindByID(ctx, id.String())
	if err != nil {
		if storex.IsRecordNotFound(err) {
			return nil, projects.ProjectsErrors.New(projects.ErrProjectNotFound).
				WithDetail("project_id", id.String())
		}
		return nil, projects.ProjectsErrors.New(projects.ErrProjectNotFound).
			WithDetail("project_id", id.String()).
			WithCause(err)
	}

	return &result, nil
}

// GetByIDWithProviders retrieves a project with its providers
func (r *projectRepository) GetByIDWithProviders(ctx context.Context, id uuid.UUID) (*models.ProjectWithProviders, error) {
	project, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	providers, err := r.GetProviders(ctx, id)
	if err != nil {
		return nil, err
	}

	return &models.ProjectWithProviders{
		Project:   *project,
		Providers: &providers,
	}, nil
}

// Update updates an existing project
func (r *projectRepository) Update(ctx context.Context, id uuid.UUID, project *models.Project) (*models.Project, error) {
	project.ID = id
	result, err := r.repo.Update(ctx, id.String(), *project)
	if err != nil {
		if strings.Contains(err.Error(), "projects_name_org_unique") {
			return nil, projects.ProjectsErrors.New(projects.ErrProjectNameExists).
				WithDetail("name", project.Name).
				WithDetail("organization_id", project.OrganizationID.String()).
				WithCause(err)
		}
		if storex.IsRecordNotFound(err) {
			return nil, projects.ProjectsErrors.New(projects.ErrProjectNotFound).
				WithDetail("project_id", id.String())
		}
		return nil, projects.ProjectsErrors.New(projects.ErrProjectUpdateFailed).
			WithDetail("project_id", id.String()).
			WithCause(err)
	}

	return &result, nil
}

// Delete deletes a project
func (r *projectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.repo.Delete(ctx, id.String())
	if err != nil {
		if storex.IsRecordNotFound(err) {
			return projects.ProjectsErrors.New(projects.ErrProjectNotFound).
				WithDetail("project_id", id.String())
		}
		return projects.ProjectsErrors.New(projects.ErrProjectDeleteFailed).
			WithDetail("project_id", id.String()).
			WithCause(err)
	}

	return nil
}

// List retrieves projects with pagination and filtering
func (r *projectRepository) List(ctx context.Context, req *dto.ProjectListRequest) (*dto.ProjectListResponse, error) {
	opts := storex.PaginationOptions{
		Page:     req.Page,
		PageSize: req.PageSize,
		SortBy:   req.SortBy,
		SortDesc: req.SortOrder == "desc",
	}

	// Build filters
	filters := make(map[string]interface{})
	if req.OrganizationID != nil {
		filters["organization_id"] = *req.OrganizationID
	}
	if req.IsActive != nil {
		filters["is_active"] = *req.IsActive
	}
	if req.Search != nil && *req.Search != "" {
		// Use ILIKE for case-insensitive search on name and description
		query := fmt.Sprintf("%%%s%%", *req.Search)
		filters["$or"] = []map[string]interface{}{
			{"name": map[string]interface{}{"$ilike": query}},
			{"description": map[string]interface{}{"$ilike": query}},
		}
	}

	opts.Filters = filters

	result, err := r.repo.Paginate(ctx, opts)
	if err != nil {
		return nil, projects.ProjectsErrors.New(projects.ErrProjectListFailed).
			WithCause(err)
	}

	// Convert to response format
	projectPtrs := make([]*models.Project, len(result.Data))
	for i := range result.Data {
		projectPtrs[i] = &result.Data[i]
	}

	return &dto.ProjectListResponse{
		Projects:    projectPtrs,
		Total:       result.Total,
		Page:        result.Page,
		PageSize:    result.PageSize,
		TotalPages:  result.TotalPages,
		HasNext:     result.HasNext,
		HasPrevious: result.HasPrevious,
	}, nil
}

// GetByOrganization retrieves all projects for an organization
func (r *projectRepository) GetByOrganization(ctx context.Context, orgID uuid.UUID) ([]*models.Project, error) {
	filters := map[string]interface{}{
		"organization_id": orgID,
		"is_active":       true,
	}

	opts := storex.PaginationOptions{
		Page:     1,
		PageSize: 1000, // Large page size to get all active projects
		SortBy:   "name",
		Filters:  filters,
	}

	result, err := r.repo.Paginate(ctx, opts)
	if err != nil {
		return nil, projects.ProjectsErrors.New(projects.ErrProjectListFailed).
			WithDetail("organization_id", orgID.String()).
			WithCause(err)
	}

	// Convert to pointer slice
	projectPtrs := make([]*models.Project, len(result.Data))
	for i := range result.Data {
		projectPtrs[i] = &result.Data[i]
	}

	return projectPtrs, nil
}

// GetByNameAndOrganization retrieves a project by name within an organization
func (r *projectRepository) GetByNameAndOrganization(ctx context.Context, name string, orgID uuid.UUID) (*models.Project, error) {
	filters := map[string]interface{}{
		"name":            name,
		"organization_id": orgID,
	}

	result, err := r.repo.FindOne(ctx, filters)
	if err != nil {
		if storex.IsRecordNotFound(err) {
			return nil, projects.ProjectsErrors.New(projects.ErrProjectNotFound).
				WithDetail("name", name).
				WithDetail("organization_id", orgID.String())
		}
		return nil, projects.ProjectsErrors.New(projects.ErrProjectNotFound).
			WithDetail("name", name).
			WithDetail("organization_id", orgID.String()).
			WithCause(err)
	}

	return &result, nil
}

// Search performs full-text search on projects
func (r *projectRepository) Search(ctx context.Context, query string, orgID uuid.UUID) ([]*models.Project, error) {
	searchQuery := fmt.Sprintf("%%%s%%", query)

	sqlQuery := `
		SELECT * FROM projects 
		WHERE organization_id = $1 
		AND (name ILIKE $2 OR description ILIKE $2)
		AND is_active = true
		ORDER BY name
	`

	var projects []models.Project
	err := r.db.SelectContext(ctx, &projects, sqlQuery, orgID, searchQuery)
	if err != nil {
		return nil, projects.ProjectsErrors.New(projects.ErrProjectSearchFailed).
			WithDetail("query", query).
			WithDetail("organization_id", orgID.String()).
			WithCause(err)
	}

	// Convert to pointer slice
	projectPtrs := make([]*models.Project, len(projects))
	for i := range projects {
		projectPtrs[i] = &projects[i]
	}

	return projectPtrs, nil
}

// Provider relationship operations

// AddProvider adds a provider to a project
func (r *projectRepository) AddProvider(ctx context.Context, projectProvider *models.ProjectProvider) (*models.ProjectProvider, error) {
	if projectProvider.ID == uuid.Nil {
		projectProvider.ID = uuid.New()
	}

	result, err := r.providerRepo.Create(ctx, *projectProvider)
	if err != nil {
		if strings.Contains(err.Error(), "project_providers_unique") {
			return nil, projects.ProjectsErrors.New(projects.ErrProjectProviderExists).
				WithDetail("project_id", projectProvider.ProjectID.String()).
				WithDetail("provider_id", projectProvider.ProviderID.String()).
				WithCause(err)
		}
		return nil, projects.ProjectsErrors.New(projects.ErrProjectProviderManagementFailed).
			WithDetail("project_id", projectProvider.ProjectID.String()).
			WithDetail("provider_id", projectProvider.ProviderID.String()).
			WithCause(err)
	}

	return &result, nil
}

// RemoveProvider removes a provider from a project
func (r *projectRepository) RemoveProvider(ctx context.Context, projectID, providerID uuid.UUID) error {
	query := `DELETE FROM project_providers WHERE project_id = $1 AND provider_id = $2`
	result, err := r.db.ExecContext(ctx, query, projectID, providerID)
	if err != nil {
		return projects.ProjectsErrors.New(projects.ErrProjectProviderManagementFailed).
			WithDetail("project_id", projectID.String()).
			WithDetail("provider_id", providerID.String()).
			WithCause(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return projects.ProjectsErrors.New(projects.ErrProjectProviderManagementFailed).
			WithDetail("project_id", projectID.String()).
			WithDetail("provider_id", providerID.String()).
			WithCause(err)
	}

	if rowsAffected == 0 {
		return projects.ProjectsErrors.New(projects.ErrProjectProviderNotFound).
			WithDetail("project_id", projectID.String()).
			WithDetail("provider_id", providerID.String())
	}

	return nil
}

// UpdateProviderRole updates the role of a provider in a project
func (r *projectRepository) UpdateProviderRole(ctx context.Context, projectID, providerID uuid.UUID, role *string) (*models.ProjectProvider, error) {
	query := `
		UPDATE project_providers 
		SET role = $1, updated_at = NOW() 
		WHERE project_id = $2 AND provider_id = $3
		RETURNING *
	`

	var result models.ProjectProvider
	err := r.db.GetContext(ctx, &result, query, role, projectID, providerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, projects.ProjectsErrors.New(projects.ErrProjectProviderNotFound).
				WithDetail("project_id", projectID.String()).
				WithDetail("provider_id", providerID.String())
		}
		return nil, projects.ProjectsErrors.New(projects.ErrProjectProviderManagementFailed).
			WithDetail("project_id", projectID.String()).
			WithDetail("provider_id", providerID.String()).
			WithCause(err)
	}

	return &result, nil
}

// GetProviders retrieves all providers for a project
func (r *projectRepository) GetProviders(ctx context.Context, projectID uuid.UUID) ([]*models.ProjectProviderDetails, error) {
	query := `
		SELECT 
			pp.provider_id,
			p.name as provider_name,
			pp.role,
			pp.is_active,
			pp.created_at as joined_at
		FROM project_providers pp
		JOIN providers p ON pp.provider_id = p.id
		WHERE pp.project_id = $1
		ORDER BY pp.created_at
	`

	var providers []*models.ProjectProviderDetails
	err := r.db.SelectContext(ctx, &providers, query, projectID)
	if err != nil {
		return nil, projects.ProjectsErrors.New(projects.ErrProjectProviderManagementFailed).
			WithDetail("project_id", projectID.String()).
			WithCause(err)
	}

	return providers, nil
}

// GetProjectsByProvider retrieves all projects for a provider
func (r *projectRepository) GetProjectsByProvider(ctx context.Context, providerID uuid.UUID) ([]*models.Project, error) {
	query := `
		SELECT p.*
		FROM projects p
		JOIN project_providers pp ON p.id = pp.project_id
		WHERE pp.provider_id = $1 AND pp.is_active = true AND p.is_active = true
		ORDER BY p.name
	`

	var projects []*models.Project
	err := r.db.SelectContext(ctx, &projects, query, providerID)
	if err != nil {
		return nil, projects.ProjectsErrors.New(projects.ErrProjectListFailed).
			WithDetail("provider_id", providerID.String()).
			WithCause(err)
	}

	return projects, nil
}

// Bulk operations

// CreateBulk creates multiple projects
func (r *projectRepository) CreateBulk(ctx context.Context, projects []*models.Project) ([]*models.Project, error) {
	// Generate IDs for projects without them
	for _, project := range projects {
		if project.ID == uuid.Nil {
			project.ID = uuid.New()
		}
	}

	// Convert to value slice for bulk operation
	valueProjects := make([]models.Project, len(projects))
	for i, project := range projects {
		valueProjects[i] = *project
	}

	err := r.bulk.BulkInsert(ctx, valueProjects)
	if err != nil {
		return nil, projects.ProjectsErrors.New(projects.ErrProjectBulkCreateFailed).
			WithCause(err)
	}

	return projects, nil
}

// UpdateBulk updates multiple projects
func (r *projectRepository) UpdateBulk(ctx context.Context, projects []*models.Project) error {
	// Convert to value slice for bulk operation
	valueProjects := make([]models.Project, len(projects))
	for i, project := range projects {
		valueProjects[i] = *project
	}

	err := r.bulk.BulkUpdate(ctx, valueProjects)
	if err != nil {
		return projects.ProjectsErrors.New(projects.ErrProjectBulkUpdateFailed).
			WithCause(err)
	}

	return nil
}

// DeleteBulk deletes multiple projects
func (r *projectRepository) DeleteBulk(ctx context.Context, ids []uuid.UUID) error {
	// Convert UUIDs to strings
	stringIDs := make([]string, len(ids))
	for i, id := range ids {
		stringIDs[i] = id.String()
	}

	err := r.bulk.BulkDelete(ctx, stringIDs)
	if err != nil {
		return projects.ProjectsErrors.New(projects.ErrProjectBulkDeleteFailed).
			WithCause(err)
	}

	return nil
}

// AddProvidersBulk adds multiple providers to a project
func (r *projectRepository) AddProvidersBulk(ctx context.Context, projectID uuid.UUID, providerIDs []uuid.UUID) error {
	if len(providerIDs) == 0 {
		return nil
	}

	// Get organization ID for the project
	project, err := r.GetByID(ctx, projectID)
	if err != nil {
		return err
	}

	// Build bulk insert query
	valueStrings := make([]string, len(providerIDs))
	args := make([]interface{}, len(providerIDs)*4)

	for i, providerID := range providerIDs {
		valueStrings[i] = fmt.Sprintf("($%d, $%d, $%d, $%d)", i*4+1, i*4+2, i*4+3, i*4+4)
		args[i*4] = uuid.New()
		args[i*4+1] = projectID
		args[i*4+2] = providerID
		args[i*4+3] = project.OrganizationID
	}

	query := fmt.Sprintf(`
		INSERT INTO project_providers (id, project_id, provider_id, organization_id)
		VALUES %s
		ON CONFLICT (project_id, provider_id) DO NOTHING
	`, strings.Join(valueStrings, ","))

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return projects.ProjectsErrors.New(projects.ErrProjectProviderManagementFailed).
			WithDetail("project_id", projectID.String()).
			WithCause(err)
	}

	return nil
}

// RemoveProvidersBulk removes multiple providers from a project
func (r *projectRepository) RemoveProvidersBulk(ctx context.Context, projectID uuid.UUID, providerIDs []uuid.UUID) error {
	if len(providerIDs) == 0 {
		return nil
	}

	// Build placeholders for IN clause
	placeholders := make([]string, len(providerIDs))
	args := make([]interface{}, len(providerIDs)+1)
	args[0] = projectID

	for i, providerID := range providerIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = providerID
	}

	query := fmt.Sprintf(`
		DELETE FROM project_providers 
		WHERE project_id = $1 AND provider_id IN (%s)
	`, strings.Join(placeholders, ","))

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return projects.ProjectsErrors.New(projects.ErrProjectProviderManagementFailed).
			WithDetail("project_id", projectID.String()).
			WithCause(err)
	}

	return nil
}

// Utility operations

// ExistsByNameAndOrganization checks if a project exists by name and organization
func (r *projectRepository) ExistsByNameAndOrganization(ctx context.Context, name string, orgID uuid.UUID, excludeID *uuid.UUID) (bool, error) {
	query := `SELECT COUNT(*) FROM projects WHERE name = $1 AND organization_id = $2`
	args := []interface{}{name, orgID}

	if excludeID != nil {
		query += ` AND id != $3`
		args = append(args, *excludeID)
	}

	var count int
	err := r.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		return false, projects.ProjectsErrors.New(projects.ErrProjectValidationFailed).
			WithDetail("name", name).
			WithDetail("organization_id", orgID.String()).
			WithCause(err)
	}

	return count > 0, nil
}

// CountByOrganization counts projects in an organization
func (r *projectRepository) CountByOrganization(ctx context.Context, orgID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM projects WHERE organization_id = $1`

	var count int64
	err := r.db.GetContext(ctx, &count, query, orgID)
	if err != nil {
		return 0, projects.ProjectsErrors.New(projects.ErrProjectListFailed).
			WithDetail("organization_id", orgID.String()).
			WithCause(err)
	}

	return count, nil
}
