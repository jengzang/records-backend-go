package flights

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jengzang/records-backend-go/internal/logger"
	"github.com/sirupsen/logrus"
)

// Handler handles flight HTTP requests
type Handler struct {
	service *Service
}

// NewHandler creates a new flight handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetFlights returns a list of all flights
// GET /api/v1/flights
func (h *Handler) GetFlights(c *gin.Context) {
	logger.Info("Listing flights", logrus.Fields{
		"client_ip": c.ClientIP(),
	})

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	flights, total, err := h.service.ListFlights(page, pageSize)
	if err != nil {
		logger.Error("Failed to list flights", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve flights"})
		return
	}

	logger.Info("Flights listed successfully", logrus.Fields{
		"count": len(flights),
		"page":  page,
	})

	c.JSON(http.StatusOK, gin.H{
		"flights": flights,
		"pagination": gin.H{
			"page":     page,
			"pageSize": pageSize,
			"total":    total,
		},
	})
}

// GetFlight returns a single flight with its tracking points
// GET /api/v1/flights/:id
func (h *Handler) GetFlight(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid flight ID"})
		return
	}

	logger.Info("Getting flight details", logrus.Fields{
		"flight_id": id,
	})

	flight, points, err := h.service.GetFlight(id)
	if err != nil {
		logger.Error("Failed to get flight", err, logrus.Fields{
			"flight_id": id,
		})
		c.JSON(http.StatusNotFound, gin.H{"error": "Flight not found"})
		return
	}

	logger.Info("Flight retrieved successfully", logrus.Fields{
		"flight_id":   id,
		"point_count": len(points),
	})

	c.JSON(http.StatusOK, gin.H{
		"flight": flight,
		"points": points,
	})
}

// GetFlightSummary returns summary statistics for all flights
// GET /api/v1/flights/summary
func (h *Handler) GetFlightSummary(c *gin.Context) {
	logger.Info("Getting flight summary", logrus.Fields{
		"client_ip": c.ClientIP(),
	})

	summary, err := h.service.GetSummary()
	if err != nil {
		logger.Error("Failed to get flight summary", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve summary"})
		return
	}

	logger.Info("Flight summary retrieved", logrus.Fields{
		"total_flights": summary.TotalFlights,
		"total_distance": summary.TotalDistance,
	})

	c.JSON(http.StatusOK, summary)
}

// GetFlightRoute returns the route visualization data for a flight
// GET /api/v1/flights/:id/route
func (h *Handler) GetFlightRoute(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid flight ID"})
		return
	}

	logger.Info("Getting flight route", logrus.Fields{
		"flight_id": id,
	})

	flight, points, err := h.service.GetFlight(id)
	if err != nil {
		logger.Error("Failed to get flight route", err, logrus.Fields{
			"flight_id": id,
		})
		c.JSON(http.StatusNotFound, gin.H{"error": "Flight not found"})
		return
	}

	// Format points for map visualization
	routePoints := make([]gin.H, len(points))
	for i, p := range points {
		routePoints[i] = gin.H{
			"longitude": p.Longitude,
			"latitude":  p.Latitude,
			"altitude":  p.Altitude,
			"speed":     p.Speed,
			"timestamp": p.UpdateTime,
		}
	}

	logger.Info("Flight route retrieved", logrus.Fields{
		"flight_id":   id,
		"point_count": len(points),
	})

	c.JSON(http.StatusOK, gin.H{
		"flightNumber": flight.FlightNumber,
		"flightDate":   flight.FlightDate,
		"route":        routePoints,
		"statistics":   flight.Statistics,
	})
}

