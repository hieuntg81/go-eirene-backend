package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"bamboo-rescue/internal/config"
	"bamboo-rescue/internal/domain/entity"
	"go.uber.org/zap"
)

// GeocodeService defines the interface for geocoding operations
type GeocodeService interface {
	ReverseGeocode(ctx context.Context, lat, lng float64) (*GeocodeResult, error)
	SearchAddress(ctx context.Context, query string) ([]GeocodeResult, error)
}

// GeocodeResult represents a geocoding result
type GeocodeResult struct {
	Address   string          `json:"address"`
	Location  entity.GeoPoint `json:"location"`
	PlaceID   string          `json:"place_id"`
	PlaceType string          `json:"place_type"`
}

type geocodeService struct {
	nominatimURL string
	httpClient   *http.Client
	log          *zap.Logger
}

// NewGeocodeService creates a new GeocodeService
func NewGeocodeService(cfg *config.Config, log *zap.Logger) GeocodeService {
	nominatimURL := cfg.Nominatim.URL
	if nominatimURL == "" {
		nominatimURL = "https://nominatim.openstreetmap.org"
	}

	return &geocodeService{
		nominatimURL: nominatimURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		log: log,
	}
}

// NominatimReverseResponse represents Nominatim reverse geocoding response
type NominatimReverseResponse struct {
	PlaceID     int     `json:"place_id"`
	Lat         string  `json:"lat"`
	Lon         string  `json:"lon"`
	DisplayName string  `json:"display_name"`
	Type        string  `json:"type"`
	Address     Address `json:"address"`
}

// Address represents address components from Nominatim
type Address struct {
	Road        string `json:"road"`
	Suburb      string `json:"suburb"`
	City        string `json:"city"`
	State       string `json:"state"`
	Postcode    string `json:"postcode"`
	Country     string `json:"country"`
	CountryCode string `json:"country_code"`
}

// NominatimSearchResponse represents Nominatim search response
type NominatimSearchResponse struct {
	PlaceID     int    `json:"place_id"`
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	DisplayName string `json:"display_name"`
	Type        string `json:"type"`
}

func (s *geocodeService) ReverseGeocode(ctx context.Context, lat, lng float64) (*GeocodeResult, error) {
	reqURL := fmt.Sprintf("%s/reverse?format=json&lat=%f&lon=%f&addressdetails=1",
		s.nominatimURL, lat, lng)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "RescueApp/1.0")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.log.Error("Failed to call Nominatim reverse API", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("nominatim returned status %d", resp.StatusCode)
	}

	var result NominatimReverseResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		s.log.Error("Failed to decode Nominatim response", zap.Error(err))
		return nil, err
	}

	return &GeocodeResult{
		Address: result.DisplayName,
		Location: entity.GeoPoint{
			Latitude:  lat,
			Longitude: lng,
		},
		PlaceID:   fmt.Sprintf("%d", result.PlaceID),
		PlaceType: result.Type,
	}, nil
}

func (s *geocodeService) SearchAddress(ctx context.Context, query string) ([]GeocodeResult, error) {
	reqURL := fmt.Sprintf("%s/search?format=json&q=%s&limit=10",
		s.nominatimURL, url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "RescueApp/1.0")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.log.Error("Failed to call Nominatim search API", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("nominatim returned status %d", resp.StatusCode)
	}

	var results []NominatimSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		s.log.Error("Failed to decode Nominatim response", zap.Error(err))
		return nil, err
	}

	geocodeResults := make([]GeocodeResult, len(results))
	for i, r := range results {
		var lat, lng float64
		fmt.Sscanf(r.Lat, "%f", &lat)
		fmt.Sscanf(r.Lon, "%f", &lng)

		geocodeResults[i] = GeocodeResult{
			Address: r.DisplayName,
			Location: entity.GeoPoint{
				Latitude:  lat,
				Longitude: lng,
			},
			PlaceID:   fmt.Sprintf("%d", r.PlaceID),
			PlaceType: r.Type,
		}
	}

	return geocodeResults, nil
}
