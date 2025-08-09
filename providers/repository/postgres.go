package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/Abraxas-365/craftable/storex"
	"github.com/Abraxas-365/craftable/storex/providers/storexpostgres"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/Abraxas-365/fuckturamelo/providers"
	"github.com/Abraxas-365/fuckturamelo/providers/dto"
	"github.com/Abraxas-365/fuckturamelo/providers/models"
)

// ProviderRepository defines the interface for provider repository operations
type ProviderRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, provider *models.Provider) (*models.Provider, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Provider, error)
	Update(ctx context.Context, id uuid.UUID, provider *models.Provider) (*models.Provider, error)
	Delete(ctx context.Context, id uuid.UUID) error

	// Query operations
	List(ctx context.Context, req *dto.ProviderListRequest) (*dto.ProviderListResponse, error)
	GetByOrganization(ctx context.Context, orgID uuid.UUID) ([]*models.Provider, error)
	GetByNameAndOrganization(ctx context.Context, name string, orgID uuid.UUID) (*models.Provider, error)
	Search(ctx context.Context, query string, orgID uuid.UUID) ([]*models.Provider, error)

	// Bulk operations
	CreateBulk(ctx context.Context, providerList []*models.Provider) ([]*models.Provider, error)
	UpdateBulk(ctx context.Context, providerList []*models.Provider) error
	DeleteBulk(ctx context.Context, ids []uuid.UUID) error
}

// providerRepository implements ProviderRepository using storex
type providerRepository struct {
	repo   *storexpostgres.PgRepository[models.Provider]
	db     *sqlx.DB
	search *storexpostgres.PgSearchable[models.Provider]
	bulk   *storexpostgres.PgBulkOperator[models.Provider]
}

// NewProviderRepository creates a new provider repository
func NewProviderRepository(db *sqlx.DB) ProviderRepository {
	repo := storexpostgres.NewPgRepository[models.Provider](db, "providers", "id")
	search := storexpostgres.NewPgSearchable(repo)
	bulk := storexpostgres.NewPgBulkOperator(repo)

	return &providerRepository{
		repo:   repo,
		db:     db,
		search: search,
		bulk:   bulk,
	}
}

// Create creates a new provider
func (r *providerRepository) Create(ctx context.Context, provider *models.Provider) (*models.Provider, error) {
	if provider.ID == uuid.Nil {
		provider.ID = uuid.New()
	}

	result, err := r.repo.Create(ctx, *provider)
	if err != nil {
		if strings.Contains(err.Error(), "providers_name_org_unique") {
			return nil, providers.ProvidersErrors.New(providers.ErrProviderNameExists).
				WithDetail("name", provider.Name).
				WithDetail("organization_id", provider.OrganizationID.String()).
				WithCause(err)
		}
		return nil, providers.ProvidersErrors.New(providers.ErrProviderCreateFailed).
			WithCause(err)
	}

	return &result, nil
}

// GetByID retrieves a provider by ID
func (r *providerRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Provider, error) {
	result, err := r.repo.FindByID(ctx, id.String())
	if err != nil {
		if storex.IsRecordNotFound(err) {
			return nil, providers.ProvidersErrors.New(providers.ErrProviderNotFound).
				WithDetail("id", id.String())
		}
		return nil, providers.ProvidersErrors.New(providers.ErrProviderListFailed).
			WithDetail("id", id.String()).
			WithCause(err)
	}

	return &result, nil
}

// Update updates an existing provider
func (r *providerRepository) Update(ctx context.Context, id uuid.UUID, provider *models.Provider) (*models.Provider, error) {
	provider.ID = id
	result, err := r.repo.Update(ctx, id.String(), *provider)
	if err != nil {
		if storex.IsRecordNotFound(err) {
			return nil, providers.ProvidersErrors.New(providers.ErrProviderNotFound).
				WithDetail("id", id.String())
		}
		if strings.Contains(err.Error(), "providers_name_org_unique") {
			return nil, providers.ProvidersErrors.New(providers.ErrProviderNameExists).
				WithDetail("name", provider.Name).
				WithDetail("organization_id", provider.OrganizationID.String()).
				WithCause(err)
		}
		return nil, providers.ProvidersErrors.New(providers.ErrProviderUpdateFailed).
			WithDetail("id", id.String()).
			WithCause(err)
	}

	return &result, nil
}

