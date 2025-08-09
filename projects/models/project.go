package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Project represents a project in the system
type Project struct {
	ID             uuid.UUID `db:"id" json:"id"`
	OrganizationID uuid.UUID `db:"organization_id" json:"organization_id"`
	Name           string    `db:"name" json:"name"`
	Description    *string   `db:"description" json:"description"`
	IsActive       bool      `db:"is_active" json:"is_active"`
	Metadata       Metadata  `db:"metadata" json:"metadata"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}

// ProjectProvider represents the relationship between projects and providers
type ProjectProvider struct {
	ID             uuid.UUID `db:"id" json:"id"`
	ProjectID      uuid.UUID `db:"project_id" json:"project_id"`
	ProviderID     uuid.UUID `db:"provider_id" json:"provider_id"`
	OrganizationID uuid.UUID `db:"organization_id" json:"organization_id"`
	Role           *string   `db:"role" json:"role"`
	IsActive       bool      `db:"is_active" json:"is_active"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}

// ProjectWithProviders represents a project with its associated providers
type ProjectWithProviders struct {
	Project   `json:",inline"`
	Providers []ProjectProviderDetails `json:"providers,omitempty"`
}

// ProjectProviderDetails represents detailed provider information within a project
type ProjectProviderDetails struct {
	ProviderID   uuid.UUID `json:"provider_id"`
	ProviderName string    `json:"provider_name"`
	Role         *string   `json:"role"`
	IsActive     bool      `json:"is_active"`
	JoinedAt     time.Time `json:"joined_at"`
}

// Metadata represents flexible metadata stored as JSONB
type Metadata map[string]any

// Value implements the driver.Valuer interface for database storage
func (m Metadata) Value() (driver.Value, error) {
	if m == nil {
		return "{}", nil
	}
	return json.Marshal(m)
}

// Scan implements the sql.Scanner interface for database retrieval
func (m *Metadata) Scan(value any) error {
	if value == nil {
		*m = make(Metadata)
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, m)
	case string:
		return json.Unmarshal([]byte(v), m)
	default:
		return fmt.Errorf("cannot scan %T into Metadata", value)
	}
}

// TableName returns the table name for the Project model
func (p Project) TableName() string {
	return "projects"
}

// TableName returns the table name for the ProjectProvider model
func (pp ProjectProvider) TableName() string {
	return "project_providers"
}

// Validate performs basic validation on the Project model
func (p *Project) Validate() error {
	if p.OrganizationID == uuid.Nil {
		return fmt.Errorf("organization_id is required")
	}
	if p.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(p.Name) > 255 {
		return fmt.Errorf("name cannot exceed 255 characters")
	}
	return nil
}
