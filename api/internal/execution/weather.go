package execution

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"workflow-code-test/api/internal/models"
)

// OpenMeteoService implements WeatherService using Open-Meteo API
type OpenMeteoService struct {
	httpClient *http.Client
}

// NewOpenMeteoService creates a new Open-Meteo weather service
func NewOpenMeteoService() *OpenMeteoService {
	return &OpenMeteoService{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GeocodingResponse represents the response from Open-Meteo geocoding API
type GeocodingResponse struct {
	Results []struct {
		Name      string  `json:"name"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Country   string  `json:"country"`
	} `json:"results"`
}

// WeatherResponse represents the response from Open-Meteo weather API
type WeatherResponse struct {
	Current struct {
		Temperature2m  float64 `json:"temperature_2m"`
		Humidity       int     `json:"relative_humidity_2m"`
		WindSpeed10m   float64 `json:"wind_speed_10m"`
		WeatherCode    int     `json:"weather_code"`
	} `json:"current"`
	CurrentUnits struct {
		Temperature2m string `json:"temperature_2m"`
		WindSpeed10m  string `json:"wind_speed_10m"`
	} `json:"current_units"`
}

// GetTemperature fetches temperature data for a given city using Open-Meteo API
func (s *OpenMeteoService) GetTemperature(ctx context.Context, city string) (*models.WeatherAPIResponse, error) {
	slog.Debug("Fetching temperature data from Open-Meteo", "city", city)
	
	// First, get coordinates for the city using geocoding
	coords, err := s.geocodeCity(ctx, city)
	if err != nil {
		return nil, fmt.Errorf("failed to geocode city: %w", err)
	}
	
	// Then, get weather data using coordinates
	weather, err := s.getWeatherData(ctx, coords.Latitude, coords.Longitude)
	if err != nil {
		return nil, fmt.Errorf("failed to get weather data: %w", err)
	}
	
	// Convert to our response format
	response := &models.WeatherAPIResponse{
		Temperature: weather.Current.Temperature2m,
		Location:    fmt.Sprintf("%s, %s", coords.Name, coords.Country),
		Description: s.getWeatherDescription(weather.Current.WeatherCode),
		Humidity:    weather.Current.Humidity,
		WindSpeed:   weather.Current.WindSpeed10m,
	}
	
	slog.Debug("Weather data fetched successfully", "city", city, "temperature", response.Temperature)
	
	return response, nil
}

// geocodeCity converts city name to coordinates using Open-Meteo geocoding API
func (s *OpenMeteoService) geocodeCity(ctx context.Context, city string) (*struct {
	Name      string
	Latitude  float64
	Longitude float64
	Country   string
}, error) {
	// Build geocoding URL
	params := url.Values{}
	params.Add("name", city)
	params.Add("count", "1")
	params.Add("language", "en")
	params.Add("format", "json")
	
	geocodingURL := fmt.Sprintf("https://geocoding-api.open-meteo.com/v1/search?%s", params.Encode())
	
	// Make geocoding request
	req, err := http.NewRequestWithContext(ctx, "GET", geocodingURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create geocoding request: %w", err)
	}
	
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make geocoding request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("geocoding request failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	var geoResp GeocodingResponse
	if err := json.NewDecoder(resp.Body).Decode(&geoResp); err != nil {
		return nil, fmt.Errorf("failed to decode geocoding response: %w", err)
	}
	
	if len(geoResp.Results) == 0 {
		return nil, fmt.Errorf("city not found: %s", city)
	}
	
	result := geoResp.Results[0]
	return &struct {
		Name      string
		Latitude  float64
		Longitude float64
		Country   string
	}{
		Name:      result.Name,
		Latitude:  result.Latitude,
		Longitude: result.Longitude,
		Country:   result.Country,
	}, nil
}

// getWeatherData fetches current weather data using coordinates
func (s *OpenMeteoService) getWeatherData(ctx context.Context, lat, lon float64) (*WeatherResponse, error) {
	// Build weather API URL
	params := url.Values{}
	params.Add("latitude", fmt.Sprintf("%.6f", lat))
	params.Add("longitude", fmt.Sprintf("%.6f", lon))
	params.Add("current", "temperature_2m,relative_humidity_2m,wind_speed_10m,weather_code")
	params.Add("timezone", "auto")
	
	weatherURL := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?%s", params.Encode())
	
	// Make weather request
	req, err := http.NewRequestWithContext(ctx, "GET", weatherURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create weather request: %w", err)
	}
	
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make weather request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("weather request failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	var weatherResp WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherResp); err != nil {
		return nil, fmt.Errorf("failed to decode weather response: %w", err)
	}
	
	return &weatherResp, nil
}

// getWeatherDescription converts WMO weather code to description
func (s *OpenMeteoService) getWeatherDescription(code int) string {
	// WMO weather interpretation codes
	descriptions := map[int]string{
		0:  "clear sky",
		1:  "mainly clear",
		2:  "partly cloudy",
		3:  "overcast",
		45: "fog",
		48: "depositing rime fog",
		51: "light drizzle",
		53: "moderate drizzle",
		55: "dense drizzle",
		56: "light freezing drizzle",
		57: "dense freezing drizzle",
		61: "slight rain",
		63: "moderate rain",
		65: "heavy rain",
		66: "light freezing rain",
		67: "heavy freezing rain",
		71: "slight snow fall",
		73: "moderate snow fall",
		75: "heavy snow fall",
		77: "snow grains",
		80: "slight rain showers",
		81: "moderate rain showers",
		82: "violent rain showers",
		85: "slight snow showers",
		86: "heavy snow showers",
		95: "thunderstorm",
		96: "thunderstorm with slight hail",
		99: "thunderstorm with heavy hail",
	}
	
	if desc, exists := descriptions[code]; exists {
		return desc
	}
	return "unknown weather"
}

// MockWeatherService provides a mock implementation for testing
type MockWeatherService struct {
	cityTemperatures map[string]float64
}

// NewMockWeatherService creates a new mock weather service
func NewMockWeatherService() *MockWeatherService {
	return &MockWeatherService{
		cityTemperatures: map[string]float64{
			"Sydney":    28.5,
			"Melbourne": 22.1,
			"Brisbane":  30.2,
			"Perth":     25.8,
			"Adelaide":  24.3,
			"Canberra":  19.7,
			"Darwin":    32.1,
			"Hobart":    18.4,
			"London":    15.2,
			"New York":  23.4,
			"Tokyo":     26.8,
			"Paris":     18.9,
			"Berlin":    16.3,
			"Rome":      24.7,
			"Madrid":    27.1,
			"Moscow":    12.8,
			"Beijing":   21.5,
			"Mumbai":    31.2,
			"Cairo":     29.6,
			"Toronto":   20.3,
		},
	}
}

// GetTemperature returns mock temperature data for testing
func (m *MockWeatherService) GetTemperature(ctx context.Context, city string) (*models.WeatherAPIResponse, error) {
	slog.Debug("Mock: Fetching temperature data", "city", city)
	
	// Simulate API delay
	time.Sleep(100 * time.Millisecond)
	
	// Look up temperature for the city
	temperature, exists := m.cityTemperatures[city]
	if !exists {
		// Return a default temperature for unknown cities
		temperature = 20.0
		slog.Warn("Unknown city, using default temperature", "city", city, "temperature", temperature)
	}
	
	response := &models.WeatherAPIResponse{
		Temperature: temperature,
		Location:    city,
		Description: "partly cloudy",
		Humidity:    65,
		WindSpeed:   5.2,
	}
	
	slog.Debug("Mock weather data returned", "city", city, "temperature", temperature)
	
	return response, nil
}

// DefaultWeatherService automatically chooses between real and mock service
type DefaultWeatherService struct {
	realService *OpenMeteoService
	mockService *MockWeatherService
	useMock     bool
}

// NewDefaultWeatherService creates a weather service that uses Open-Meteo by default, with mock fallback
func NewDefaultWeatherService(useMock bool) *DefaultWeatherService {
	return &DefaultWeatherService{
		realService: NewOpenMeteoService(),
		mockService: NewMockWeatherService(),
		useMock:     useMock,
	}
}

// GetTemperature chooses between real and mock service
func (d *DefaultWeatherService) GetTemperature(ctx context.Context, city string) (*models.WeatherAPIResponse, error) {
	if d.useMock {
		slog.Debug("Using mock weather service")
		return d.mockService.GetTemperature(ctx, city)
	}
	
	slog.Debug("Using Open-Meteo weather API")
	return d.realService.GetTemperature(ctx, city)
}