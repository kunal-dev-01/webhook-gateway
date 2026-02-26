package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"shared/utils"
)

// ============================================================================
// JSON SCHEMA (for payload validation)
// ============================================================================

// JSONSchema represents an optional JSON Schema definition for payload validation
type JSONSchema map[string]any

func (s JSONSchema) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

func (s *JSONSchema) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONSchema value")
	}
	return json.Unmarshal(bytes, s)
}

// ============================================================================
// EVENT TYPE MODEL
// ============================================================================

// EventType defines an event type within a project
// Example: 'payment.success', 'user.created', 'order.cancelled'
type EventType struct {
	ID        uuid.UUID `db:"id" json:"id"`
	ProjectID uuid.UUID `db:"project_id" json:"project_id"`

	// Event type name - dot notation recommended
	Name        string  `db:"name" json:"name"`
	Description *string `db:"description" json:"description,omitempty"`

	// Optional JSON Schema for payload validation
	// NULL = no validation, accept anything
	SchemaDefinition *JSONSchema `db:"schema_definition" json:"schema_definition,omitempty"`

	IsActive bool `db:"is_active" json:"is_active"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ============================================================================
// EVENT TYPE HELPERS
// ============================================================================

// HasSchema checks if event type has schema validation
func (e *EventType) HasSchema() bool {
	return e.SchemaDefinition != nil
}

// ============================================================================
// VALIDATION
// ============================================================================

func (e *EventType) Validate() error {
	if e.ProjectID == uuid.Nil {
		return fmt.Errorf("project ID is required")
	}
	if err := utils.ValidateRequired(e.Name, "name"); err != nil {
		return err
	}
	if err := utils.ValidateLength(e.Name, 1, 255, "name"); err != nil {
		return err
	}
	return nil
}

// ============================================================================
// PUBLIC RESPONSE
// ============================================================================

type PublicEventType struct {
	ID               uuid.UUID  `json:"id"`
	ProjectID        uuid.UUID  `json:"project_id"`
	Name             string     `json:"name"`
	Description      *string    `json:"description,omitempty"`
	HasSchema        bool       `json:"has_schema"`
	IsActive         bool       `json:"is_active"`
	CreatedAt        time.Time  `json:"created_at"`
}

func (e *EventType) ToPublic() *PublicEventType {
	return &PublicEventType{
		ID:          e.ID,
		ProjectID:   e.ProjectID,
		Name:        e.Name,
		Description: e.Description,
		HasSchema:   e.HasSchema(),
		IsActive:    e.IsActive,
		CreatedAt:   e.CreatedAt,
	}
}

// ============================================================================
// ENDPOINT EVENT TYPE MODEL (Junction table)
// ============================================================================

// EndpointEventType represents an event type subscription for an endpoint
// Agar endpoint ke liye koi row nahi = saare events milenge (no filter)
// Agar rows hain = sirf wahi events milenge (whitelist)
type EndpointEventType struct {
	EndpointID  uuid.UUID `db:"endpoint_id" json:"endpoint_id"`
	EventTypeID uuid.UUID `db:"event_type_id" json:"event_type_id"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

// ============================================================================
// ENDPOINT EVENT TYPE HELPERS
// ============================================================================

// NewEndpointEventType creates a new subscription
func NewEndpointEventType(endpointID, eventTypeID uuid.UUID) *EndpointEventType {
	return &EndpointEventType{
		EndpointID:  endpointID,
		EventTypeID: eventTypeID,
		CreatedAt:   time.Now(),
	}
}

func (e *EndpointEventType) Validate() error {
	if e.EndpointID == uuid.Nil {
		return fmt.Errorf("endpoint ID is required")
	}
	if e.EventTypeID == uuid.Nil {
		return fmt.Errorf("event type ID is required")
	}
	return nil
}