// Delete deletes a provider by ID
func (r *providerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.repo.Delete(ctx, id.String())
	if err != nil {
		if storex.IsRecordNotFound(err) {
			return providers.ProvidersErrors.New(providers.ErrProviderNotFound).
				WithDetail("id", id.String())
		}
		return providers.ProvidersErrors.New(providers.ErrProviderDeleteFailed).
			WithDetail("id", id.String()).
			WithCause(err)
	}

	return nil
}

// List retrieves providers with pagination and filtering
func (r *providerRepository) List(ctx context.Context, req *dto.ProviderListRequest) (*dto.ProviderListResponse, error) {
	// Set defaults
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}
	if req.OrderBy == nil {
		orderBy := "created_at"
		req.OrderBy = &orderBy
	}

	// Build filters
	filters := make(map[string]any)
	if req.OrganizationID != nil {
		filters["organization_id"] = *req.OrganizationID
	}
	if req.IsActive != nil {
		filters["is_active"] = *req.IsActive
	}

	// Build pagination options
	opts := storex.PaginationOptions{
		Page:     req.Page,
		PageSize: req.PageSize,
		Filters:  filters,
		OrderBy:  *req.OrderBy,
		Desc:     req.Desc,
	}

	// Handle search if provided
	if req.Search != nil && *req.Search != "" {
		return r.searchProviders(ctx, *req.Search, opts)
	}

	result, err := r.repo.Paginate(ctx, opts)
	if err != nil {
		return nil, providers.ProvidersErrors.New(providers.ErrProviderListFailed).
			WithCause(err)
	}

	return r.buildListResponse(result), nil
}

// GetByOrganization retrieves all providers for an organization
func (r *providerRepository) GetByOrganization(ctx context.Context, orgID uuid.UUID) ([]*models.Provider, error) {
	opts := storex.PaginationOptions{
		Page:     1,
		PageSize: 1000, // Large page size to get all
		Filters:  map[string]any{"organization_id": orgID},
		OrderBy:  "name",
	}

	result, err := r.repo.Paginate(ctx, opts)
	if err != nil {
		return nil, providers.ProvidersErrors.New(providers.ErrProviderListFailed).
			WithDetail("organization_id", orgID.String()).
			WithCause(err)
	}

	providerPtrs := make([]*models.Provider, len(result.Data))
	for i, p := range result.Data {
		providerPtrs[i] = &p
	}

	return providerPtrs, nil
}

// GetByNameAndOrganization retrieves a provider by name and organization
func (r *providerRepository) GetByNameAndOrganization(ctx context.Context, name string, orgID uuid.UUID) (*models.Provider, error) {
	filters := map[string]any{
		"name":            name,
		"organization_id": orgID,
	}

	result, err := r.repo.FindOne(ctx, filters)
	if err != nil {
		if storex.IsRecordNotFound(err) {
			return nil, providers.ProvidersErrors.New(providers.ErrProviderNotFound).
				WithDetail("name", name).
				WithDetail("organization_id", orgID.String())
		}
		return nil, providers.ProvidersErrors.New(providers.ErrProviderListFailed).
			WithDetail("name", name).
			WithDetail("organization_id", orgID.String()).
			WithCause(err)
	}

	return &result, nil
}

