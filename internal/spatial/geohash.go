package spatial

import (
	"math"
)

// Base32 encoding for geohash
const base32 = "0123456789bcdefghjkmnpqrstuvwxyz"

// EncodeGeohash encodes latitude and longitude into a geohash string
// precision: number of characters in the geohash (1-12)
func EncodeGeohash(lat, lon float64, precision int) string {
	if precision < 1 {
		precision = 1
	}
	if precision > 12 {
		precision = 12
	}

	latRange := []float64{-90.0, 90.0}
	lonRange := []float64{-180.0, 180.0}

	geohash := make([]byte, 0, precision)
	bits := 0
	bit := 0
	ch := 0

	for len(geohash) < precision {
		if bit%2 == 0 {
			// Longitude
			mid := (lonRange[0] + lonRange[1]) / 2
			if lon > mid {
				ch |= (1 << (4 - bits))
				lonRange[0] = mid
			} else {
				lonRange[1] = mid
			}
		} else {
			// Latitude
			mid := (latRange[0] + latRange[1]) / 2
			if lat > mid {
				ch |= (1 << (4 - bits))
				latRange[0] = mid
			} else {
				latRange[1] = mid
			}
		}

		bits++
		if bits == 5 {
			geohash = append(geohash, base32[ch])
			bits = 0
			ch = 0
		}
		bit++
	}

	return string(geohash)
}

// DecodeGeohash decodes a geohash string into latitude and longitude
// Returns center point of the geohash cell
func DecodeGeohash(geohash string) (lat, lon float64) {
	latRange := []float64{-90.0, 90.0}
	lonRange := []float64{-180.0, 180.0}

	isLon := true
	for i := 0; i < len(geohash); i++ {
		ch := geohash[i]
		idx := indexOfBase32(ch)
		if idx == -1 {
			continue
		}

		for mask := 16; mask > 0; mask >>= 1 {
			if isLon {
				mid := (lonRange[0] + lonRange[1]) / 2
				if idx&mask != 0 {
					lonRange[0] = mid
				} else {
					lonRange[1] = mid
				}
			} else {
				mid := (latRange[0] + latRange[1]) / 2
				if idx&mask != 0 {
					latRange[0] = mid
				} else {
					latRange[1] = mid
				}
			}
			isLon = !isLon
		}
	}

	lat = (latRange[0] + latRange[1]) / 2
	lon = (lonRange[0] + lonRange[1]) / 2
	return
}

// GeohashBounds returns the bounding box of a geohash cell
// Returns (minLat, minLon, maxLat, maxLon)
func GeohashBounds(geohash string) (float64, float64, float64, float64) {
	latRange := []float64{-90.0, 90.0}
	lonRange := []float64{-180.0, 180.0}

	isLon := true
	for i := 0; i < len(geohash); i++ {
		ch := geohash[i]
		idx := indexOfBase32(ch)
		if idx == -1 {
			continue
		}

		for mask := 16; mask > 0; mask >>= 1 {
			if isLon {
				mid := (lonRange[0] + lonRange[1]) / 2
				if idx&mask != 0 {
					lonRange[0] = mid
				} else {
					lonRange[1] = mid
				}
			} else {
				mid := (latRange[0] + latRange[1]) / 2
				if idx&mask != 0 {
					latRange[0] = mid
				} else {
					latRange[1] = mid
				}
			}
			isLon = !isLon
		}
	}

	return latRange[0], lonRange[0], latRange[1], lonRange[1]
}

// GeohashNeighbors returns the 8 neighboring geohash cells
func GeohashNeighbors(geohash string) []string {
	lat, lon := DecodeGeohash(geohash)
	precision := len(geohash)

	// Calculate cell size
	minLat, minLon, maxLat, maxLon := GeohashBounds(geohash)
	latDelta := maxLat - minLat
	lonDelta := maxLon - minLon

	neighbors := make([]string, 0, 8)
	for dLat := -1; dLat <= 1; dLat++ {
		for dLon := -1; dLon <= 1; dLon++ {
			if dLat == 0 && dLon == 0 {
				continue
			}
			newLat := lat + float64(dLat)*latDelta
			newLon := lon + float64(dLon)*lonDelta

			// Handle wrapping
			if newLat > 90 {
				newLat = 90
			}
			if newLat < -90 {
				newLat = -90
			}
			if newLon > 180 {
				newLon -= 360
			}
			if newLon < -180 {
				newLon += 360
			}

			neighbors = append(neighbors, EncodeGeohash(newLat, newLon, precision))
		}
	}

	return neighbors
}

// GeohashCellSize returns the approximate cell size in meters for a given precision
func GeohashCellSize(precision int) float64 {
	// Approximate cell sizes at equator
	sizes := map[int]float64{
		1:  5000000,  // ±2500 km
		2:  625000,   // ±312.5 km
		3:  123000,   // ±61.5 km
		4:  19500,    // ±9.75 km
		5:  3900,     // ±1.95 km
		6:  610,      // ±305 m
		7:  120,      // ±60 m
		8:  19,       // ±9.5 m
		9:  3.7,      // ±1.85 m
		10: 0.6,      // ±30 cm
		11: 0.12,     // ±6 cm
		12: 0.019,    // ±0.95 cm
	}

	if size, ok := sizes[precision]; ok {
		return size
	}
	return 0
}

// indexOfBase32 finds the index of a character in the base32 alphabet
func indexOfBase32(ch byte) int {
	for i := 0; i < len(base32); i++ {
		if base32[i] == ch {
			return i
		}
	}
	return -1
}

// GeohashPrecisionForDistance returns the appropriate geohash precision for a given distance
func GeohashPrecisionForDistance(distanceMeters float64) int {
	for precision := 1; precision <= 12; precision++ {
		if GeohashCellSize(precision) <= distanceMeters {
			return precision
		}
	}
	return 12
}
