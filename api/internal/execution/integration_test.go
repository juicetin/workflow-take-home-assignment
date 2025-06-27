package execution

import (
	"context"
	"testing"

	"workflow-code-test/api/internal/models"
)

func TestIntegrationService_ExecuteIntegration_Success(t *testing.T) {
	// Create mock API client
	mockClient := NewMockAPIClient()
	mockClient.SetDefaultWeatherResponse()

	// Create integration service
	service := NewIntegrationService(mockClient)

	// Test data using strongly typed structures
	nodeData := models.IntegrationNodeData{
		Label:       "Weather API",
		Description: "Fetch weather data",
		Metadata: models.IntegrationNodeMetadata{
			HasHandles:      models.HandleConfig{Source: true, Target: true},
			InputVariables:  []string{"city"},
			APIEndpoint:     "https://api.open-meteo.com/v1/forecast?latitude={lat}&longitude={lon}&current_weather=true",
			OutputVariables: []string{"temperature"},
			Options: []models.LocationOption{
				{City: "Sydney", Lat: -33.8688, Lon: 151.2093},
				{City: "Melbourne", Lat: -37.8136, Lon: 144.9631},
			},
		},
	}

	inputVariables := map[string]interface{}{
		"city": "Sydney",
	}

	// Execute integration
	result, err := service.ExecuteIntegration(context.Background(), nodeData, inputVariables)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify strongly typed result structure
	if result.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", result.StatusCode)
	}

	temperature, ok := result.ProcessedData["temperature"].(float64)
	if !ok {
		t.Error("Expected temperature to be float64")
	}
	if temperature != 28.5 {
		t.Errorf("Expected temperature 28.5, got %v", temperature)
	}

	location, ok := result.ProcessedData["location"].(string)
	if !ok {
		t.Error("Expected location to be string")
	}
	if location != "Sydney" {
		t.Errorf("Expected location Sydney, got %v", location)
	}

	// Should include the raw API response
	if result.APIResponse == nil {
		t.Error("Expected APIResponse in result")
	}

	if result.EndpointCalled == "" {
		t.Error("Expected EndpointCalled to be set")
	}
}

func TestIntegrationService_ExecuteIntegration_InvalidCity(t *testing.T) {
	// Create integration service
	service := NewIntegrationService(NewMockAPIClient())

	// Test data with limited city options
	nodeData := models.IntegrationNodeData{
		Label:       "Weather API",
		Description: "Fetch weather data",
		Metadata: models.IntegrationNodeMetadata{
			HasHandles:  models.HandleConfig{Source: true, Target: true},
			APIEndpoint: "https://api.open-meteo.com/v1/forecast?latitude={lat}&longitude={lon}&current_weather=true",
			Options: []models.LocationOption{
				{City: "Sydney", Lat: -33.8688, Lon: 151.2093},
			},
		},
	}

	inputVariables := map[string]interface{}{
		"city": "Perth", // Not in available options
	}

	// Execute integration - should fail
	_, err := service.ExecuteIntegration(context.Background(), nodeData, inputVariables)
	if err == nil {
		t.Error("Expected error for invalid city, got none")
	}

	expectedMsg := "city 'Perth' not found in available options"
	if err.Error()[:len(expectedMsg)] != expectedMsg {
		t.Errorf("Expected error message to start with '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestIntegrationService_ExecuteIntegration_APIFailure(t *testing.T) {
	// Create mock API client that fails
	mockClient := NewMockAPIClient()
	mockClient.SetAPIError("network timeout")

	// Create integration service
	service := NewIntegrationService(mockClient)

	// Test data
	nodeData := models.IntegrationNodeData{
		Label:       "Weather API",
		Description: "Fetch weather data",
		Metadata: models.IntegrationNodeMetadata{
			HasHandles:  models.HandleConfig{Source: true, Target: true},
			APIEndpoint: "https://api.open-meteo.com/v1/forecast?latitude={lat}&longitude={lon}&current_weather=true",
			Options: []models.LocationOption{
				{City: "Sydney", Lat: -33.8688, Lon: 151.2093},
			},
		},
	}

	inputVariables := map[string]interface{}{
		"city": "Sydney",
	}

	// Execute integration - should fail
	_, err := service.ExecuteIntegration(context.Background(), nodeData, inputVariables)
	if err == nil {
		t.Error("Expected error for API failure, got none")
	}

	expectedMsg := "API call failed"
	if err.Error()[:len(expectedMsg)] != expectedMsg {
		t.Errorf("Expected error message to start with '%s', got '%s'", expectedMsg, err.Error())
	}
}
