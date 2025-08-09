package providersapi

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/Abraxas-365/fuckturamelo/providers"
	"github.com/Abraxas-365/fuckturamelo/providers/dto"
	postgres "github.com/Abraxas-365/fuckturamelo/providers/repository"
	"github.com/Abraxas-365/fuckturamelo/providers/service"
)

// ProvidersAPI contains the complete API setup for the providers domain
type ProvidersAPI struct {
	service service.ProviderService
	repo    postgres.ProviderRepository
}

// Config contains configuration for the providers API
type Config struct {
	DB *sqlx.DB
}

// New creates a new ProvidersAPI instance
func New(config Config) (*ProvidersAPI, error) {
	if config.DB == nil {
		return nil, providers.ProvidersErrors.New(providers.ErrProviderValidationFailed).
			WithDetail("error", "Database connection is required")
	}

	// Initialize layers from bottom up
	repo := postgres.NewProviderRepository(config.DB)
	svc := service.NewProviderService(repo)

	return &ProvidersAPI{
		service: svc,
		repo:    repo,
	}, nil
}

// SetupRoutes registers all provider routes with the given Fiber router group
func (api *ProvidersAPI) SetupRoutes(router fiber.Router) {
	// Basic CRUD routes
	router.Post("/", api.createProvider)
	router.Get("/", api.listProviders)
	router.Get("/:id", api.getProvider)
	router.Put("/:id", api.updateProvider)
	router.Delete("/:id", api.deleteProvider)

	// Special operation routes
	router.Post("/:id/activate", api.activateProvider)
	router.Post("/:id/deactivate", api.deactivateProvider)
	router.Post("/:id/duplicate", api.duplicateProvider)

	// Query routes
	router.Get("/search", api.searchProviders)
	router.Get("/organization/:orgId", api.getProvidersByOrganization)

	// Health check route
	router.Get("/health", api.healthCheck)
}

// GetService returns the service layer for dependency injection
func (api *ProvidersAPI) GetService() service.ProviderService {
	return api.service
}

// GetRepository returns the repository layer for dependency injection
func (api *ProvidersAPI) GetRepository() postgres.ProviderRepository {
	return api.repo
}

