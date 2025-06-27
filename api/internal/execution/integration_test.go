package execution

import (
	"context"
	"encoding/json"
	"testing"
)

func TestIntegrationService_ExecuteIntegration_Success(t *testing.T) {
	// Create mock API client
	mockClient := NewMockAPIClient()
	mockClient.SetDefaultWeatherResponse()

	// Create integration service
	service := NewIntegrationService(mockClient)

	// Test data from the workflow JSON
	nodeData := json.RawMessage(`{
		"metadata": {
			"apiEndpoint": "https://api.open-meteo.com/v1/forecast?latitude={lat}&longitude={lon}&current_weather=true",
			"inputVariables": ["city"],
			"outputVariables": ["temperature"],
			"options": [
				{"city": "Sydney", "lat": -33.8688, "lon": 151.2093},
				{"city": "Melbourne", "lat": -37.8136, "lon": 144.9631}
			]
		}
	}`)

	inputVariables := map[string]interface{}{
		"city": "Sydney",
	}

	// Execute integration
	result, err := service.ExecuteIntegration(context.Background(), nodeData, inputVariables)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify result structure
	temperature, ok := result["temperature"].(float64)
	if !ok {
		t.Error("Expected temperature to be float64")
	}
	if temperature != 28.5 {
		t.Errorf("Expected temperature 28.5, got %v", temperature)
	}

	location, ok := result["location"].(string)
	if !ok {
		t.Error("Expected location to be string")
	}
	if location != "Sydney" {
		t.Errorf("Expected location Sydney, got %v", location)
	}

	// Should include the raw API response
	if _, ok := result["apiResponse"]; !ok {
		t.Error("Expected apiResponse in result")
	}
}

func TestIntegrationService_ExecuteIntegration_InvalidCity(t *testing.T) {
	// Create integration service
	service := NewIntegrationService(NewMockAPIClient())

	// Test data with limited city options
	nodeData := json.RawMessage(`{
		"metadata": {
			"apiEndpoint": "https://api.open-meteo.com/v1/forecast?latitude={lat}&longitude={lon}&current_weather=true",
			"options": [
				{"city": "Sydney", "lat": -33.8688, "lon": 151.2093}
			]
		}
	}`)

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
	nodeData := json.RawMessage(`{
		"metadata": {
			"apiEndpoint": "https://api.open-meteo.com/v1/forecast?latitude={lat}&longitude={lon}&current_weather=true",
			"options": [
				{"city": "Sydney", "lat": -33.8688, "lon": 151.2093}
			]
		}
	}`)

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
