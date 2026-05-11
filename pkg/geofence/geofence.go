package geofence

import "math"

const earthRadiusMeters = 6_371_000.0 // Earth's mean radius in meters

// Result holds the output of a geofence check.
type Result struct {
	Distance        float64 `json:"distance_meters"`    // Actual distance from center
	RadiusMeters    int     `json:"radius_meters"`      // Allowed radius
	IsWithinFence   bool    `json:"is_within_geofence"` // Whether position is inside the fence
}

// Check determines if a given coordinate (lat, lon) is within the geofence
// defined by a center point (centerLat, centerLon) and a radius in meters.
// Uses the Haversine formula for accurate great-circle distance on Earth's surface.
func Check(centerLat, centerLon, pointLat, pointLon float64, radiusMeters int) Result {
	distance := haversineDistance(centerLat, centerLon, pointLat, pointLon)

	return Result{
		Distance:      math.Round(distance*100) / 100, // Round to 2 decimal places
		RadiusMeters:  radiusMeters,
		IsWithinFence: distance <= float64(radiusMeters),
	}
}

// haversineDistance calculates the great-circle distance in meters between
// two points on Earth given their latitude and longitude in decimal degrees.
//
// Formula:
//   a = sin²(Δlat/2) + cos(lat1) × cos(lat2) × sin²(Δlon/2)
//   c = 2 × atan2(√a, √(1−a))
//   d = R × c
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := degreesToRadians(lat2 - lat1)
	dLon := degreesToRadians(lon2 - lon1)

	lat1Rad := degreesToRadians(lat1)
	lat2Rad := degreesToRadians(lat2)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusMeters * c
}

// degreesToRadians converts decimal degrees to radians.
func degreesToRadians(deg float64) float64 {
	return deg * math.Pi / 180
}
