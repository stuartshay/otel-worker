package calculator

import (
	"math"
	"testing"
)

func TestHaversine(t *testing.T) {
	tests := []struct {
		name     string
		lat1     float64
		lon1     float64
		lat2     float64
		lon2     float64
		expected float64
		tolerance float64
	}{
		{
			name:      "Same location",
			lat1:      40.736097,
			lon1:      -74.039373,
			lat2:      40.736097,
			lon2:      -74.039373,
			expected:  0.0,
			tolerance: 0.001,
		},
		{
			name:      "New York to Jersey City (~3.3 km)",
			lat1:      40.736097,
			lon1:      -74.039373,
			lat2:      40.728333,
			lon2:      -74.077778,
			expected:  3.35,
			tolerance: 0.5,
		},
		{
			name:      "New York to Boston (~306 km)",
			lat1:      40.7128,
			lon1:      -74.0060,
			lat2:      42.3601,
			lon2:      -71.0589,
			expected:  306.0,
			tolerance: 5.0,
		},
		{
			name:      "Equator crossing",
			lat1:      1.0,
			lon1:      0.0,
			lat2:      -1.0,
			lon2:      0.0,
			expected:  222.4,
			tolerance: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Haversine(tt.lat1, tt.lon1, tt.lat2, tt.lon2)
			if math.Abs(result-tt.expected) > tt.tolerance {
				t.Errorf("Haversine() = %.2f km, expected %.2f km (±%.2f km)", result, tt.expected, tt.tolerance)
			}
		})
	}
}

func TestDistanceFromHome(t *testing.T) {
	homeLat := 40.736097
	homeLon := -74.039373

	tests := []struct {
		name      string
		lat       float64
		lon       float64
		threshold float64
	}{
		{
			name:      "At home",
			lat:       40.736097,
			lon:       -74.039373,
			threshold: 0.001,
		},
		{
			name:      "Near home (100m)",
			lat:       40.737,
			lon:       -74.039,
			threshold: 0.2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			distance := DistanceFromHome(homeLat, homeLon, tt.lat, tt.lon)
			if distance > tt.threshold {
				t.Errorf("DistanceFromHome() = %.4f km, expected < %.4f km", distance, tt.threshold)
			}
		})
	}
}

func TestCalculateMetrics(t *testing.T) {
	homeLat := 40.736097
	homeLon := -74.039373

	t.Run("empty locations", func(t *testing.T) {
		metrics := CalculateMetrics(homeLat, homeLon, []Location{})
		if metrics.TotalLocations != 0 {
			t.Errorf("expected TotalLocations 0, got %d", metrics.TotalLocations)
		}
		if metrics.TotalDistanceKM != 0 {
			t.Errorf("expected TotalDistanceKM 0, got %.2f", metrics.TotalDistanceKM)
		}
	})

	t.Run("single location at home", func(t *testing.T) {
		locations := []Location{
			{Latitude: homeLat, Longitude: homeLon},
		}
		metrics := CalculateMetrics(homeLat, homeLon, locations)
		if metrics.TotalLocations != 1 {
			t.Errorf("expected TotalLocations 1, got %d", metrics.TotalLocations)
		}
		if metrics.MaxDistanceKM > 0.001 {
			t.Errorf("expected MaxDistanceKM ~0, got %.4f", metrics.MaxDistanceKM)
		}
	})

	t.Run("multiple locations", func(t *testing.T) {
		locations := []Location{
			{Latitude: homeLat, Longitude: homeLon},            // At home
			{Latitude: 40.748817, Longitude: -73.985428},       // Empire State Building (~5 km)
			{Latitude: 40.758896, Longitude: -73.985130},       // Times Square (~7 km)
		}
		metrics := CalculateMetrics(homeLat, homeLon, locations)

		if metrics.TotalLocations != 3 {
			t.Errorf("expected TotalLocations 3, got %d", metrics.TotalLocations)
		}
		if metrics.MinDistanceKM > 0.001 {
			t.Errorf("expected MinDistanceKM ~0, got %.4f", metrics.MinDistanceKM)
		}
		if metrics.MaxDistanceKM < 5.0 || metrics.MaxDistanceKM > 10.0 {
			t.Errorf("expected MaxDistanceKM between 5-10 km, got %.2f", metrics.MaxDistanceKM)
		}
		if metrics.AvgDistanceKM < 0 {
			t.Errorf("expected AvgDistanceKM > 0, got %.2f", metrics.AvgDistanceKM)
		}
	})
}

func TestDegreesToRadians(t *testing.T) {
	tests := []struct {
		degrees  float64
		expected float64
	}{
		{0, 0},
		{90, math.Pi / 2},
		{180, math.Pi},
		{360, 2 * math.Pi},
		{-90, -math.Pi / 2},
	}

	for _, tt := range tests {
		result := degreesToRadians(tt.degrees)
		if math.Abs(result-tt.expected) > 0.0001 {
			t.Errorf("degreesToRadians(%.2f) = %.4f, expected %.4f", tt.degrees, result, tt.expected)
		}
	}
}

func BenchmarkHaversine(b *testing.B) {
	lat1, lon1 := 40.736097, -74.039373
	lat2, lon2 := 40.748817, -73.985428

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Haversine(lat1, lon1, lat2, lon2)
	}
}

func BenchmarkCalculateMetrics(b *testing.B) {
	homeLat, homeLon := 40.736097, -74.039373
	locations := make([]Location, 1000)
	for i := 0; i < 1000; i++ {
		locations[i] = Location{
			Latitude:  homeLat + float64(i)*0.001,
			Longitude: homeLon + float64(i)*0.001,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateMetrics(homeLat, homeLon, locations)
	}
}
