package projectsapi

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/Abraxas-365/fuckturamelo/projects"
	"github.com/Abraxas-365/fuckturamelo/projects/dto"
	"github.com/Abraxas-365/fuckturamelo/projects/projectsrv"
	postgres "github.com/Abraxas-365/fuckturamelo/projects/repository"
)

// ProjectsAPI contains the complete API setup for the projects domain
type ProjectsAPI struct {
	service projectsrv.ProjectService
	repo    postgres.ProjectRepository
}

// Config contains configuration for the projects API
type Config struct {
	DB *sqlx.DB
}

// New creates a new ProjectsAPI instance
func New(config Config) (*ProjectsAPI, error) {
	if config.DB == nil {
		return nil, projects.ProjectsErrors.New(projects.ErrProjectValidationFailed).
			WithDetail("error", "Database connection is required")
	}

	// Initialize layers from bottom up
	repo := postgres.NewProjectRepository(config.DB)
	svc := projectsrv.NewProjectService(repo)

	return &ProjectsAPI{
		service: svc,
		repo:    repo,
	}, nil
}

// SetupRoutes registers all project routes with the given Fiber router group
func (api *ProjectsAPI) SetupRoutes(router fiber.Router) {
	// Basic CRUD routes
	router.Post("/", api.createProject)
	router.Get("/", api.listProjects)
	router.Get("/:id", api.getProject)
	router.Get("/:id/with-providers", api.getProjectWithProviders)
	router.Put("/:id", api.updateProject)
	router.Delete("/:id", api.deleteProject)

	// Special operation routes
	router.Post("/:id/activate", api.activateProject)
	router.Post("/:id/deactivate", api.deactivateProject)
	router.Post("/:id/duplicate", api.duplicateProject)

	// Provider management routes
	router.Post("/:id/providers", api.addProvider)
	router.Delete("/:id/providers/:providerId", api.removeProvider)
	router.Put("/:id/providers/:providerId/role", api.updateProviderRole)
	router.Get("/:id/providers", api.getProjectProviders)
	router.Post("/:id/providers/bulk", api.addProvidersBulk)
	router.Delete("/:id/providers/bulk", api.removeProvidersBulk)

	// Query routes
	router.Get("/search", api.searchProjects)
	router.Get("/organization/:orgId", api.getProjectsByOrganization)
	router.Get("/organization/:orgId/stats", api.getProjectStats)

	// Health check route
	router.Get("/health", api.healthCheck)
}

// GetService returns the service layer for dependency injection
func (api *ProjectsAPI) GetService() projectsrv.ProjectService {
	return api.service
}

// GetRepository returns the repository layer for dependency injection
func (api *ProjectsAPI) GetRepository() postgres.ProjectRepository {
	return api.repo
}

// Basic CRUD handlers

