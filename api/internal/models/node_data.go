package models

import (
	"encoding/json"
	"fmt"
)

// NodeData is the interface that all node data types must implement
type NodeData interface {
	// GetNodeType returns the type of node this data belongs to
	GetNodeType() string
	// Validate performs validation specific to this node type
	Validate() error
}

// StartNodeData represents data for start nodes
type StartNodeData struct {
	Label       string            `json:"label"`
	Description string            `json:"description"`
	Metadata    StartNodeMetadata `json:"metadata"`
}

type StartNodeMetadata struct {
	HasHandles      HandleConfig `json:"hasHandles"`
	OutputVariables []string     `json:"outputVariables,omitempty"`
}

func (d StartNodeData) GetNodeType() string { return NodeTypeStart }
func (d StartNodeData) Validate() error     { return nil }

// FormNodeData represents data for form nodes
type FormNodeData struct {
	Label       string           `json:"label"`
	Description string           `json:"description"`
	Metadata    FormNodeMetadata `json:"metadata"`
}

type FormNodeMetadata struct {
	HasHandles      HandleConfig `json:"hasHandles"`
	InputFields     []string     `json:"inputFields"`
	OutputVariables []string     `json:"outputVariables"`
}

func (d FormNodeData) GetNodeType() string { return NodeTypeForm }
func (d FormNodeData) Validate() error {
	if len(d.Metadata.InputFields) == 0 {
		return fmt.Errorf("form node must have at least one input field")
	}
	return nil
}

// IntegrationNodeData represents data for integration nodes
type IntegrationNodeData struct {
	Label       string                  `json:"label"`
	Description string                  `json:"description"`
	Metadata    IntegrationNodeMetadata `json:"metadata"`
}

type IntegrationNodeMetadata struct {
	HasHandles      HandleConfig     `json:"hasHandles"`
	InputVariables  []string         `json:"inputVariables"`
	APIEndpoint     string           `json:"apiEndpoint"`
	Options         []LocationOption `json:"options"`
	OutputVariables []string         `json:"outputVariables"`
}

type LocationOption struct {
	City string  `json:"city"`
	Lat  float64 `json:"lat"`
	Lon  float64 `json:"lon"`
}

func (d IntegrationNodeData) GetNodeType() string { return NodeTypeIntegration }
func (d IntegrationNodeData) Validate() error {
	if d.Metadata.APIEndpoint == "" {
		return fmt.Errorf("integration node must have an API endpoint")
	}
	return nil
}

// ConditionNodeData represents data for condition nodes
type ConditionNodeData struct {
	Label       string                `json:"label"`
	Description string                `json:"description"`
	Metadata    ConditionNodeMetadata `json:"metadata"`
}

type ConditionNodeMetadata struct {
	HasHandles          HandleConfigWithBranches `json:"hasHandles"`
	ConditionExpression string                   `json:"conditionExpression"`
	OutputVariables     []string                 `json:"outputVariables"`
}

type HandleConfigWithBranches struct {
	Source []string `json:"source"` // ["true", "false"] for conditions
	Target bool     `json:"target"`
}

func (d ConditionNodeData) GetNodeType() string { return NodeTypeCondition }
func (d ConditionNodeData) Validate() error {
	if d.Metadata.ConditionExpression == "" {
		return fmt.Errorf("condition node must have a condition expression")
	}
	return nil
}

// EmailNodeData represents data for email nodes
type EmailNodeData struct {
	Label       string            `json:"label"`
	Description string            `json:"description"`
	Metadata    EmailNodeMetadata `json:"metadata"`
}

type EmailNodeMetadata struct {
	HasHandles      HandleConfig  `json:"hasHandles"`
	InputVariables  []string      `json:"inputVariables"`
	EmailTemplate   EmailTemplate `json:"emailTemplate"`
	OutputVariables []string      `json:"outputVariables"`
}

type EmailTemplate struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func (d EmailNodeData) GetNodeType() string { return NodeTypeEmail }
func (d EmailNodeData) Validate() error {
	if d.Metadata.EmailTemplate.Subject == "" {
		return fmt.Errorf("email node must have a subject")
	}
	if d.Metadata.EmailTemplate.Body == "" {
		return fmt.Errorf("email node must have a body")
	}
	return nil
}

