package execution

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"workflow-code-test/api/internal/models"
)

// IntegrationService handles API integration calls for workflow nodes
type IntegrationService struct {
	apiClient APIClient
}

// NewIntegrationService creates a new integration service
func NewIntegrationService(apiClient APIClient) *IntegrationService {
	return &IntegrationService{
		apiClient: apiClient,
	}
}

// ExecuteIntegration executes an integration node based on its configuration
func (s *IntegrationService) ExecuteIntegration(ctx context.Context, nodeData json.RawMessage, inputVariables map[string]interface{}) (map[string]interface{}, error) {
	// Parse node configuration
	var integrationData models.IntegrationNodeData
	if err := json.Unmarshal(nodeData, &integrationData); err != nil {
		return nil, fmt.Errorf("failed to parse integration node data: %w", err)
	}

	// Get required input variable (city)
	cityValue, ok := inputVariables["city"]
	if !ok {
		return nil, fmt.Errorf("required input variable 'city' not found")
	}

	city, ok := cityValue.(string)
	if !ok {
		return nil, fmt.Errorf("city must be a string")
	}

	// Find coordinates for the city
	coordinates, err := s.findCityCoordinates(city, integrationData.Metadata.Options)
	if err != nil {
		return nil, err
	}

	// Build and call API
	apiURL := s.buildAPIURL(integrationData.Metadata.APIEndpoint, coordinates.Lat, coordinates.Lon)
	slog.Debug("Making integration API call", 
		"url", apiURL, 
		"city", city, 
		"lat", coordinates.Lat, 
		"lon", coordinates.Lon)

	apiResponse, err := s.apiClient.CallAPI(ctx, apiURL)
	if err != nil {
		return nil, fmt.Errorf("API call failed: %w", err)
	}

	// Extract temperature from response
	temperature, err := s.extractTemperature(apiResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to extract temperature: %w", err)
	}

	// Return structured response
	return map[string]interface{}{
		"temperature": temperature,
		"location":    coordinates.City,
		"apiResponse": apiResponse,
	}, nil
}

// findCityCoordinates finds the coordinates for a given city from available options
func (s *IntegrationService) findCityCoordinates(city string, options []models.CityOption) (*models.CityOption, error) {
	for _, option := range options {
		if strings.EqualFold(option.City, city) {
			return &option, nil
		}
	}
	return nil, fmt.Errorf("city '%s' not found in available options: %v", city, s.getCityNames(options))
}

// getCityNames extracts city names from options for error messages
func (s *IntegrationService) getCityNames(options []models.CityOption) []string {
	names := make([]string, len(options))
	for i, option := range options {
		names[i] = option.City
	}
	return names
}

// buildAPIURL builds the API URL by substituting coordinate placeholders
func (s *IntegrationService) buildAPIURL(template string, lat, lon float64) string {
	url := strings.ReplaceAll(template, "{lat}", fmt.Sprintf("%.6f", lat))
	url = strings.ReplaceAll(url, "{lon}", fmt.Sprintf("%.6f", lon))
	return url
}

// extractTemperature extracts temperature from Open-Meteo API response
// This could be made configurable for different API providers
func (s *IntegrationService) extractTemperature(apiResponse map[string]interface{}) (float64, error) {
	// Open-Meteo current weather response structure
	currentWeather, ok := apiResponse["current_weather"]
	if !ok {
		return 0, fmt.Errorf("current_weather not found in API response")
	}

	currentWeatherMap, ok := currentWeather.(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("current_weather is not an object")
	}

	temperature, ok := currentWeatherMap["temperature"]
	if !ok {
		return 0, fmt.Errorf("temperature not found in current_weather")
	}

	// Handle different numeric types
	switch temp := temperature.(type) {
	case float64:
		return temp, nil
	case float32:
		return float64(temp), nil
	case int:
		return float64(temp), nil
	case int64:
		return float64(temp), nil
	default:
		return 0, fmt.Errorf("temperature is not a numeric value: %T", temperature)
	}
}