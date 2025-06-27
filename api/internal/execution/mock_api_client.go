package execution

import (
	"context"
	"fmt"
	"strings"
)

// MockAPIClient provides a mock implementation for testing
type MockAPIClient struct {
	responses map[string]map[string]interface{}
	errors    map[string]error
}

// NewMockAPIClient creates a new mock API client
func NewMockAPIClient() *MockAPIClient {
	return &MockAPIClient{
		responses: make(map[string]map[string]interface{}),
		errors:    make(map[string]error),
	}
}

// SetResponse sets a mock response for a given URL pattern
func (m *MockAPIClient) SetResponse(urlPattern string, response map[string]interface{}) {
	m.responses[urlPattern] = response
}

// SetError sets an error response for a given URL pattern
func (m *MockAPIClient) SetError(urlPattern string, err error) {
	m.errors[urlPattern] = err
}

// CallAPI returns a mock response based on URL patterns
func (m *MockAPIClient) CallAPI(ctx context.Context, url string) (map[string]interface{}, error) {
	// Check for exact matches first
	if response, ok := m.responses[url]; ok {
		return response, nil
	}
	if err, ok := m.errors[url]; ok {
		return nil, err
	}

	// Check for pattern matches
	for pattern, response := range m.responses {
		if strings.Contains(url, pattern) {
			return response, nil
		}
	}

	for pattern, err := range m.errors {
		if strings.Contains(url, pattern) {
			return nil, err
		}
	}

	// Default response for unknown URLs
	return map[string]interface{}{
		"current_weather": map[string]interface{}{
			"temperature": 25.0,
			"time":        "2024-01-01T12:00",
		},
	}, nil
}

// SetDefaultWeatherResponse sets up typical weather API responses for testing
func (m *MockAPIClient) SetDefaultWeatherResponse() {
	sydneyResponse := map[string]interface{}{
		"current_weather": map[string]interface{}{
			"temperature": 28.5,
			"time":        "2024-01-01T12:00",
		},
	}

	melbourneResponse := map[string]interface{}{
		"current_weather": map[string]interface{}{
			"temperature": 22.1,
			"time":        "2024-01-01T12:00",
		},
	}

	// Set responses for Sydney coordinates
	m.SetResponse("latitude=-33.868800", sydneyResponse)

	// Set responses for Melbourne coordinates
	m.SetResponse("latitude=-37.813600", melbourneResponse)
}

// SetAPIError sets up an API error for testing error scenarios
func (m *MockAPIClient) SetAPIError(message string) {
	m.SetError("api.open-meteo.com", fmt.Errorf("API error: %s", message))
}
