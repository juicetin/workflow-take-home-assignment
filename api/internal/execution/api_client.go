package execution

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPAPIClient handles generic HTTP API calls
type HTTPAPIClient struct {
	httpClient *http.Client
}

// NewHTTPAPIClient creates a new HTTP API client
func NewHTTPAPIClient() *HTTPAPIClient {
	return &HTTPAPIClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// OpenMeteoCurrentWeatherResponse represents the current weather response from Open-Meteo
type OpenMeteoCurrentWeatherResponse struct {
	Current struct {
		Temperature float64 `json:"temperature"`
		Time        string  `json:"time"`
	} `json:"current_weather"`
}

// CallAPI makes a generic API call and returns the raw response
func (c *HTTPAPIClient) CallAPI(ctx context.Context, url string) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return result, nil
}