// SearchFlights searches flights with filters
// GET /api/v1/flights/search
func (h *Handler) SearchFlights(c *gin.Context) {
	logger.Info("Searching flights", logrus.Fields{
		"client_ip": c.ClientIP(),
		"query":     c.Request.URL.RawQuery,
	})

	// Parse filters
	filters := SearchFilters{
		FlightNumber: c.Query("flightNumber"),
		Airline:      c.Query("airline"),
		DateFrom:     c.Query("dateFrom"),
		DateTo:       c.Query("dateTo"),
		SortBy:       c.DefaultQuery("sortBy", "flight_date"),
		SortOrder:    c.DefaultQuery("sortOrder", "desc"),
	}

	// Parse numeric filters
	if minDist := c.Query("minDistance"); minDist != "" {
		if val, err := strconv.ParseFloat(minDist, 64); err == nil {
			filters.MinDistance = val
		}
	}

	if maxDist := c.Query("maxDistance"); maxDist != "" {
		if val, err := strconv.ParseFloat(maxDist, 64); err == nil {
			filters.MaxDistance = val
		}
	}

	// Parse pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filters.Limit = pageSize
	filters.Offset = (page - 1) * pageSize

	// Search flights
	flights, total, err := h.service.SearchFlights(filters)
	if err != nil {
		logger.Error("Failed to search flights", err, logrus.Fields{
			"filters": filters,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search flights"})
		return
	}

	logger.Info("Flights search completed", logrus.Fields{
		"result_count": len(flights),
		"total":        total,
		"page":         page,
	})

	c.JSON(http.StatusOK, gin.H{
		"flights": flights,
		"pagination": gin.H{
			"page":       page,
			"pageSize":   pageSize,
			"total":      total,
			"totalPages": (total + pageSize - 1) / pageSize,
		},
		"filters": filters,
	})
}

// GetAirlines returns all unique airlines
// GET /api/v1/flights/airlines
func (h *Handler) GetAirlines(c *gin.Context) {
	logger.Info("Getting airlines list", logrus.Fields{
		"client_ip": c.ClientIP(),
	})

	airlines, err := h.service.GetAirlines()
	if err != nil {
		logger.Error("Failed to get airlines", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve airlines"})
		return
	}

	logger.Info("Airlines retrieved", logrus.Fields{
		"count": len(airlines),
	})

	c.JSON(http.StatusOK, gin.H{
		"airlines": airlines,
		"count":    len(airlines),
	})
}

// GetAirlineStatistics returns statistics grouped by airline
// GET /api/v1/flights/statistics/airlines
func (h *Handler) GetAirlineStatistics(c *gin.Context) {
	logger.Info("Getting airline statistics", logrus.Fields{
		"client_ip": c.ClientIP(),
	})

	stats, err := h.service.GetAirlineStatistics()
	if err != nil {
		logger.Error("Failed to get airline statistics", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve statistics"})
		return
	}

	logger.Info("Airline statistics retrieved", logrus.Fields{
		"airline_count": len(stats),
	})

	c.JSON(http.StatusOK, gin.H{
		"airlines": stats,
		"count":    len(stats),
	})
}

// GetDateRange returns the date range of all flights
// GET /api/v1/flights/date-range
func (h *Handler) GetDateRange(c *gin.Context) {
	logger.Info("Getting flight date range", logrus.Fields{
		"client_ip": c.ClientIP(),
	})

	minDate, maxDate, err := h.service.GetDateRange()
	if err != nil {
		logger.Error("Failed to get date range", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve date range"})
		return
	}

	logger.Info("Date range retrieved", logrus.Fields{
		"min_date": minDate,
		"max_date": maxDate,
	})

	c.JSON(http.StatusOK, gin.H{
		"minDate": minDate,
		"maxDate": maxDate,
	})
}

// GetTravelFootprint returns travel footprint analysis
// GET /api/v1/flights/travel-footprint
func (h *Handler) GetTravelFootprint(c *gin.Context) {
	logger.Info("Getting travel footprint", logrus.Fields{
		"client_ip": c.ClientIP(),
	})

	footprint, err := h.service.GetTravelFootprint()
	if err != nil {
		logger.Error("Failed to get travel footprint", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve travel footprint"})
		return
	}

	logger.Info("Travel footprint retrieved", logrus.Fields{
		"cities_count":    len(footprint.VisitedCities),
		"countries_count": len(footprint.VisitedCountries),
		"routes_count":    len(footprint.FlightRoutes),
	})

	c.JSON(http.StatusOK, footprint)
}

// GetTravelStatisticsEnhanced returns enhanced travel statistics
// GET /api/v1/flights/statistics/enhanced
func (h *Handler) GetTravelStatisticsEnhanced(c *gin.Context) {
	logger.Info("Getting enhanced travel statistics", logrus.Fields{
		"client_ip": c.ClientIP(),
	})

	stats, err := h.service.GetTravelStatisticsEnhanced()
	if err != nil {
		logger.Error("Failed to get enhanced travel statistics", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve enhanced statistics"})
		return
	}

	logger.Info("Enhanced travel statistics retrieved", logrus.Fields{
		"yearly_records":    len(stats.MileageRankings.ByYear),
		"monthly_records":   len(stats.MonthlyBreakdown),
		"achievements":      len(stats.Achievements),
	})

	c.JSON(http.StatusOK, stats)
}

// GetCarbonEmission returns carbon emission analysis
// GET /api/v1/flights/carbon-emission
func (h *Handler) GetCarbonEmission(c *gin.Context) {
	logger.Info("Getting carbon emission analysis", logrus.Fields{
		"client_ip": c.ClientIP(),
	})

	analysis, err := h.service.GetCarbonEmission()
	if err != nil {
		logger.Error("Failed to get carbon emission analysis", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve carbon emission analysis"})
		return
	}

	logger.Info("Carbon emission analysis retrieved", logrus.Fields{
		"total_emission":   analysis.TotalEmission,
		"flight_emission":  analysis.FlightEmission,
		"railway_emission": analysis.RailwayEmission,
	})

	c.JSON(http.StatusOK, analysis)
}