// createProject handles POST /projects
func (api *ProjectsAPI) createProject(c *fiber.Ctx) error {
	var req dto.CreateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return projects.ProjectsErrors.New(projects.ErrProjectValidationFailed).
			WithDetail("error", "Invalid JSON in request body").
			WithCause(err)
	}

	result, err := api.service.CreateProject(c.Context(), &req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// getProject handles GET /projects/:id
func (api *ProjectsAPI) getProject(c *fiber.Ctx) error {
	id, err := api.parseUUIDParam(c, "id")
	if err != nil {
		return err
	}

	result, err := api.service.GetProject(c.Context(), id)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// getProjectWithProviders handles GET /projects/:id/with-providers
func (api *ProjectsAPI) getProjectWithProviders(c *fiber.Ctx) error {
	id, err := api.parseUUIDParam(c, "id")
	if err != nil {
		return err
	}

	result, err := api.service.GetProjectWithProviders(c.Context(), id)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// updateProject handles PUT /projects/:id
func (api *ProjectsAPI) updateProject(c *fiber.Ctx) error {
	id, err := api.parseUUIDParam(c, "id")
	if err != nil {
		return err
	}

	var req dto.UpdateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return projects.ProjectsErrors.New(projects.ErrProjectValidationFailed).
			WithDetail("error", "Invalid JSON in request body").
			WithCause(err)
	}

	result, err := api.service.UpdateProject(c.Context(), id, &req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// deleteProject handles DELETE /projects/:id
func (api *ProjectsAPI) deleteProject(c *fiber.Ctx) error {
	id, err := api.parseUUIDParam(c, "id")
	if err != nil {
		return err
	}

	err = api.service.DeleteProject(c.Context(), id)
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// listProjects handles GET /projects
func (api *ProjectsAPI) listProjects(c *fiber.Ctx) error {
	req, err := api.parseListRequest(c)
	if err != nil {
		return err
	}

	result, err := api.service.ListProjects(c.Context(), req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// Special operation handlers

// activateProject handles POST /projects/:id/activate
func (api *ProjectsAPI) activateProject(c *fiber.Ctx) error {
	id, err := api.parseUUIDParam(c, "id")
	if err != nil {
		return err
	}

	result, err := api.service.ActivateProject(c.Context(), id)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// deactivateProject handles POST /projects/:id/deactivate
func (api *ProjectsAPI) deactivateProject(c *fiber.Ctx) error {
	id, err := api.parseUUIDParam(c, "id")
	if err != nil {
		return err
	}

	result, err := api.service.DeactivateProject(c.Context(), id)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// duplicateProject handles POST /projects/:id/duplicate
func (api *ProjectsAPI) duplicateProject(c *fiber.Ctx) error {
	id, err := api.parseUUIDParam(c, "id")
	if err != nil {
		return err
	}

	var req struct {
		Name string `json:"name" validate:"required,min=1,max=255"`
	}
	if err := c.BodyParser(&req); err != nil {
		return projects.ProjectsErrors.New(projects.ErrProjectValidationFailed).
			WithDetail("error", "Invalid JSON in request body").
			WithCause(err)
	}

	if req.Name == "" {
		return projects.ProjectsErrors.New(projects.ErrProjectValidationFailed).
			WithDetail("error", "Name is required")
	}

	result, err := api.service.DuplicateProject(c.Context(), id, req.Name)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// Provider management handlers

// addProvider handles POST /projects/:id/providers
func (api *ProjectsAPI) addProvider(c *fiber.Ctx) error {
	projectID, err := api.parseUUIDParam(c, "id")
	if err != nil {
		return err
	}

	var req dto.AddProviderRequest
	if err := c.BodyParser(&req); err != nil {
		return projects.ProjectsErrors.New(projects.ErrProjectValidationFailed).
			WithDetail("error", "Invalid JSON in request body").
			WithCause(err)
	}

	result, err := api.service.AddProvider(c.Context(), projectID, &req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// removeProvider handles DELETE /projects/:id/providers/:providerId
func (api *ProjectsAPI) removeProvider(c *fiber.Ctx) error {
	projectID, err := api.parseUUIDParam(c, "id")
	if err != nil {
		return err
	}

	providerID, err := api.parseUUIDParam(c, "providerId")
	if err != nil {
		return err
	}

	err = api.service.RemoveProvider(c.Context(), projectID, providerID)
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// updateProviderRole handles PUT /projects/:id/providers/:providerId/role
func (api *ProjectsAPI) updateProviderRole(c *fiber.Ctx) error {
	projectID, err := api.parseUUIDParam(c, "id")
	if err != nil {
		return err
	}

	providerID, err := api.parseUUIDParam(c, "providerId")
	if err != nil {
		return err
	}

	var req dto.UpdateProviderRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return projects.ProjectsErrors.New(projects.ErrProjectValidationFailed).
			WithDetail("error", "Invalid JSON in request body").
			WithCause(err)
	}

	result, err := api.service.UpdateProviderRole(c.Context(), projectID, providerID, &req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// getProjectProviders handles GET /projects/:id/providers
func (api *ProjectsAPI) getProjectProviders(c *fiber.Ctx) error {
	projectID, err := api.parseUUIDParam(c, "id")
	if err != nil {
		return err
	}

	result, err := api.service.GetProjectProviders(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// addProvidersBulk handles POST /projects/:id/providers/bulk
func (api *ProjectsAPI) addProvidersBulk(c *fiber.Ctx) error {
	projectID, err := api.parseUUIDParam(c, "id")
	if err != nil {
		return err
	}

	var req dto.BulkProviderRequest
	if err := c.BodyParser(&req); err != nil {
		return projects.ProjectsErrors.New(projects.ErrProjectValidationFailed).
			WithDetail("error", "Invalid JSON in request body").
			WithCause(err)
	}

	err = api.service.AddProvidersBulk(c.Context(), projectID, &req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Providers added successfully",
	})
}

// removeProvidersBulk handles DELETE /projects/:id/providers/bulk
func (api *ProjectsAPI) removeProvidersBulk(c *fiber.Ctx) error {
	projectID, err := api.parseUUIDParam(c, "id")
	if err != nil {
		return err
	}

	var req dto.BulkProviderRequest
	if err := c.BodyParser(&req); err != nil {
		return projects.ProjectsErrors.New(projects.ErrProjectValidationFailed).
			WithDetail("error", "Invalid JSON in request body").
			WithCause(err)
	}

	err = api.service.RemoveProvidersBulk(c.Context(), projectID, &req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Providers removed successfully",
	})
}

// Query handlers

// getProjectsByOrganization handles GET /projects/organization/:orgId
func (api *ProjectsAPI) getProjectsByOrganization(c *fiber.Ctx) error {
	orgID, err := api.parseUUIDParam(c, "orgId")
	if err != nil {
		return err
	}

	result, err := api.service.GetProjectsByOrganization(c.Context(), orgID)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// searchProjects handles GET /projects/search
func (api *ProjectsAPI) searchProjects(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return projects.ProjectsErrors.New(projects.ErrProjectValidationFailed).
			WithDetail("error", "Search query parameter 'q' is required")
	}

	orgIDStr := c.Query("organization_id")
	if orgIDStr == "" {
		return projects.ProjectsErrors.New(projects.ErrProjectValidationFailed).
			WithDetail("error", "Organization ID parameter is required")
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		return projects.ProjectsErrors.New(projects.ErrProjectValidationFailed).
			WithDetail("error", "Invalid organization ID format").
			WithCause(err)
	}

	result, err := api.service.SearchProjects(c.Context(), query, orgID)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// getProjectStats handles GET /projects/organization/:orgId/stats
func (api *ProjectsAPI) getProjectStats(c *fiber.Ctx) error {
	orgID, err := api.parseUUIDParam(c, "orgId")
	if err != nil {
		return err
	}

	result, err := api.service.GetProjectStats(c.Context(), orgID)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// healthCheck provides a health check endpoint
func (api *ProjectsAPI) healthCheck(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "healthy",
		"service": "projects",
	})
}

// Helper methods

func (api *ProjectsAPI) parseUUIDParam(c *fiber.Ctx, paramName string) (uuid.UUID, error) {
	paramValue := c.Params(paramName)
	if paramValue == "" {
		return uuid.Nil, projects.ProjectsErrors.New(projects.ErrProjectValidationFailed).
			WithDetail("error", "Missing required parameter: "+paramName)
	}

	id, err := uuid.Parse(paramValue)
	if err != nil {
		return uuid.Nil, projects.ProjectsErrors.New(projects.ErrProjectValidationFailed).
			WithDetail("error", "Invalid UUID format for parameter: "+paramName).
			WithCause(err)
	}

	return id, nil
}

func (api *ProjectsAPI) parseListRequest(c *fiber.Ctx) (*dto.ProjectListRequest, error) {
	req := &dto.ProjectListRequest{}

	// Parse organization_id
	if orgIDStr := c.Query("organization_id"); orgIDStr != "" {
		orgID, err := uuid.Parse(orgIDStr)
		if err != nil {
			return nil, projects.ProjectsErrors.New(projects.ErrProjectValidationFailed).
				WithDetail("error", "Invalid organization_id format").
				WithCause(err)
		}
		req.OrganizationID = &orgID
	}

	// Parse is_active
	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		isActive := isActiveStr == "true"
		req.IsActive = &isActive
	}

	// Parse search
	if search := c.Query("search"); search != "" {
		req.Search = &search
	}

	// Parse pagination
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	req.Page = page

	pageSize := 20
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}
	req.PageSize = pageSize

	// Parse sorting
	req.SortBy = c.Query("sort_by", "created_at")
	req.SortOrder = c.Query("sort_order", "desc")

	return req, nil
}