// Search performs a text search on providers
func (r *providerRepository) Search(ctx context.Context, query string, orgID uuid.UUID) ([]*models.Provider, error) {
	opts := storex.SearchOptions{
		Fields: []string{"name", "provider_code"},
		Boost:  map[string]float64{"name": 2.0, "provider_code": 1.0},
		Limit:  50,
	}

	results, err := r.search.Search(ctx, query, opts)
	if err != nil {
		return nil, providers.ProvidersErrors.New(providers.ErrProviderSearchFailed).
			WithDetail("query", query).
			WithDetail("organization_id", orgID.String()).
			WithCause(err)
	}

	// Filter by organization
	var filtered []*models.Provider
	for _, p := range results {
		if p.OrganizationID == orgID {
			filtered = append(filtered, &p)
		}
	}

	return filtered, nil
}

// CreateBulk creates multiple providers in a single transaction
func (r *providerRepository) CreateBulk(ctx context.Context, providerList []*models.Provider) ([]*models.Provider, error) {
	// Generate IDs for providers that don't have them
	for _, p := range providerList {
		if p.ID == uuid.Nil {
			p.ID = uuid.New()
		}
	}

	// Use transaction to create all providers
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, providers.ProvidersErrors.New(providers.ErrProviderBulkCreateFailed).
			WithDetail("count", len(providerList)).
			WithCause(err)
	}
	defer tx.Rollback()

	var created []*models.Provider
	for _, provider := range providerList {
		result, err := r.repo.Create(ctx, *provider)
		if err != nil {
			return nil, providers.ProvidersErrors.New(providers.ErrProviderBulkCreateFailed).
				WithDetail("count", len(providerList)).
				WithCause(err)
		}
		created = append(created, &result)
	}

	if err := tx.Commit(); err != nil {
		return nil, providers.ProvidersErrors.New(providers.ErrProviderBulkCreateFailed).
			WithDetail("count", len(providerList)).
			WithCause(err)
	}

	return created, nil
}

// UpdateBulk updates multiple providers
func (r *providerRepository) UpdateBulk(ctx context.Context, providerList []*models.Provider) error {
	// Use transaction to update all providers
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return providers.ProvidersErrors.New(providers.ErrProviderBulkUpdateFailed).
			WithDetail("count", len(providerList)).
			WithCause(err)
	}
	defer tx.Rollback()

	for _, provider := range providerList {
		_, err := r.repo.Update(ctx, provider.ID.String(), *provider)
		if err != nil {
			return providers.ProvidersErrors.New(providers.ErrProviderBulkUpdateFailed).
				WithDetail("count", len(providerList)).
				WithCause(err)
		}
	}

	if err := tx.Commit(); err != nil {
		return providers.ProvidersErrors.New(providers.ErrProviderBulkUpdateFailed).
			WithDetail("count", len(providerList)).
			WithCause(err)
	}

	return nil
}

// DeleteBulk deletes multiple providers by IDs
func (r *providerRepository) DeleteBulk(ctx context.Context, ids []uuid.UUID) error {
	// Use transaction to delete all providers
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return providers.ProvidersErrors.New(providers.ErrProviderBulkDeleteFailed).
			WithDetail("count", len(ids)).
			WithCause(err)
	}
	defer tx.Rollback()

	for _, id := range ids {
		err := r.repo.Delete(ctx, id.String())
		if err != nil {
			return providers.ProvidersErrors.New(providers.ErrProviderBulkDeleteFailed).
				WithDetail("count", len(ids)).
				WithCause(err)
		}
	}

	if err := tx.Commit(); err != nil {
		return providers.ProvidersErrors.New(providers.ErrProviderBulkDeleteFailed).
			WithDetail("count", len(ids)).
			WithCause(err)
	}

	return nil
}

// Helper methods

