// Package calculator provides GPS distance calculations using the Haversine formula
// to compute great-circle distances between geographic coordinates.
package calculator

import (
	"math"
)

const (
	// EarthRadiusKM is the Earth's radius in kilometers
	EarthRadiusKM = 6371.0
)

// Location represents a GPS coordinate
type Location struct {
	Latitude  float64
	Longitude float64
}

// DistanceFromHome calculates the great-circle distance between a location
// and the home coordinates using the Haversine formula
func DistanceFromHome(homeLat, homeLon, lat, lon float64) float64 {
	return Haversine(homeLat, homeLon, lat, lon)
}

// Haversine calculates the great-circle distance between two points
// on the Earth's surface given their latitudes and longitudes in decimal degrees
//
// Formula:
// a = sin²(Δφ/2) + cos φ1 ⋅ cos φ2 ⋅ sin²(Δλ/2)
// c = 2 ⋅ atan2( √a, √(1−a) )
// d = R ⋅ c
//
// where:
// φ is latitude, λ is longitude, R is earth's radius (6371 km)
// Δφ is the difference in latitude, Δλ is the difference in longitude
func Haversine(lat1, lon1, lat2, lon2 float64) float64 {
	// Convert latitude and longitude from degrees to radians
	lat1Rad := degreesToRadians(lat1)
	lon1Rad := degreesToRadians(lon1)
	lat2Rad := degreesToRadians(lat2)
	lon2Rad := degreesToRadians(lon2)

	// Calculate differences
	deltaLat := lat2Rad - lat1Rad
	deltaLon := lon2Rad - lon1Rad

	// Apply Haversine formula
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	// Distance in kilometers
	return EarthRadiusKM * c
}

// degreesToRadians converts degrees to radians
func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

// DistanceMetrics holds calculated distance statistics
type DistanceMetrics struct {
	TotalDistanceKM float64
	MaxDistanceKM   float64
	MinDistanceKM   float64
	TotalLocations  int
	AvgDistanceKM   float64
}

// CalculateMetrics computes distance metrics for a set of locations
func CalculateMetrics(homeLat, homeLon float64, locations []Location) DistanceMetrics {
	if len(locations) == 0 {
		return DistanceMetrics{}
	}

	metrics := DistanceMetrics{
		TotalLocations: len(locations),
		MinDistanceKM:  math.MaxFloat64,
	}

	var totalDistance float64

	for _, loc := range locations {
		distance := DistanceFromHome(homeLat, homeLon, loc.Latitude, loc.Longitude)
		totalDistance += distance

		if distance > metrics.MaxDistanceKM {
			metrics.MaxDistanceKM = distance
		}
		if distance < metrics.MinDistanceKM {
			metrics.MinDistanceKM = distance
		}
	}

	metrics.TotalDistanceKM = totalDistance
	metrics.AvgDistanceKM = totalDistance / float64(len(locations))

	// Handle case where no locations were processed
	if metrics.MinDistanceKM == math.MaxFloat64 {
		metrics.MinDistanceKM = 0
	}

	return metrics
}
