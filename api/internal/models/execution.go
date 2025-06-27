package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// ExecutionRequest represents the request payload for workflow execution
type ExecutionRequest struct {
	FormData  map[string]interface{} `json:"formData"`
	Condition map[string]interface{} `json:"condition"`
}

// ExecutionResponse represents the complete execution result
type ExecutionResponse struct {
	ExecutedAt time.Time       `json:"executedAt"`
	Status     string          `json:"status"`
	Steps      []ExecutionStep `json:"steps"`
	Error      *string         `json:"error,omitempty"`
}

// ExecutionStep represents a single step in the workflow execution
type ExecutionStep struct {
	NodeID      string          `json:"nodeId"`
	Type        string          `json:"type"`
	Label       string          `json:"label"`
	Description string          `json:"description"`
	Status      string          `json:"status"`
	Output      ExecutionOutput `json:"output,omitempty"` // Strongly typed output
	RawOutput   json.RawMessage `json:"-"`                // For database storage
	Error       *string         `json:"error,omitempty"`
	Duration    *int64          `json:"duration,omitempty"` // milliseconds
}

// ExecutionContext holds the runtime state during workflow execution
type ExecutionContext struct {
	WorkflowID string
	FormData   map[string]interface{}
	Variables  map[string]interface{}
	Steps      []ExecutionStep
	StartTime  time.Time
}

// NewExecutionContext creates a new execution context
func NewExecutionContext(workflowID string, formData map[string]interface{}) *ExecutionContext {
	return &ExecutionContext{
		WorkflowID: workflowID,
		FormData:   formData,
		Variables:  make(map[string]interface{}),
		Steps:      make([]ExecutionStep, 0),
		StartTime:  time.Now(),
	}
}

// AddStep adds a step to the execution context
func (ctx *ExecutionContext) AddStep(step ExecutionStep) {
	ctx.Steps = append(ctx.Steps, step)
}

// SetVariable sets a variable in the execution context
func (ctx *ExecutionContext) SetVariable(key string, value interface{}) {
	ctx.Variables[key] = value
}

// GetVariable gets a variable from the execution context
func (ctx *ExecutionContext) GetVariable(key string) (interface{}, bool) {
	value, ok := ctx.Variables[key]
	return value, ok
}

// Legacy data structures - these are replaced by the strongly typed versions in node_data.go
// Keeping for backward compatibility during migration

// FormField represents a single form field configuration (legacy)
type FormField struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Label       string   `json:"label"`
	Required    bool     `json:"required"`
	Options     []string `json:"options,omitempty"`    // for select fields
	Validation  string   `json:"validation,omitempty"` // validation rules
	Placeholder string   `json:"placeholder,omitempty"`
}

// CityOption represents a city with its coordinates (legacy)
type CityOption struct {
	City string  `json:"city"`
	Lat  float64 `json:"lat"`
	Lon  float64 `json:"lon"`
}

// WeatherAPIResponse represents the response from weather API
type WeatherAPIResponse struct {
	Temperature float64 `json:"temperature"`
	Location    string  `json:"location"`
	Description string  `json:"description,omitempty"`
	Humidity    int     `json:"humidity,omitempty"`
	WindSpeed   float64 `json:"windSpeed,omitempty"`
}

// LoadOutputFromRaw parses the RawOutput into the strongly typed Output field
func (step *ExecutionStep) LoadOutputFromRaw() error {
	if step.RawOutput == nil {
		return fmt.Errorf("no raw output to parse")
	}

	parsedOutput, err := ParseExecutionOutput(step.Type, step.RawOutput)
	if err != nil {
		return fmt.Errorf("failed to parse execution output: %w", err)
	}

	step.Output = parsedOutput
	return nil
}

// UpdateRawOutputFromOutput marshals the strongly typed Output into RawOutput
func (step *ExecutionStep) UpdateRawOutputFromOutput() error {
	if step.Output == nil {
		return nil // No output to marshal
	}

	rawOutput, err := json.Marshal(step.Output)
	if err != nil {
		return fmt.Errorf("failed to marshal execution output: %w", err)
	}

	step.RawOutput = rawOutput
	return nil
}
