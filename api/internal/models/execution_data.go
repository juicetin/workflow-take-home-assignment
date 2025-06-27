package models

import (
	"encoding/json"
	"fmt"
)

// ExecutionOutput is the interface that all execution output types must implement
type ExecutionOutput interface {
	// GetOutputType returns the type of output this represents
	GetOutputType() string
	// Validate performs validation specific to this output type
	Validate() error
}

// StartExecutionOutput represents output from start node execution
type StartExecutionOutput struct {
	Message   string                 `json:"message"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

func (o StartExecutionOutput) GetOutputType() string { return NodeTypeStart }
func (o StartExecutionOutput) Validate() error       { return nil }

// FormExecutionOutput represents output from form node execution
type FormExecutionOutput struct {
	CollectedData map[string]interface{} `json:"collectedData"`
	FormFields    []string               `json:"formFields"`
}

func (o FormExecutionOutput) GetOutputType() string { return NodeTypeForm }
func (o FormExecutionOutput) Validate() error {
	if len(o.CollectedData) == 0 {
		return fmt.Errorf("form execution must collect at least one data field")
	}
	return nil
}

// IntegrationExecutionOutput represents output from integration node execution
type IntegrationExecutionOutput struct {
	APIResponse    map[string]interface{} `json:"apiResponse"`
	ProcessedData  map[string]interface{} `json:"processedData"`
	EndpointCalled string                 `json:"endpointCalled"`
	StatusCode     int                    `json:"statusCode"`
}

func (o IntegrationExecutionOutput) GetOutputType() string { return NodeTypeIntegration }
func (o IntegrationExecutionOutput) Validate() error {
	if o.EndpointCalled == "" {
		return fmt.Errorf("integration execution must specify endpoint called")
	}
	if o.StatusCode == 0 {
		return fmt.Errorf("integration execution must specify status code")
	}
	return nil
}

// ConditionExecutionOutput represents output from condition node execution
type ConditionExecutionOutput struct {
	ConditionMet        bool                   `json:"conditionMet"`
	EvaluatedExpression string                 `json:"evaluatedExpression"`
	NextPath            string                 `json:"nextPath"` // "true" or "false"
	Variables           map[string]interface{} `json:"variables,omitempty"`
}

func (o ConditionExecutionOutput) GetOutputType() string { return NodeTypeCondition }
func (o ConditionExecutionOutput) Validate() error {
	if o.NextPath != "true" && o.NextPath != "false" {
		return fmt.Errorf("condition execution next path must be 'true' or 'false', got: %s", o.NextPath)
	}
	return nil
}

// EmailExecutionOutput represents output from email node execution
type EmailExecutionOutput struct {
	EmailSent     bool   `json:"emailSent"`
	RecipientInfo string `json:"recipientInfo"`
	Subject       string `json:"subject"`
	MessageID     string `json:"messageId,omitempty"`
	Error         string `json:"error,omitempty"`
}

func (o EmailExecutionOutput) GetOutputType() string { return NodeTypeEmail }
func (o EmailExecutionOutput) Validate() error {
	if o.Subject == "" {
		return fmt.Errorf("email execution must specify subject")
	}
	return nil
}

// EndExecutionOutput represents output from end node execution
type EndExecutionOutput struct {
	Message       string                 `json:"message"`
	FinalState    map[string]interface{} `json:"finalState"`
	ExecutionTime int64                  `json:"executionTime,omitempty"`
}

func (o EndExecutionOutput) GetOutputType() string { return NodeTypeEnd }
func (o EndExecutionOutput) Validate() error       { return nil }

// ExecutionOutputUnion represents a union type for all possible execution outputs
type ExecutionOutputUnion struct {
	Type   string `json:"-"` // Set during unmarshaling
	Output ExecutionOutput
}

// UnmarshalJSON implements custom unmarshaling for ExecutionOutputUnion
func (eou *ExecutionOutputUnion) UnmarshalJSON(data []byte) error {
	// Try each type until one works
	types := []struct {
		outputType string
		target     ExecutionOutput
	}{
		{NodeTypeStart, &StartExecutionOutput{}},
		{NodeTypeForm, &FormExecutionOutput{}},
		{NodeTypeIntegration, &IntegrationExecutionOutput{}},
		{NodeTypeCondition, &ConditionExecutionOutput{}},
		{NodeTypeEmail, &EmailExecutionOutput{}},
		{NodeTypeEnd, &EndExecutionOutput{}},
	}

	var lastErr error
	for _, t := range types {
		if err := json.Unmarshal(data, t.target); err == nil {
			// Validate that this is the correct type
			if err := t.target.Validate(); err == nil {
				eou.Type = t.outputType
				eou.Output = t.target
				return nil
			}
		} else {
			lastErr = err
		}
	}

	return fmt.Errorf("failed to unmarshal execution output into any known type: %v", lastErr)
}

// MarshalJSON implements custom marshaling for ExecutionOutputUnion
func (eou ExecutionOutputUnion) MarshalJSON() ([]byte, error) {
	if eou.Output == nil {
		return []byte("null"), nil
	}
	return json.Marshal(eou.Output)
}

// ParseExecutionOutput parses raw JSON into the appropriate strongly typed ExecutionOutput
func ParseExecutionOutput(nodeType string, rawData []byte) (ExecutionOutput, error) {
	switch nodeType {
	case NodeTypeStart:
		var output StartExecutionOutput
		if err := json.Unmarshal(rawData, &output); err != nil {
			return nil, fmt.Errorf("failed to parse start execution output: %w", err)
		}
		return output, output.Validate()

	case NodeTypeForm:
		var output FormExecutionOutput
		if err := json.Unmarshal(rawData, &output); err != nil {
			return nil, fmt.Errorf("failed to parse form execution output: %w", err)
		}
		return output, output.Validate()

	case NodeTypeIntegration:
		var output IntegrationExecutionOutput
		if err := json.Unmarshal(rawData, &output); err != nil {
			return nil, fmt.Errorf("failed to parse integration execution output: %w", err)
		}
		return output, output.Validate()

	case NodeTypeCondition:
		var output ConditionExecutionOutput
		if err := json.Unmarshal(rawData, &output); err != nil {
			return nil, fmt.Errorf("failed to parse condition execution output: %w", err)
		}
		return output, output.Validate()

	case NodeTypeEmail:
		var output EmailExecutionOutput
		if err := json.Unmarshal(rawData, &output); err != nil {
			return nil, fmt.Errorf("failed to parse email execution output: %w", err)
		}
		return output, output.Validate()

	case NodeTypeEnd:
		var output EndExecutionOutput
		if err := json.Unmarshal(rawData, &output); err != nil {
			return nil, fmt.Errorf("failed to parse end execution output: %w", err)
		}
		return output, output.Validate()

	default:
		return nil, fmt.Errorf("unknown node type for execution output: %s", nodeType)
	}
}
