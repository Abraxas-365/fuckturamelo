package models

import (
	"time"

	"github.com/google/uuid"
)

// Provider represents the provider domain entity
type Provider struct {
	ID             uuid.UUID              `json:"id" db:"id"`
	UserID         *uuid.UUID             `json:"user_id,omitempty" db:"user_id"`
	OrganizationID uuid.UUID              `json:"organization_id" db:"organization_id"`
	Name           string                 `json:"name" db:"name"`
	ProviderCode   *string                `json:"provider_code,omitempty" db:"provider_code"`
	IsActive       bool                   `json:"is_active" db:"is_active"`
	Metadata       map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at" db:"updated_at"`
}
