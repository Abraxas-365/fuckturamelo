package providers

import (
	"net/http"

	"github.com/Abraxas-365/craftable/errx"
)

// ProvidersErrors is the error registry for providers domain
var ProvidersErrors = errx.NewRegistry("PROVIDERS")

// Provider error codes
var (
	// Basic CRUD errors
	ErrProviderNotFound = ProvidersErrors.Register(
		"NOT_FOUND",
		errx.TypeNotFound,
		http.StatusNotFound,
		"Provider not found",
	)

	ErrProviderNameExists = ProvidersErrors.Register(
		"NAME_EXISTS",
		errx.TypeConflict,
		http.StatusConflict,
		"Provider with this name already exists in the organization",
	)

	ErrProviderCreateFailed = ProvidersErrors.Register(
		"CREATE_FAILED",
		errx.TypeInternal,
		http.StatusInternalServerError,
		"Failed to create provider",
	)

	ErrProviderUpdateFailed = ProvidersErrors.Register(
		"UPDATE_FAILED",
		errx.TypeInternal,
		http.StatusInternalServerError,
		"Failed to update provider",
	)

	ErrProviderDeleteFailed = ProvidersErrors.Register(
		"DELETE_FAILED",
		errx.TypeInternal,
		http.StatusInternalServerError,
		"Failed to delete provider",
	)

	// Query errors
	ErrProviderListFailed = ProvidersErrors.Register(
		"LIST_FAILED",
		errx.TypeInternal,
		http.StatusInternalServerError,
		"Failed to list providers",
	)

	ErrProviderSearchFailed = ProvidersErrors.Register(
		"SEARCH_FAILED",
		errx.TypeInternal,
		http.StatusInternalServerError,
		"Failed to search providers",
	)

	// Validation errors
	ErrProviderValidationFailed = ProvidersErrors.Register(
		"VALIDATION_FAILED",
		errx.TypeValidation,
		http.StatusBadRequest,
		"Provider validation failed",
	)

	// Business logic errors
	ErrProviderCannotDelete = ProvidersErrors.Register(
		"CANNOT_DELETE",
		errx.TypeBusiness,
		http.StatusConflict,
		"Provider cannot be deleted due to existing references",
	)

	ErrProviderInactive = ProvidersErrors.Register(
		"INACTIVE",
		errx.TypeBusiness,
		http.StatusConflict,
		"Provider is inactive and cannot be used",
	)

	// Bulk operation errors
	ErrProviderBulkCreateFailed = ProvidersErrors.Register(
		"BULK_CREATE_FAILED",
		errx.TypeInternal,
		http.StatusInternalServerError,
		"Failed to create providers in bulk",
	)

	ErrProviderBulkUpdateFailed = ProvidersErrors.Register(
		"BULK_UPDATE_FAILED",
		errx.TypeInternal,
		http.StatusInternalServerError,
		"Failed to update providers in bulk",
	)

	ErrProviderBulkDeleteFailed = ProvidersErrors.Register(
		"BULK_DELETE_FAILED",
		errx.TypeInternal,
		http.StatusInternalServerError,
		"Failed to delete providers in bulk",
	)
)

// Helper functions for error checking
func IsProviderNotFound(err error) bool {
	return errx.IsCode(err, ErrProviderNotFound)
}

func IsProviderNameExists(err error) bool {
	return errx.IsCode(err, ErrProviderNameExists)
}

func IsProviderValidationFailed(err error) bool {
	return errx.IsCode(err, ErrProviderValidationFailed)
}
