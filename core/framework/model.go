package framework

import (
	"time"

	"github.com/google/uuid"
)

type (
	// Model is the baseModel underlaying all models
	Model struct {
		ID        uuid.UUID `json:"id" gorm:"primarykey"`
		CreatedAt time.Time `json:"createdAt"`
		UpdatedAt time.Time `json:"updatedAt"`
	}
)

// GetID returns the ID of the model
func (m *Model) GetID() uuid.UUID {
	return m.ID
}

// Validate validates the struct tags against the input
func (m *Model) Validate() bool {
	return true
}
