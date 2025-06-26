package execution

import (
	"context"
	"encoding/json"
	"testing"

	"workflow-code-test/api/internal/models"
)

func TestEngine_ExecuteWorkflow(t *testing.T) {
	// Create services
	weatherService := NewMockWeatherService()
	emailService := NewInMemoryEmailService()
	validator := NewDefaultInputValidator()
	
	// Create engine
	engine := NewEngine(weatherService, emailService, validator)
	
	// Create test workflow
	workflow := &models.WorkflowResponse{
		ID: "test-workflow",
		Nodes: []models.NodeResponse{
			{
				ID:   "start",
				Type: models.NodeTypeStart,
				Data: json.RawMessage(`{}`),
			},
			{
				ID:   "form",
				Type: models.NodeTypeForm,
				Data: json.RawMessage(`{}`),
			},
			{
				ID:   "weather",
				Type: models.NodeTypeIntegration,
				Data: json.RawMessage(`{}`),
			},
			{
				ID:   "condition",
				Type: models.NodeTypeCondition,
				Data: json.RawMessage(`{}`),
			},
			{
				ID:   "email",
				Type: models.NodeTypeEmail,
				Data: json.RawMessage(`{}`),
			},
			{
				ID:   "end",
				Type: models.NodeTypeEnd,
				Data: json.RawMessage(`{}`),
			},
		},
		Edges: []models.EdgeResponse{
			{ID: "e1", Source: "start", Target: "form"},
			{ID: "e2", Source: "form", Target: "weather"},
			{ID: "e3", Source: "weather", Target: "condition"},
			{ID: "e4", Source: "condition", Target: "email", SourceHandle: stringPtr("true")},
			{ID: "e5", Source: "email", Target: "end"},
		},
	}
	
	// Create execution request
	req := &models.ExecutionRequest{
		FormData: map[string]interface{}{
			"name":  "Alice",
			"email": "alice@example.com",
			"city":  "Sydney",
		},
		Condition: map[string]interface{}{
			"operator":  "greater_than",
			"threshold": 25.0,
		},
	}
	
	// Execute workflow
	result, err := engine.ExecuteWorkflow(context.Background(), workflow, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Verify result
	if result.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", result.Status)
	}
	
	if len(result.Steps) == 0 {
		t.Error("Expected execution steps, got none")
	}
	
	// Verify email was tracked
	sentEmails := emailService.GetSentEmails()
	if len(sentEmails) != 1 {
		t.Errorf("Expected 1 email to be tracked, got %d", len(sentEmails))
	}
	
	if len(sentEmails) > 0 {
		email := sentEmails[0]
		if email.To != "alice@example.com" {
			t.Errorf("Expected email to 'alice@example.com', got '%s'", email.To)
		}
	}
}

func TestDefaultInputValidator_ValidateFormData(t *testing.T) {
	validator := NewDefaultInputValidator()
	
	// Test with no validation rules (should pass)
	err := validator.ValidateFormData(
		map[string]interface{}{
			"name": "Alice",
			"email": "alice@example.com",
		},
		nil,
	)
	if err != nil {
		t.Errorf("Expected no error with nil nodeData, got %v", err)
	}
	
	// Test with field definitions
	formData := &models.FormNodeData{
		Fields: []models.FormField{
			{
				Name:     "name",
				Type:     "text",
				Required: true,
			},
			{
				Name:     "email",
				Type:     "email",
				Required: true,
			},
		},
	}
	
	// Valid data should pass
	err = validator.ValidateFormData(
		map[string]interface{}{
			"name":  "Alice",
			"email": "alice@example.com",
		},
		formData,
	)
	if err != nil {
		t.Errorf("Expected no error with valid data, got %v", err)
	}
	
	// Missing required field should fail
	err = validator.ValidateFormData(
		map[string]interface{}{
			"name": "Alice",
			// missing email
		},
		formData,
	)
	if err == nil {
		t.Error("Expected error for missing required field, got none")
	}
}

