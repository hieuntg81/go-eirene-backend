package entity

import (
	"encoding/json"
	"math"
)

// GeoPoint represents a geographic point with latitude and longitude
type GeoPoint struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// NewGeoPoint creates a new GeoPoint
func NewGeoPoint(lat, lng float64) *GeoPoint {
	return &GeoPoint{
		Latitude:  lat,
		Longitude: lng,
	}
}

// IsZero returns true if the GeoPoint is at the origin
func (g *GeoPoint) IsZero() bool {
	return g == nil || (g.Latitude == 0 && g.Longitude == 0)
}

// MarshalJSON implements json.Marshaler
func (g GeoPoint) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}{
		Latitude:  g.Latitude,
		Longitude: g.Longitude,
	})
}

// UnmarshalJSON implements json.Unmarshaler
func (g *GeoPoint) UnmarshalJSON(data []byte) error {
	var aux struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	g.Latitude = aux.Latitude
	g.Longitude = aux.Longitude
	return nil
}

// DistanceKm calculates the distance in kilometers between two points using Haversine formula
func (g *GeoPoint) DistanceKm(other *GeoPoint) float64 {
	if g == nil || other == nil {
		return 0
	}

	const earthRadiusKm = 6371.0

	lat1Rad := g.Latitude * (math.Pi / 180)
	lat2Rad := other.Latitude * (math.Pi / 180)
	deltaLat := (other.Latitude - g.Latitude) * (math.Pi / 180)
	deltaLng := (other.Longitude - g.Longitude) * (math.Pi / 180)

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(deltaLng/2)*math.Sin(deltaLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}

// DistanceKmBetween calculates the distance in kilometers between two coordinates
func DistanceKmBetween(lat1, lng1, lat2, lng2 float64) float64 {
	p1 := NewGeoPoint(lat1, lng1)
	p2 := NewGeoPoint(lat2, lng2)
	return p1.DistanceKm(p2)
}

// BoundingBox represents a geographic bounding box for filtering
type BoundingBox struct {
	MinLat float64
	MaxLat float64
	MinLng float64
	MaxLng float64
}

// NewBoundingBox creates a bounding box around a center point with a given radius in km
func NewBoundingBox(lat, lng float64, radiusKm float64) *BoundingBox {
	// Approximate degrees per km at the equator
	// 1 degree latitude = ~111 km
	// 1 degree longitude = ~111 km * cos(latitude)
	const kmPerDegLat = 111.0

	latDelta := radiusKm / kmPerDegLat
	lngDelta := radiusKm / (kmPerDegLat * math.Cos(lat*math.Pi/180))

	return &BoundingBox{
		MinLat: lat - latDelta,
		MaxLat: lat + latDelta,
		MinLng: lng - lngDelta,
		MaxLng: lng + lngDelta,
	}
}

// Contains checks if a point is within the bounding box
func (bb *BoundingBox) Contains(lat, lng float64) bool {
	return lat >= bb.MinLat && lat <= bb.MaxLat &&
		lng >= bb.MinLng && lng <= bb.MaxLng
}
