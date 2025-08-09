package projects

import (
	"net/http"

	"github.com/Abraxas-365/craftable/errx"
)

// ProjectsErrors is the error registry for projects domain
var ProjectsErrors = errx.NewRegistry("PROJECTS")

// Project error codes
var (
	// Basic CRUD errors
	ErrProjectNotFound = ProjectsErrors.Register(
		"NOT_FOUND",
		errx.TypeNotFound,
		http.StatusNotFound,
		"Project not found",
	)

	ErrProjectNameExists = ProjectsErrors.Register(
		"NAME_EXISTS",
		errx.TypeConflict,
		http.StatusConflict,
		"Project with this name already exists in the organization",
	)

	ErrProjectCreateFailed = ProjectsErrors.Register(
		"CREATE_FAILED",
		errx.TypeInternal,
		http.StatusInternalServerError,
		"Failed to create project",
	)

	ErrProjectUpdateFailed = ProjectsErrors.Register(
		"UPDATE_FAILED",
		errx.TypeInternal,
		http.StatusInternalServerError,
		"Failed to update project",
	)

	ErrProjectDeleteFailed = ProjectsErrors.Register(
		"DELETE_FAILED",
		errx.TypeInternal,
		http.StatusInternalServerError,
		"Failed to delete project",
	)

	// Query errors
	ErrProjectListFailed = ProjectsErrors.Register(
		"LIST_FAILED",
		errx.TypeInternal,
		http.StatusInternalServerError,
		"Failed to list projects",
	)

	ErrProjectSearchFailed = ProjectsErrors.Register(
		"SEARCH_FAILED",
		errx.TypeInternal,
		http.StatusInternalServerError,
		"Failed to search projects",
	)

	// Validation errors
	ErrProjectValidationFailed = ProjectsErrors.Register(
		"VALIDATION_FAILED",
		errx.TypeValidation,
		http.StatusBadRequest,
		"Project validation failed",
	)

	// Business logic errors
	ErrProjectCannotDelete = ProjectsErrors.Register(
		"CANNOT_DELETE",
		errx.TypeBusiness,
		http.StatusConflict,
		"Project cannot be deleted due to existing references",
	)

	ErrProjectInactive = ProjectsErrors.Register(
		"INACTIVE",
		errx.TypeBusiness,
		http.StatusConflict,
		"Project is inactive and cannot be used",
	)

	// Provider relationship errors
	ErrProjectProviderNotFound = ProjectsErrors.Register(
		"PROVIDER_NOT_FOUND",
		errx.TypeNotFound,
		http.StatusNotFound,
		"Project provider relationship not found",
	)

	ErrProjectProviderExists = ProjectsErrors.Register(
		"PROVIDER_EXISTS",
		errx.TypeConflict,
		http.StatusConflict,
		"Provider is already associated with this project",
	)

	ErrProjectProviderManagementFailed = ProjectsErrors.Register(
		"PROVIDER_MANAGEMENT_FAILED",
		errx.TypeInternal,
		http.StatusInternalServerError,
		"Failed to manage project-provider relationship",
	)

	// Bulk operation errors
	ErrProjectBulkCreateFailed = ProjectsErrors.Register(
		"BULK_CREATE_FAILED",
		errx.TypeInternal,
		http.StatusInternalServerError,
		"Failed to create projects in bulk",
	)

	ErrProjectBulkUpdateFailed = ProjectsErrors.Register(
		"BULK_UPDATE_FAILED",
		errx.TypeInternal,
		http.StatusInternalServerError,
		"Failed to update projects in bulk",
	)

	ErrProjectBulkDeleteFailed = ProjectsErrors.Register(
		"BULK_DELETE_FAILED",
		errx.TypeInternal,
		http.StatusInternalServerError,
		"Failed to delete projects in bulk",
	)
)

// Helper functions for error checking
func IsProjectNotFound(err error) bool {
	return errx.IsCode(err, ErrProjectNotFound)
}

func IsProjectNameExists(err error) bool {
	return errx.IsCode(err, ErrProjectNameExists)
}

func IsProjectValidationFailed(err error) bool {
	return errx.IsCode(err, ErrProjectValidationFailed)
}

func IsProjectProviderNotFound(err error) bool {
	return errx.IsCode(err, ErrProjectProviderNotFound)
}

func IsProjectProviderExists(err error) bool {
	return errx.IsCode(err, ErrProjectProviderExists)
}
