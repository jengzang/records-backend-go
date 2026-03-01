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
