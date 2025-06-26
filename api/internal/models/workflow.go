package models

import (
	"time"

	"github.com/google/uuid"
)

// Workflow represents a complete workflow definition
type Workflow struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"-" db:"created_at"`
	UpdatedAt time.Time `json:"-" db:"updated_at"`
}

// WorkflowResponse represents a complete workflow as returned to the frontend
type WorkflowResponse struct {
	ID    string         `json:"id"`
	Name  string         `json:"name,omitempty"`
	Nodes []NodeResponse `json:"nodes"`
	Edges []EdgeResponse `json:"edges"`
}

// WorkflowRequest represents the workflow data sent from the frontend
type WorkflowRequest struct {
	ID    string        `json:"id"`
	Name  string        `json:"name,omitempty"`
	Nodes []NodeRequest `json:"nodes"`
	Edges []EdgeRequest `json:"edges"`
}
