package railway

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jengzang/records-backend-go/internal/logger"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetAllLines handles GET /api/v1/railway/lines
func (h *Handler) GetAllLines(c *gin.Context) {
	logger.Info("Fetching all railway lines", logrus.Fields{
		"client_ip": c.ClientIP(),
	})

	lines, err := h.service.GetAllLines()
	if err != nil {
		logger.Error("Failed to fetch railway lines", err, logrus.Fields{})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch lines"})
		return
	}

	logger.Info("Railway lines fetched successfully", logrus.Fields{
		"count": len(lines),
	})

	c.JSON(http.StatusOK, lines)
}

// GetLineByID handles GET /api/v1/railway/lines/:id
func (h *Handler) GetLineByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid line ID"})
		return
	}

	logger.Info("Fetching railway line", logrus.Fields{
		"line_id": id,
	})

	line, err := h.service.GetLineByID(id)
	if err != nil {
		logger.Error("Failed to fetch railway line", err, logrus.Fields{
			"line_id": id,
		})
		c.JSON(http.StatusNotFound, gin.H{"error": "Line not found"})
		return
	}

	c.JSON(http.StatusOK, line)
}

// GetLineRoute handles GET /api/v1/railway/lines/:id/route
func (h *Handler) GetLineRoute(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid line ID"})
		return
	}

	logger.Info("Fetching railway line route", logrus.Fields{
		"line_id": id,
	})

	line, err := h.service.GetLineWithRoute(id)
	if err != nil {
		logger.Error("Failed to fetch railway line route", err, logrus.Fields{
			"line_id": id,
		})
		c.JSON(http.StatusNotFound, gin.H{"error": "Line not found"})
		return
	}

	logger.Info("Railway line route fetched successfully", logrus.Fields{
		"line_id":       id,
		"segment_count": len(line.Segments),
	})

	c.JSON(http.StatusOK, line)
}

// GetAllTrips handles GET /api/v1/railway/trips
func (h *Handler) GetAllTrips(c *gin.Context) {
	logger.Info("Fetching all railway trips", logrus.Fields{
		"client_ip": c.ClientIP(),
	})

	trips, err := h.service.GetAllTrips()
	if err != nil {
		logger.Error("Failed to fetch railway trips", err, logrus.Fields{})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch trips"})
		return
	}

	logger.Info("Railway trips fetched successfully", logrus.Fields{
		"count": len(trips),
	})

	c.JSON(http.StatusOK, trips)
}

// GetTripByID handles GET /api/v1/railway/trips/:id
func (h *Handler) GetTripByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid trip ID"})
		return
	}

	logger.Info("Fetching railway trip", logrus.Fields{
		"trip_id": id,
	})

	trip, err := h.service.GetTripByID(id)
	if err != nil {
		logger.Error("Failed to fetch railway trip", err, logrus.Fields{
			"trip_id": id,
		})
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip not found"})
		return
	}

	c.JSON(http.StatusOK, trip)
}

// GetStatistics handles GET /api/v1/railway/statistics
func (h *Handler) GetStatistics(c *gin.Context) {
	logger.Info("Fetching railway statistics", logrus.Fields{
		"client_ip": c.ClientIP(),
	})

	stats, err := h.service.GetStatistics()
	if err != nil {
		logger.Error("Failed to fetch railway statistics", err, logrus.Fields{})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch statistics"})
		return
	}

	logger.Info("Railway statistics fetched successfully", logrus.Fields{
		"total_lines": stats.TotalLines,
		"total_trips": stats.TotalTrips,
	})

	c.JSON(http.StatusOK, stats)
}

// CreateTrip handles POST /api/v1/railway/trips
func (h *Handler) CreateTrip(c *gin.Context) {
	var trip RailwayTrip
	if err := c.ShouldBindJSON(&trip); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logger.Info("Creating railway trip", logrus.Fields{
		"train_number": trip.TrainNumber,
	})

	if err := h.service.CreateTrip(&trip); err != nil {
		logger.Error("Failed to create railway trip", err, logrus.Fields{
			"train_number": trip.TrainNumber,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create trip"})
		return
	}

	logger.Info("Railway trip created successfully", logrus.Fields{
		"trip_id":      trip.ID,
		"train_number": trip.TrainNumber,
	})

	c.JSON(http.StatusCreated, trip)
}

// UploadKML handles POST /api/v1/railway/upload-kml
func (h *Handler) UploadKML(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	lineType := c.PostForm("line_type")
	if lineType == "" {
		lineType = "普速"
	}

	logger.Info("Uploading KML file", logrus.Fields{
		"filename":  file.Filename,
		"line_type": lineType,
	})

	// Open file
	f, err := file.Open()
	if err != nil {
		logger.Error("Failed to open uploaded file", err, logrus.Fields{
			"filename": file.Filename,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer f.Close()

	// Parse KML
	kmlData, err := h.service.ParseKML(f)
	if err != nil {
		logger.Error("Failed to parse KML", err, logrus.Fields{
			"filename": file.Filename,
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse KML: " + err.Error()})
		return
	}

	// Import line
	line, err := h.service.ImportKMLLine(kmlData, lineType)
	if err != nil {
		logger.Error("Failed to import KML line", err, logrus.Fields{
			"filename": file.Filename,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to import line"})
		return
	}

	logger.Info("KML file imported successfully", logrus.Fields{
		"filename": file.Filename,
		"line_id":  line.ID,
		"line_name": line.LineName,
	})

	c.JSON(http.StatusCreated, gin.H{
		"message": "KML imported successfully",
		"line":    line,
	})
}
