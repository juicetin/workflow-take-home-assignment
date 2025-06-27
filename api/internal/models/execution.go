package models

import (
	"encoding/json"
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
	Output      json.RawMessage `json:"output,omitempty"`
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

// FormNodeData represents the data structure for form nodes
type FormNodeData struct {
	Fields []FormField `json:"fields"`
}

// FormField represents a single form field configuration
type FormField struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Label       string   `json:"label"`
	Required    bool     `json:"required"`
	Options     []string `json:"options,omitempty"`    // for select fields
	Validation  string   `json:"validation,omitempty"` // validation rules
	Placeholder string   `json:"placeholder,omitempty"`
}

// IntegrationNodeData represents the data structure for integration nodes
type IntegrationNodeData struct {
	Metadata IntegrationMetadata `json:"metadata"`
}

// IntegrationMetadata contains the integration configuration
type IntegrationMetadata struct {
	APIEndpoint     string       `json:"apiEndpoint"`
	InputVariables  []string     `json:"inputVariables"`
	OutputVariables []string     `json:"outputVariables"`
	Options         []CityOption `json:"options"`
}

// CityOption represents a city with its coordinates
type CityOption struct {
	City string  `json:"city"`
	Lat  float64 `json:"lat"`
	Lon  float64 `json:"lon"`
}

// ConditionNodeData represents the data structure for condition nodes
type ConditionNodeData struct {
	Field     string      `json:"field"`
	Operator  string      `json:"operator"`
	Value     interface{} `json:"value"`
	TruePath  string      `json:"truePath,omitempty"`
	FalsePath string      `json:"falsePath,omitempty"`
}

// EmailNodeData represents the data structure for email nodes
type EmailNodeData struct {
	ToField      string `json:"toField"`
	Subject      string `json:"subject"`
	BodyTemplate string `json:"bodyTemplate"`
	FromEmail    string `json:"fromEmail"`
}

// WeatherAPIResponse represents the response from weather API
type WeatherAPIResponse struct {
	Temperature float64 `json:"temperature"`
	Location    string  `json:"location"`
	Description string  `json:"description,omitempty"`
	Humidity    int     `json:"humidity,omitempty"`
	WindSpeed   float64 `json:"windSpeed,omitempty"`
}
