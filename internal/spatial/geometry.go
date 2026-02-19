package spatial

import (
	"math"
)

// Point represents a 2D point with latitude and longitude
type Point struct {
	Lat float64
	Lon float64
}

// Centroid calculates the geographic centroid of a set of points
func Centroid(points []Point) Point {
	if len(points) == 0 {
		return Point{}
	}

	var sumLat, sumLon float64
	for _, p := range points {
		sumLat += p.Lat
		sumLon += p.Lon
	}

	return Point{
		Lat: sumLat / float64(len(points)),
		Lon: sumLon / float64(len(points)),
	}
}

// WeightedCentroid calculates the weighted centroid of a set of points
func WeightedCentroid(points []Point, weights []float64) Point {
	if len(points) == 0 {
		return Point{}
	}

	var sumLat, sumLon, sumWeights float64
	for i, p := range points {
		w := 1.0
		if i < len(weights) {
			w = weights[i]
		}
		sumLat += p.Lat * w
		sumLon += p.Lon * w
		sumWeights += w
	}

	if sumWeights == 0 {
		return Centroid(points)
	}

	return Point{
		Lat: sumLat / sumWeights,
		Lon: sumLon / sumWeights,
	}
}

// RadiusOfGyration calculates the radius of gyration for a set of points
// This measures the spatial dispersion around the centroid
func RadiusOfGyration(points []Point) float64 {
	if len(points) == 0 {
		return 0
	}

	center := Centroid(points)

	var sumSquaredDist float64
	for _, p := range points {
		dist := HaversineDistance(center.Lat, center.Lon, p.Lat, p.Lon)
		sumSquaredDist += dist * dist
	}

	return math.Sqrt(sumSquaredDist / float64(len(points)))
}

// WeightedRadiusOfGyration calculates the weighted radius of gyration
func WeightedRadiusOfGyration(points []Point, weights []float64) float64 {
	if len(points) == 0 {
		return 0
	}

	center := WeightedCentroid(points, weights)

	var sumWeightedSquaredDist, sumWeights float64
	for i, p := range points {
		w := 1.0
		if i < len(weights) {
			w = weights[i]
		}
		dist := HaversineDistance(center.Lat, center.Lon, p.Lat, p.Lon)
		sumWeightedSquaredDist += w * dist * dist
		sumWeights += w
	}

	if sumWeights == 0 {
		return RadiusOfGyration(points)
	}

	return math.Sqrt(sumWeightedSquaredDist / sumWeights)
}

// BoundingBox calculates the bounding box of a set of points
// Returns (minLat, minLon, maxLat, maxLon)
func BoundingBox(points []Point) (float64, float64, float64, float64) {
	if len(points) == 0 {
		return 0, 0, 0, 0
	}

	minLat, maxLat := points[0].Lat, points[0].Lat
	minLon, maxLon := points[0].Lon, points[0].Lon

	for _, p := range points[1:] {
		if p.Lat < minLat {
			minLat = p.Lat
		}
		if p.Lat > maxLat {
			maxLat = p.Lat
		}
		if p.Lon < minLon {
			minLon = p.Lon
		}
		if p.Lon > maxLon {
			maxLon = p.Lon
		}
	}

	return minLat, minLon, maxLat, maxLon
}

// BoundingBoxArea calculates the area of a bounding box in square meters
func BoundingBoxArea(minLat, minLon, maxLat, maxLon float64) float64 {
	// Calculate width and height
	width := HaversineDistance(minLat, minLon, minLat, maxLon)
	height := HaversineDistance(minLat, minLon, maxLat, minLon)
	return width * height
}

// PathLength calculates the total length of a path (sequence of points) in meters
func PathLength(points []Point) float64 {
	if len(points) < 2 {
		return 0
	}

	var totalDist float64
	for i := 1; i < len(points); i++ {
		dist := HaversineDistance(points[i-1].Lat, points[i-1].Lon, points[i].Lat, points[i].Lon)
		totalDist += dist
	}

	return totalDist
}