// EndNodeData represents data for end nodes
type EndNodeData struct {
	Label       string          `json:"label"`
	Description string          `json:"description"`
	Metadata    EndNodeMetadata `json:"metadata"`
}

type EndNodeMetadata struct {
	HasHandles     HandleConfig `json:"hasHandles"`
	InputVariables []string     `json:"inputVariables,omitempty"`
}

func (d EndNodeData) GetNodeType() string { return NodeTypeEnd }
func (d EndNodeData) Validate() error     { return nil }

// HandleConfig represents the standard handle configuration
type HandleConfig struct {
	Source bool `json:"source"`
	Target bool `json:"target"`
}

// NodeDataUnion represents a union type for all possible node data
type NodeDataUnion struct {
	Type string `json:"-"` // Set during unmarshaling
	Data NodeData
}

// UnmarshalJSON implements custom unmarshaling for NodeDataUnion
func (ndu *NodeDataUnion) UnmarshalJSON(data []byte) error {
	// First, extract just the type information if available
	// We need to determine the type from context since it's not in the data itself

	// Try each type until one works
	types := []struct {
		nodeType string
		target   NodeData
	}{
		{NodeTypeStart, &StartNodeData{}},
		{NodeTypeForm, &FormNodeData{}},
		{NodeTypeIntegration, &IntegrationNodeData{}},
		{NodeTypeCondition, &ConditionNodeData{}},
		{NodeTypeEmail, &EmailNodeData{}},
		{NodeTypeEnd, &EndNodeData{}},
	}

	var lastErr error
	for _, t := range types {
		if err := json.Unmarshal(data, t.target); err == nil {
			// Validate that this is the correct type
			if err := t.target.Validate(); err == nil {
				ndu.Type = t.nodeType
				ndu.Data = t.target
				return nil
			}
		} else {
			lastErr = err
		}
	}

	return fmt.Errorf("failed to unmarshal node data into any known type: %v", lastErr)
}

// MarshalJSON implements custom marshaling for NodeDataUnion
func (ndu NodeDataUnion) MarshalJSON() ([]byte, error) {
	if ndu.Data == nil {
		return []byte("null"), nil
	}
	return json.Marshal(ndu.Data)
}

// ParseNodeData parses raw JSON into the appropriate strongly typed NodeData
func ParseNodeData(nodeType string, rawData []byte) (NodeData, error) {
	switch nodeType {
	case NodeTypeStart:
		var data StartNodeData
		if err := json.Unmarshal(rawData, &data); err != nil {
			return nil, fmt.Errorf("failed to parse start node data: %w", err)
		}
		return data, data.Validate()

	case NodeTypeForm:
		var data FormNodeData
		if err := json.Unmarshal(rawData, &data); err != nil {
			return nil, fmt.Errorf("failed to parse form node data: %w", err)
		}
		return data, data.Validate()

	case NodeTypeIntegration:
		var data IntegrationNodeData
		if err := json.Unmarshal(rawData, &data); err != nil {
			return nil, fmt.Errorf("failed to parse integration node data: %w", err)
		}
		return data, data.Validate()

	case NodeTypeCondition:
		var data ConditionNodeData
		if err := json.Unmarshal(rawData, &data); err != nil {
			return nil, fmt.Errorf("failed to parse condition node data: %w", err)
		}
		return data, data.Validate()

	case NodeTypeEmail:
		var data EmailNodeData
		if err := json.Unmarshal(rawData, &data); err != nil {
			return nil, fmt.Errorf("failed to parse email node data: %w", err)
		}
		return data, data.Validate()

	case NodeTypeEnd:
		var data EndNodeData
		if err := json.Unmarshal(rawData, &data); err != nil {
			return nil, fmt.Errorf("failed to parse end node data: %w", err)
		}
		return data, data.Validate()

	default:
		return nil, fmt.Errorf("unknown node type: %s", nodeType)
	}
}
