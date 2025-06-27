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
	PositionX  float64         `json:"-" db:"position_x"`
	PositionY  float64         `json:"-" db:"position_y"`
	Data       NodeData        `json:"data" db:"-"` // Strongly typed data
	RawData    json.RawMessage `json:"-" db:"data"` // For database storage
	WorkflowID uuid.UUID       `json:"-" db:"workflow_id"`
	CreatedAt  time.Time       `json:"-" db:"created_at"`
	UpdatedAt  time.Time       `json:"-" db:"updated_at"`
}

// Position represents the x,y coordinates of a node
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// NodeResponse represents a node as returned to the frontend
type NodeResponse struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Position Position `json:"position"`
	Data     NodeData `json:"data"`
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
	Data     NodeData        `json:"data"`
	RawData  json.RawMessage `json:"-"` // For storing raw JSON during unmarshaling
}

// UnmarshalJSON implements custom JSON unmarshaling for NodeRequest
func (nr *NodeRequest) UnmarshalJSON(data []byte) error {
	// First unmarshal into a temporary struct to get the basic fields
	var temp struct {
		ID       string          `json:"id"`
		Type     string          `json:"type"`
		Position Position        `json:"position"`
		Data     json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("failed to unmarshal node request: %w", err)
	}

	// Set the basic fields
	nr.ID = temp.ID
	nr.Type = temp.Type
	nr.Position = temp.Position
	nr.RawData = temp.Data

	// Parse the strongly typed data using the node type
	if len(temp.Data) > 0 {
		parsedData, err := ParseNodeData(temp.Type, temp.Data)
		if err != nil {
			return fmt.Errorf("failed to parse node data for type %s: %w", temp.Type, err)
		}
		nr.Data = parsedData
	}

	return nil
}

// ToNode converts a NodeRequest to a Node for database storage
func (nr *NodeRequest) ToNode() (*Node, error) {
	// Marshal the strongly typed data to JSON for database storage
	rawData, err := json.Marshal(nr.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal node data: %w", err)
	}

	return &Node{
		ID:        nr.ID,
		Type:      nr.Type,
		PositionX: nr.Position.X,
		PositionY: nr.Position.Y,
		Data:      nr.Data,
		RawData:   rawData,
	}, nil
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

// LoadDataFromRaw parses the RawData into the strongly typed Data field
func (n *Node) LoadDataFromRaw() error {
	if n.RawData == nil {
		return fmt.Errorf("no raw data to parse")
	}

	parsedData, err := ParseNodeData(n.Type, n.RawData)
	if err != nil {
		return fmt.Errorf("failed to parse node data: %w", err)
	}

	n.Data = parsedData
	return nil
}

// UpdateRawDataFromData marshals the strongly typed Data into RawData for database storage
func (n *Node) UpdateRawDataFromData() error {
	if n.Data == nil {
		return fmt.Errorf("no data to marshal")
	}

	rawData, err := json.Marshal(n.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal node data: %w", err)
	}

	n.RawData = rawData
	return nil
}

// ValidateData validates the strongly typed data
func (n *Node) ValidateData() error {
	if n.Data == nil {
		return fmt.Errorf("no data to validate")
	}

	if n.Data.GetNodeType() != n.Type {
		return fmt.Errorf("data type mismatch: expected %s, got %s", n.Type, n.Data.GetNodeType())
	}

	return n.Data.Validate()
}