// createProvider handles POST /providers
func (api *ProvidersAPI) createProvider(c *fiber.Ctx) error {
	var req dto.CreateProviderRequest
	if err := c.BodyParser(&req); err != nil {
		return providers.ProvidersErrors.New(providers.ErrProviderValidationFailed).
			WithDetail("error", "Invalid JSON in request body").
			WithCause(err)
	}

	result, err := api.service.CreateProvider(c.Context(), &req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// getProvider handles GET /providers/:id
func (api *ProvidersAPI) getProvider(c *fiber.Ctx) error {
	id, err := api.parseUUIDParam(c, "id")
	if err != nil {
		return err
	}

	result, err := api.service.GetProvider(c.Context(), id)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// updateProvider handles PUT /providers/:id
func (api *ProvidersAPI) updateProvider(c *fiber.Ctx) error {
	id, err := api.parseUUIDParam(c, "id")
	if err != nil {
		return err
	}

	var req dto.UpdateProviderRequest
	if err := c.BodyParser(&req); err != nil {
		return providers.ProvidersErrors.New(providers.ErrProviderValidationFailed).
			WithDetail("error", "Invalid JSON in request body").
			WithCause(err)
	}

	result, err := api.service.UpdateProvider(c.Context(), id, &req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// deleteProvider handles DELETE /providers/:id
func (api *ProvidersAPI) deleteProvider(c *fiber.Ctx) error {
	id, err := api.parseUUIDParam(c, "id")
	if err != nil {
		return err
	}

	err = api.service.DeleteProvider(c.Context(), id)
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// listProviders handles GET /providers
func (api *ProvidersAPI) listProviders(c *fiber.Ctx) error {
	req, err := api.parseListRequest(c)
	if err != nil {
		return err
	}

	result, err := api.service.ListProviders(c.Context(), req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// getProvidersByOrganization handles GET /providers/organization/:orgId
func (api *ProvidersAPI) getProvidersByOrganization(c *fiber.Ctx) error {
	orgID, err := api.parseUUIDParam(c, "orgId")
	if err != nil {
		return err
	}

	result, err := api.service.GetProvidersByOrganization(c.Context(), orgID)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// searchProviders handles GET /providers/search
func (api *ProvidersAPI) searchProviders(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return providers.ProvidersErrors.New(providers.ErrProviderValidationFailed).
			WithDetail("error", "Search query parameter 'q' is required")
	}

	orgIDStr := c.Query("organization_id")
	if orgIDStr == "" {
		return providers.ProvidersErrors.New(providers.ErrProviderValidationFailed).
			WithDetail("error", "Organization ID parameter is required")
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		return providers.ProvidersErrors.New(providers.ErrProviderValidationFailed).
			WithDetail("error", "Invalid organization ID format").
			WithCause(err)
	}

	result, err := api.service.SearchProviders(c.Context(), query, orgID)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// activateProvider handles POST /providers/:id/activate
func (api *ProvidersAPI) activateProvider(c *fiber.Ctx) error {
	id, err := api.parseUUIDParam(c, "id")
	if err != nil {
		return err
	}

	result, err := api.service.ActivateProvider(c.Context(), id)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// deactivateProvider handles POST /providers/:id/deactivate
func (api *ProvidersAPI) deactivateProvider(c *fiber.Ctx) error {
	id, err := api.parseUUIDParam(c, "id")
	if err != nil {
		return err
	}

	result, err := api.service.DeactivateProvider(c.Context(), id)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// duplicateProvider handles POST /providers/:id/duplicate
func (api *ProvidersAPI) duplicateProvider(c *fiber.Ctx) error {
	id, err := api.parseUUIDParam(c, "id")
	if err != nil {
		return err
	}

	var req struct {
		Name string `json:"name" validate:"required,min=1,max=255"`
	}
	if err := c.BodyParser(&req); err != nil {
		return providers.ProvidersErrors.New(providers.ErrProviderValidationFailed).
			WithDetail("error", "Invalid JSON in request body").
			WithCause(err)
	}

	if req.Name == "" {
		return providers.ProvidersErrors.New(providers.ErrProviderValidationFailed).
			WithDetail("error", "Name is required")
	}

	result, err := api.service.DuplicateProvider(c.Context(), id, req.Name)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// healthCheck provides a health check endpoint
func (api *ProvidersAPI) healthCheck(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "healthy",
		"service": "providers",
	})
}

func (api *ProvidersAPI) parseUUIDParam(c *fiber.Ctx, paramName string) (uuid.UUID, error) {
	paramValue := c.Params(paramName)
	if paramValue == "" {
		return uuid.Nil, providers.ProvidersErrors.New(providers.ErrProviderValidationFailed).
			WithDetail("error", "Missing required parameter: "+paramName)
	}

	id, err := uuid.Parse(paramValue)
	if err != nil {
		return uuid.Nil, providers.ProvidersErrors.New(providers.ErrProviderValidationFailed).
			WithDetail("error", "Invalid UUID format for parameter: "+paramName).
			WithCause(err)
	}

	return id, nil
}

func (api *ProvidersAPI) parseListRequest(c *fiber.Ctx) (*dto.ProviderListRequest, error) {
	req := &dto.ProviderListRequest{}

	// Parse organization_id
	if orgIDStr := c.Query("organization_id"); orgIDStr != "" {
		orgID, err := uuid.Parse(orgIDStr)
		if err != nil {
			return nil, providers.ProvidersErrors.New(providers.ErrProviderValidationFailed).
				WithDetail("error", "Invalid organization_id format").
				WithCause(err)
		}
		req.OrganizationID = &orgID
	}

	// Parse is_active
	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		isActive, err := strconv.ParseBool(isActiveStr)
		if err != nil {
			return nil, providers.ProvidersErrors.New(providers.ErrProviderValidationFailed).
				WithDetail("error", "Invalid is_active format").
				WithCause(err)
		}
		req.IsActive = &isActive
	}

	// Parse search
	if search := c.Query("search"); search != "" {
		req.Search = &search
	}

	// Parse page
	if pageStr := c.Query("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			return nil, providers.ProvidersErrors.New(providers.ErrProviderValidationFailed).
				WithDetail("error", "Invalid page number")
		}
		req.Page = page
	}

	// Parse page_size
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil || pageSize < 1 || pageSize > 100 {
			return nil, providers.ProvidersErrors.New(providers.ErrProviderValidationFailed).
				WithDetail("error", "Invalid page size (must be 1-100)")
		}
		req.PageSize = pageSize
	}

	// Parse order_by
	if orderBy := c.Query("order_by"); orderBy != "" {
		req.OrderBy = &orderBy
	}

	// Parse desc
	if descStr := c.Query("desc"); descStr != "" {
		desc, err := strconv.ParseBool(descStr)
		if err != nil {
			return nil, providers.ProvidersErrors.New(providers.ErrProviderValidationFailed).
				WithDetail("error", "Invalid desc format").
				WithCause(err)
		}
		req.Desc = desc
	}

	return req, nil
}