// Tortuosity calculates the tortuosity of a path
// Tortuosity = actual path length / straight-line distance
// Value of 1 means straight line, >1 means curved/winding path
func Tortuosity(points []Point) float64 {
	if len(points) < 2 {
		return 1.0
	}

	pathLen := PathLength(points)
	straightDist := HaversineDistance(points[0].Lat, points[0].Lon, points[len(points)-1].Lat, points[len(points)-1].Lon)

	if straightDist == 0 {
		return 1.0
	}

	return pathLen / straightDist
}

// PolygonArea calculates the area of a polygon using the spherical excess formula
// Points should be in order (clockwise or counter-clockwise)
// Returns area in square meters
func PolygonArea(points []Point) float64 {
	if len(points) < 3 {
		return 0
	}

	// Use the shoelace formula for small polygons (approximation)
	// For large polygons, should use spherical geometry
	var sum float64
	for i := 0; i < len(points); i++ {
		j := (i + 1) % len(points)
		sum += (points[j].Lon - points[i].Lon) * (points[j].Lat + points[i].Lat)
	}

	// Convert to square meters (approximate)
	latRad := points[0].Lat * math.Pi / 180
	metersPerDegreeLat := 111320.0
	metersPerDegreeLon := 111320.0 * math.Cos(latRad)

	area := math.Abs(sum) * metersPerDegreeLat * metersPerDegreeLon / 2.0
	return area
}

// ConvexHullArea calculates the area of the convex hull (simplified version)
// For production use, consider using a proper convex hull algorithm
func ConvexHullArea(points []Point) float64 {
	// Simplified: use bounding box as approximation
	// For accurate convex hull, use scipy or implement Graham scan
	minLat, minLon, maxLat, maxLon := BoundingBox(points)
	return BoundingBoxArea(minLat, minLon, maxLat, maxLon)
}

// PointInPolygon checks if a point is inside a polygon using ray casting
func PointInPolygon(point Point, polygon []Point) bool {
	if len(polygon) < 3 {
		return false
	}

	inside := false
	j := len(polygon) - 1

	for i := 0; i < len(polygon); i++ {
		if ((polygon[i].Lat > point.Lat) != (polygon[j].Lat > point.Lat)) &&
			(point.Lon < (polygon[j].Lon-polygon[i].Lon)*(point.Lat-polygon[i].Lat)/(polygon[j].Lat-polygon[i].Lat)+polygon[i].Lon) {
			inside = !inside
		}
		j = i
	}

	return inside
}

// SimplifyPath simplifies a path using the Ramer-Douglas-Peucker algorithm
// epsilon: maximum distance (meters) from the simplified path
func SimplifyPath(points []Point, epsilon float64) []Point {
	if len(points) < 3 {
		return points
	}

	// Find the point with maximum distance from the line segment
	maxDist := 0.0
	maxIndex := 0

	for i := 1; i < len(points)-1; i++ {
		dist := perpendicularDistance(points[i], points[0], points[len(points)-1])
		if dist > maxDist {
			maxDist = dist
			maxIndex = i
		}
	}

	// If max distance is greater than epsilon, recursively simplify
	if maxDist > epsilon {
		// Recursive call
		left := SimplifyPath(points[:maxIndex+1], epsilon)
		right := SimplifyPath(points[maxIndex:], epsilon)

		// Combine results (remove duplicate middle point)
		result := make([]Point, len(left)+len(right)-1)
		copy(result, left)
		copy(result[len(left):], right[1:])
		return result
	}

	// If max distance is less than epsilon, return endpoints
	return []Point{points[0], points[len(points)-1]}
}

// perpendicularDistance calculates the perpendicular distance from a point to a line segment
func perpendicularDistance(point, lineStart, lineEnd Point) float64 {
	// Calculate the perpendicular distance using cross product
	x0, y0 := point.Lat, point.Lon
	x1, y1 := lineStart.Lat, lineStart.Lon
	x2, y2 := lineEnd.Lat, lineEnd.Lon

	num := math.Abs((y2-y1)*x0 - (x2-x1)*y0 + x2*y1 - y2*x1)
	den := math.Sqrt((y2-y1)*(y2-y1) + (x2-x1)*(x2-x1))

	if den == 0 {
		return HaversineDistance(point.Lat, point.Lon, lineStart.Lat, lineStart.Lon)
	}

	// Convert to meters (approximate)
	metersPerDegree := 111320.0
	return (num / den) * metersPerDegree
}