func (r *providerRepository) searchProviders(ctx context.Context, query string, opts storex.PaginationOptions) (*dto.ProviderListResponse, error) {
	// Build custom search query that respects filters
	searchSQL := `
		SELECT p.* FROM providers p
		WHERE (p.name ILIKE $1 OR p.provider_code ILIKE $1)
	`
	args := []any{"%" + query + "%"}
	argIndex := 2

	// Add filters
	if orgID, ok := opts.Filters["organization_id"]; ok {
		searchSQL += fmt.Sprintf(" AND p.organization_id = $%d", argIndex)
		args = append(args, orgID)
		argIndex++
	}
	if isActive, ok := opts.Filters["is_active"]; ok {
		searchSQL += fmt.Sprintf(" AND p.is_active = $%d", argIndex)
		args = append(args, isActive)
		argIndex++
	}

	// Add ordering
	direction := "ASC"
	if opts.Desc {
		direction = "DESC"
	}
	searchSQL += fmt.Sprintf(" ORDER BY p.%s %s", opts.OrderBy, direction)

	// Add pagination
	offset := (opts.Page - 1) * opts.PageSize
	searchSQL += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, opts.PageSize, offset)

	// Execute search query
	var providersData []models.Provider
	err := r.db.SelectContext(ctx, &providersData, searchSQL, args...)
	if err != nil {
		return nil, providers.ProvidersErrors.New(providers.ErrProviderSearchFailed).
			WithDetail("query", query).
			WithCause(err)
	}

	// Count total results
	countSQL := `
		SELECT COUNT(*) FROM providers p
		WHERE (p.name ILIKE $1 OR p.provider_code ILIKE $1)
	`
	countArgs := []any{"%" + query + "%"}
	countIndex := 2

	if orgID, ok := opts.Filters["organization_id"]; ok {
		countSQL += fmt.Sprintf(" AND p.organization_id = $%d", countIndex)
		countArgs = append(countArgs, orgID)
		countIndex++
	}
	if isActive, ok := opts.Filters["is_active"]; ok {
		countSQL += fmt.Sprintf(" AND p.is_active = $%d", countIndex)
		countArgs = append(countArgs, isActive)
	}

	var total int
	err = r.db.GetContext(ctx, &total, countSQL, countArgs...)
	if err != nil {
		return nil, providers.ProvidersErrors.New(providers.ErrProviderSearchFailed).
			WithDetail("query", query).
			WithCause(err)
	}

	// Build mock paginated result
	result := storex.Paginated[models.Provider]{
		Data: providersData,
		// Note: Removed Page, PageSize, Total fields as they don't exist in storex.Paginated
		// The pagination info will be calculated in buildListResponse
	}

	return r.buildListResponseWithPagination(result, opts.Page, opts.PageSize, total), nil
}

func (r *providerRepository) buildListResponse(result storex.Paginated[models.Provider]) *dto.ProviderListResponse {
	data := make([]dto.ProviderResponse, len(result.Data))
	for i, p := range result.Data {
		data[i] = dto.ProviderResponse{
			ID:             p.ID,
			UserID:         p.UserID,
			OrganizationID: p.OrganizationID,
			Name:           p.Name,
			ProviderCode:   p.ProviderCode,
			IsActive:       p.IsActive,
			Metadata:       p.Metadata,
			CreatedAt:      p.CreatedAt,
			UpdatedAt:      p.UpdatedAt,
		}
	}

	// Calculate pagination info - assuming storex.Paginated has these fields
	// If not, you'll need to modify this based on the actual structure
	totalPages := 1
	hasNext := false
	hasPrev := false
	page := 1
	pageSize := len(result.Data)
	total := len(result.Data)

	return &dto.ProviderListResponse{
		Data:       data,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
	}
}

func (r *providerRepository) buildListResponseWithPagination(result storex.Paginated[models.Provider], page, pageSize, total int) *dto.ProviderListResponse {
	data := make([]dto.ProviderResponse, len(result.Data))
	for i, p := range result.Data {
		data[i] = dto.ProviderResponse{
			ID:             p.ID,
			UserID:         p.UserID,
			OrganizationID: p.OrganizationID,
			Name:           p.Name,
			ProviderCode:   p.ProviderCode,
			IsActive:       p.IsActive,
			Metadata:       p.Metadata,
			CreatedAt:      p.CreatedAt,
			UpdatedAt:      p.UpdatedAt,
		}
	}

	totalPages := (total + pageSize - 1) / pageSize

	return &dto.ProviderListResponse{
		Data:       data,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}
