package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Valid node types
const (
	NodeTypeStart       = "start"
	NodeTypeForm        = "form"
	NodeTypeIntegration = "integration"
	NodeTypeCondition   = "condition"
	NodeTypeEmail       = "email"
	NodeTypeEnd         = "end"
)

// ValidNodeTypes contains all allowed node types as a set for O(1) lookups
var ValidNodeTypes = map[string]bool{
	NodeTypeStart:       true,
	NodeTypeForm:        true,
	NodeTypeIntegration: true,
	NodeTypeCondition:   true,
	NodeTypeEmail:       true,
	NodeTypeEnd:         true,
}

// Node represents a workflow node with its position and data
type Node struct {
	ID         string          `json:"id" db:"id"`
	Type       string          `json:"type" db:"type"`
	PositionX  int             `json:"-" db:"position_x"`
	PositionY  int             `json:"-" db:"position_y"`
	Data       json.RawMessage `json:"data" db:"data"`
	WorkflowID uuid.UUID       `json:"-" db:"workflow_id"`
	CreatedAt  time.Time       `json:"-" db:"created_at"`
	UpdatedAt  time.Time       `json:"-" db:"updated_at"`
}

// Position represents the x,y coordinates of a node
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// NodeResponse represents a node as returned to the frontend
type NodeResponse struct {
	ID       string          `json:"id"`
	Type     string          `json:"type"`
	Position Position        `json:"position"`
	Data     json.RawMessage `json:"data"`
}

// ToResponse converts a Node to NodeResponse format for API responses
func (n *Node) ToResponse() NodeResponse {
	return NodeResponse{
		ID:   n.ID,
		Type: n.Type,
		Position: Position{
			X: n.PositionX,
			Y: n.PositionY,
		},
		Data: n.Data,
	}
}

// NodeRequest represents a node as sent from the frontend
type NodeRequest struct {
	ID       string          `json:"id"`
	Type     string          `json:"type"`
	Position Position        `json:"position"`
	Data     json.RawMessage `json:"data"`
}

// ToNode converts a NodeRequest to a Node for database storage
func (nr *NodeRequest) ToNode() *Node {
	return &Node{
		ID:        nr.ID,
		Type:      nr.Type,
		PositionX: nr.Position.X,
		PositionY: nr.Position.Y,
		Data:      nr.Data,
	}
}

// Validate checks if the node has a valid type
func (n *Node) Validate() error {
	return ValidateNodeType(n.Type)
}

// Validate checks if the node request has a valid type
func (nr *NodeRequest) Validate() error {
	return ValidateNodeType(nr.Type)
}

// ValidateNodeType checks if the given type is a valid node type
func ValidateNodeType(nodeType string) error {
	if ValidNodeTypes[nodeType] {
		return nil
	}

	// Get valid types for error message
	validTypes := make([]string, 0, len(ValidNodeTypes))
	for nodeType := range ValidNodeTypes {
		validTypes = append(validTypes, nodeType)
	}

	return fmt.Errorf("invalid node type '%s', must be one of: %v", nodeType, validTypes)
}

// IsStartNode returns true if this is a start node
func (n *Node) IsStartNode() bool {
	return n.Type == NodeTypeStart
}

// IsEndNode returns true if this is an end node
func (n *Node) IsEndNode() bool {
	return n.Type == NodeTypeEnd
}
