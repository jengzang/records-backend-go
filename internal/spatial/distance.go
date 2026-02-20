package spatial

import (
	"math"

	"github.com/golang/geo/s2"
)

// HaversineDistance calculates the great-circle distance between two points in meters
// using the Haversine formula
func HaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	p1 := s2.LatLngFromDegrees(lat1, lon1)
	p2 := s2.LatLngFromDegrees(lat2, lon2)
	return p1.Distance(p2).Radians() * EarthRadiusMeters
}

// Bearing calculates the initial bearing (forward azimuth) from point 1 to point 2
// Returns bearing in degrees (0-360), where 0 is North, 90 is East, etc.
func Bearing(lat1, lon1, lat2, lon2 float64) float64 {
	// Convert to radians
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lonDiff := (lon2 - lon1) * math.Pi / 180

	// Calculate bearing
	y := math.Sin(lonDiff) * math.Cos(lat2Rad)
	x := math.Cos(lat1Rad)*math.Sin(lat2Rad) - math.Sin(lat1Rad)*math.Cos(lat2Rad)*math.Cos(lonDiff)
	bearing := math.Atan2(y, x)

	// Convert to degrees and normalize to 0-360
	bearingDeg := bearing * 180 / math.Pi
	return math.Mod(bearingDeg+360, 360)
}

// BearingS2 calculates bearing using S2 geometry library
func BearingS2(p1, p2 s2.LatLng) float64 {
	lat1 := p1.Lat.Radians()
	lat2 := p2.Lat.Radians()
	lonDiff := p2.Lng.Radians() - p1.Lng.Radians()

	y := math.Sin(lonDiff) * math.Cos(lat2)
	x := math.Cos(lat1)*math.Sin(lat2) - math.Sin(lat1)*math.Cos(lat2)*math.Cos(lonDiff)
	bearing := math.Atan2(y, x)

	// Convert to degrees and normalize to 0-360
	bearingDeg := bearing * 180 / math.Pi
	return math.Mod(bearingDeg+360, 360)
}

// DestinationPoint calculates the destination point given a start point, bearing, and distance
// bearing: degrees (0-360), distance: meters
func DestinationPoint(lat, lon, bearing, distance float64) (float64, float64) {
	p := s2.LatLngFromDegrees(lat, lon)
	bearingRad := bearing * math.Pi / 180
	angularDistance := distance / EarthRadiusMeters

	latRad := p.Lat.Radians()
	lonRad := p.Lng.Radians()

	lat2 := math.Asin(math.Sin(latRad)*math.Cos(angularDistance) +
		math.Cos(latRad)*math.Sin(angularDistance)*math.Cos(bearingRad))

	lon2 := lonRad + math.Atan2(
		math.Sin(bearingRad)*math.Sin(angularDistance)*math.Cos(latRad),
		math.Cos(angularDistance)-math.Sin(latRad)*math.Sin(lat2))

	return lat2 * 180 / math.Pi, lon2 * 180 / math.Pi
}

// Midpoint calculates the midpoint between two points
func Midpoint(lat1, lon1, lat2, lon2 float64) (float64, float64) {
	p1 := s2.LatLngFromDegrees(lat1, lon1)
	p2 := s2.LatLngFromDegrees(lat2, lon2)

	// Use S2 interpolation
	mid := s2.Interpolate(0.5, s2.PointFromLatLng(p1), s2.PointFromLatLng(p2))
	midLatLng := s2.LatLngFromPoint(mid)

	return midLatLng.Lat.Degrees(), midLatLng.Lng.Degrees()
}

// Constants
const (
	EarthRadiusMeters = 6371000.0 // Earth's mean radius in meters
	EarthRadiusKm     = 6371.0    // Earth's mean radius in kilometers
)
