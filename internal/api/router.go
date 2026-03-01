package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jengzang/records-backend-go/internal/config"
	"github.com/jengzang/records-backend-go/internal/database"
	"github.com/jengzang/records-backend-go/internal/flights"
	"github.com/jengzang/records-backend-go/internal/handler"
	"github.com/jengzang/records-backend-go/internal/health"
	"github.com/jengzang/records-backend-go/internal/keyboard"
	"github.com/jengzang/records-backend-go/internal/logger"
	"github.com/jengzang/records-backend-go/internal/middleware"
	"github.com/jengzang/records-backend-go/internal/railway"
	"github.com/jengzang/records-backend-go/internal/repository"
	"github.com/jengzang/records-backend-go/internal/screentime"
	"github.com/jengzang/records-backend-go/internal/service"
	"github.com/sirupsen/logrus"
	_ "modernc.org/sqlite"
)

// SetupRouter 设置路由
func SetupRouter(cfg *config.Config) *gin.Engine {
	// Create Gin engine without default middleware
	r := gin.New()

	// Add custom middleware
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimit(3, time.Second)) // 3 requests per second
	r.Use(gin.Recovery())

	// Initialize database
	db := database.GetDB()

	// Initialize repositories
	trackRepo := repository.NewTrackRepository(db)
	statsRepo := repository.NewStatsRepository(db)
	geocodingRepo := repository.NewGeocodingRepository(db)
	analysisTaskRepo := repository.NewAnalysisTaskRepository(db)
	segmentRepo := repository.NewSegmentRepository(db)
	stayRepo := repository.NewStayRepository(db)
	tripRepo := repository.NewTripRepository(db)
	gridRepo := repository.NewGridRepository(db)
	vizRepo := repository.NewVisualizationRepository(db)

	// Initialize services
	trackService := service.NewTrackService(trackRepo)
	statsService := service.NewStatsService(statsRepo)
	geocodingService := service.NewGeocodingService(geocodingRepo)
	analysisTaskService := service.NewAnalysisTaskService(analysisTaskRepo, db)
	segmentService := service.NewSegmentService(segmentRepo)
	stayService := service.NewStayService(stayRepo)
	tripService := service.NewTripService(tripRepo)
	gridService := service.NewGridService(gridRepo)
	vizService := service.NewVisualizationService(vizRepo)
	importService := service.NewImportService(db)

	// Inject dependencies for auto-trigger pipeline
	importService.SetGeocodingService(geocodingService)
	importService.SetAnalysisTaskService(analysisTaskService)

	// Initialize handlers
	trackHandler := handler.NewTrackHandler(trackService)
	statsHandler := handler.NewStatsHandler(statsService)
	geocodingHandler := handler.NewGeocodingHandler(geocodingService)
	analysisTaskHandler := handler.NewAnalysisTaskHandler(analysisTaskService)
	segmentHandler := handler.NewSegmentHandler(segmentService)
	stayHandler := handler.NewStayHandler(stayService)
	tripHandler := handler.NewTripHandler(tripService)
	gridHandler := handler.NewGridHandler(gridService)
	vizHandler := handler.NewVisualizationHandler(vizService)
	importHandler := handler.NewImportHandler(importService)

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Records Backend API is running",
		})
	})

	// API 路由组
	api := r.Group("/api/v1")
	{
		// 轨迹相关接口
		tracks := api.Group("/tracks")
		{
			// Track points endpoints
			tracks.GET("/points", trackHandler.GetTrackPoints)
			tracks.GET("/points/:id", trackHandler.GetTrackPointByID)
			tracks.GET("/ungeocoded", trackHandler.GetUngeocodedPoints)

			// Segments endpoints
			tracks.GET("/segments", segmentHandler.GetSegments)
			tracks.GET("/segments/:id", segmentHandler.GetSegmentByID)

			// Stays endpoints
			tracks.GET("/stays", stayHandler.GetStays)
			tracks.GET("/stays/:id", stayHandler.GetStayByID)

			// Trips endpoints
			tracks.GET("/trips", tripHandler.GetTrips)
			tracks.GET("/trips/:id", tripHandler.GetTripByID)

			// Statistics endpoints
			stats := tracks.Group("/statistics")
			{
				stats.GET("/footprint", statsHandler.GetFootprintStatistics)
				stats.GET("/time-distribution", statsHandler.GetTimeDistribution)
				stats.GET("/speed-distribution", statsHandler.GetSpeedDistribution)
			}
		}

		// 统计排行榜接口
		stats := api.Group("/stats")
		{
			stats.GET("/footprint/rankings", statsHandler.GetFootprintRankings)
			stats.GET("/stay/rankings", statsHandler.GetStayRankings)
			stats.GET("/extreme-events", statsHandler.GetExtremeEvents)
			stats.GET("/admin-crossings", statsHandler.GetAdminCrossings)
			stats.GET("/admin-view", statsHandler.GetAdminView)

			// Speed-space coupling endpoints
			stats.GET("/speed-space", statsHandler.GetSpeedSpaceStats)
			stats.GET("/speed-space/high-speed-zones", statsHandler.GetHighSpeedZones)
			stats.GET("/speed-space/slow-life-zones", statsHandler.GetSlowLifeZones)

			// Directional bias endpoints
			stats.GET("/directional-bias", statsHandler.GetDirectionalBiasStats)
			stats.GET("/directional-bias/top-areas", statsHandler.GetTopDirectionalAreas)
			stats.GET("/directional-bias/bidirectional", statsHandler.GetBidirectionalPatterns)

			// Revisit patterns endpoints
			stats.GET("/revisit-patterns", statsHandler.GetRevisitPatterns)
			stats.GET("/revisit-patterns/top-locations", statsHandler.GetTopRevisitLocations)
			stats.GET("/revisit-patterns/habitual", statsHandler.GetHabitualLocations)
			stats.GET("/revisit-patterns/periodic", statsHandler.GetPeriodicLocations)

			// Spatial utilization endpoints
			stats.GET("/spatial-utilization", statsHandler.GetSpatialUtilization)
			stats.GET("/spatial-utilization/destinations", statsHandler.GetDestinationAreas)
			stats.GET("/spatial-utilization/corridors", statsHandler.GetTransitCorridors)
			stats.GET("/spatial-utilization/deep-engagement", statsHandler.GetDeepEngagementAreas)

			// Density structure endpoints
			stats.GET("/density", statsHandler.GetDensityGrids)
			stats.GET("/density/core", statsHandler.GetCoreAreas)
			stats.GET("/density/rare", statsHandler.GetRareVisits)
			stats.GET("/density/clusters", statsHandler.GetDensityClusters)

			// Altitude dimension endpoints
			stats.GET("/altitude", statsHandler.GetAltitudeStats)
			stats.GET("/altitude/highest-spans", statsHandler.GetHighestAltitudeSpans)
			stats.GET("/altitude/highest-intensity", statsHandler.GetHighestVerticalIntensity)

			// Time-space compression endpoints
			stats.GET("/time-space-compression", statsHandler.GetTimeSpaceCompression)
			stats.GET("/time-space-compression/highest-intensity", statsHandler.GetHighestMovementIntensity)
			stats.GET("/time-space-compression/burst-periods", statsHandler.GetBurstPeriods)

			// Time-space slicing endpoints
			stats.GET("/time-space-slices", statsHandler.GetTimeSpaceSlices)
			stats.GET("/time-space-slices/weekly-pattern", statsHandler.GetWeeklyPattern)
			stats.GET("/time-space-slices/hourly-pattern", statsHandler.GetHourlyPattern)

			// Spatial complexity endpoint
			stats.GET("/spatial-complexity", statsHandler.GetSpatialComplexity)

			// Road overlap endpoint
			stats.GET("/road-overlap", statsHandler.GetRoadOverlapSummary)

			// Spatial persona endpoint
			stats.GET("/spatial-persona", statsHandler.GetSpatialPersona)
		}

		// 可视化接口
		viz := api.Group("/viz")
		{
			viz.GET("/grid-cells", gridHandler.GetGridCells)
			viz.GET("/heatmap", gridHandler.GetHeatmapData)
			viz.GET("/rendering", vizHandler.GetRenderingMetadata)
			viz.GET("/time-slices", vizHandler.GetTimeSliceData)
		}

		// 键盘鼠标统计接口
		keyboardHandler, err := keyboard.NewHandler(cfg.KeyboardDBPath)
		if err != nil {
			// Log error but don't fail server startup
			r.GET("/api/keyboard/*any", func(c *gin.Context) {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "keyboard module unavailable"})
			})
		} else {
			// Note: Don't close the handler here - it needs to stay open for the server lifetime
			keyboardHandler.RegisterRoutes(api)
		}

		// 飞机火车路线接口
		flightsDB, err := database.OpenDB(cfg.FlightsDBPath)
		if err != nil {
			logger.Warn("Failed to open flights database", logrus.Fields{
				"db_path": cfg.FlightsDBPath,
				"error":   err.Error(),
			})
			// Provide placeholder endpoint
			flights := api.Group("/flights")
			{
				flights.GET("", func(c *gin.Context) {
					c.JSON(http.StatusServiceUnavailable, gin.H{"error": "flights module unavailable"})
				})
			}
		} else {
			logger.Info("Flights database initialized", logrus.Fields{
				"db_path": cfg.FlightsDBPath,
			})

			// Initialize flights module
			flightsRepo := flights.NewRepository(flightsDB)
			flightsService := flights.NewService(flightsRepo)
			flightsHandler := flights.NewHandler(flightsService)

			// Register routes
			flightsGroup := api.Group("/flights")
			{
				flightsGroup.GET("", flightsHandler.GetFlights)
				flightsGroup.GET("/search", flightsHandler.SearchFlights)
				flightsGroup.GET("/summary", flightsHandler.GetFlightSummary)
				flightsGroup.GET("/airlines", flightsHandler.GetAirlines)
				flightsGroup.GET("/date-range", flightsHandler.GetDateRange)
				flightsGroup.GET("/travel-footprint", flightsHandler.GetTravelFootprint)
				flightsGroup.GET("/statistics/airlines", flightsHandler.GetAirlineStatistics)
				flightsGroup.GET("/:id", flightsHandler.GetFlight)
				flightsGroup.GET("/:id/route", flightsHandler.GetFlightRoute)
			}
		}


		// Railway module
		railwayDB, err := database.OpenDB(cfg.RailwayDBPath)
		if err != nil {
			logger.Warn("Failed to open railway database", logrus.Fields{
				"error": err.Error(),
				"path":  cfg.RailwayDBPath,
			})
			railwayFallback := api.Group("/railway")
			{
				railwayFallback.GET("", func(c *gin.Context) {
					c.JSON(http.StatusServiceUnavailable, gin.H{"error": "railway module unavailable"})
				})
			}
		} else {
			logger.Info("Railway database opened successfully", logrus.Fields{
				"path": cfg.RailwayDBPath,
			})
			railwayRepo := railway.NewRepository(railwayDB)
			railwayService := railway.NewService(railwayRepo)
			railwayHandler := railway.NewHandler(railwayService)
			railwayGroup := api.Group("/railway")
			{
				railwayGroup.GET("/lines", railwayHandler.GetAllLines)
				railwayGroup.GET("/lines/:id", railwayHandler.GetLineByID)
				railwayGroup.GET("/lines/:id/route", railwayHandler.GetLineRoute)
				railwayGroup.GET("/trips", railwayHandler.GetAllTrips)
				railwayGroup.GET("/trips/:id", railwayHandler.GetTripByID)
				railwayGroup.POST("/trips", railwayHandler.CreateTrip)
				railwayGroup.GET("/statistics", railwayHandler.GetStatistics)
				railwayGroup.POST("/upload-kml", railwayHandler.UploadKML)
			}
		}
		// 屏幕使用时间接口
		screentimeHandler, err := screentime.NewHandler(cfg.ScreentimeDBPath)
		if err != nil {
			// Log error but don't fail server startup
			r.GET("/api/screentime/*any", func(c *gin.Context) {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "screentime module unavailable"})
			})
		} else {
			// Initialize device manager for cross-device analysis
			deviceManager, err := screentime.NewDeviceManager(cfg.ScreentimeDevicesDB, cfg.ScreentimeDataDir)
			if err != nil {
				logger.Warn("Failed to initialize device manager", logrus.Fields{
					"error": err.Error(),
				})
			}

			// Create multi-device handler
			var multiDeviceHandler *screentime.MultiDeviceHandler
			if deviceManager != nil {
				multiDeviceHandler = screentime.NewMultiDeviceHandler(deviceManager)
			}

			st := api.Group("/screentime")
			{
				// Basic single-device endpoints
				st.GET("/summary", screentimeHandler.GetSummary)
				st.GET("/daily", screentimeHandler.GetDailyStats)
				st.GET("/rankings", screentimeHandler.GetRankings)
				st.GET("/categories", screentimeHandler.GetCategories)
				st.GET("/hourly", screentimeHandler.GetHourlyStats)
				st.GET("/trends", screentimeHandler.GetTrends)
				st.GET("/app/:packageId", screentimeHandler.GetAppDetail)

				// Advanced analysis endpoints
				analysis := st.Group("/analysis")
				{
					analysis.GET("/time-waste", screentimeHandler.GetTimeWasteDetection)
					analysis.GET("/app-dependency", screentimeHandler.GetAppDependencyAnalysis)
					analysis.GET("/weekday-weekend", screentimeHandler.GetWeekdayWeekendComparison)
					analysis.GET("/productivity-entertainment-trend", screentimeHandler.GetProductivityEntertainmentTrend)
				}

				// Cross-device analysis endpoints (if device manager is available)
				if multiDeviceHandler != nil {
					crossDevice := st.Group("/cross-device")
					{
						crossDevice.GET("/comparison", multiDeviceHandler.GetCrossDeviceComparison)
						crossDevice.GET("/work-life-balance", multiDeviceHandler.GetWorkLifeBalance)
						crossDevice.GET("/total-screentime", multiDeviceHandler.GetTotalScreentime)
						crossDevice.GET("/switching-patterns", multiDeviceHandler.GetDeviceSwitchingPatterns)
						crossDevice.GET("/app-ecosystem", multiDeviceHandler.GetAppEcosystem)
						crossDevice.GET("/time-allocation", multiDeviceHandler.GetTimeAllocation)
						crossDevice.GET("/user-profile", multiDeviceHandler.GetUserProfile)
						crossDevice.GET("/productivity-deep", multiDeviceHandler.GetProductivityDeep)
						crossDevice.GET("/focus-analysis", multiDeviceHandler.GetFocusAnalysis)
						crossDevice.GET("/recommendations", multiDeviceHandler.GetCrossDeviceRecommendations)
					}

					// Multi-device endpoints (device parameter support)
					st.GET("/multi/summary", multiDeviceHandler.GetSummary)
					st.GET("/multi/daily", multiDeviceHandler.GetDailyStats)
					st.GET("/multi/rankings", multiDeviceHandler.GetRankings)
					st.GET("/devices", multiDeviceHandler.ListDevices)
				}
			}
		}

		// Apple健康数据接口
		healthDB, err := database.OpenDB(cfg.HealthDBPath)
		if err != nil {
			logger.Warn("Failed to open health database", logrus.Fields{
				"db_path": cfg.HealthDBPath,
				"error":   err.Error(),
			})
			// Provide placeholder endpoint
			healthGroup := api.Group("/health")
			{
				healthGroup.GET("", func(c *gin.Context) {
					c.JSON(http.StatusServiceUnavailable, gin.H{"error": "health module unavailable"})
				})
			}
		} else {
			logger.Info("Health database initialized", logrus.Fields{
				"db_path": cfg.HealthDBPath,
			})

			// Initialize health module
			healthHandler := health.NewHandler(healthDB)

			// Register routes
			healthGroup := api.Group("/health")
			{
				// Basic queries
				healthGroup.GET("/summary", healthHandler.GetSummary)
				healthGroup.GET("/records", healthHandler.GetRecords)
				healthGroup.GET("/workouts", healthHandler.GetWorkouts)
				healthGroup.GET("/workouts/:id", healthHandler.GetWorkout)
				healthGroup.GET("/workouts/:id/route", healthHandler.GetWorkoutRoute)

				// Statistics
				healthGroup.GET("/statistics/daily", healthHandler.GetDailyStatistics)
				healthGroup.GET("/statistics/weekly", healthHandler.GetWeeklyStatistics)
				healthGroup.GET("/statistics/monthly", healthHandler.GetMonthlyStatistics)
				healthGroup.GET("/statistics/trends", healthHandler.GetTrends)
				healthGroup.GET("/statistics/sleep", healthHandler.GetSleepStatistics)

				// Advanced analysis
				healthGroup.GET("/analysis/activity-patterns", healthHandler.GetActivityPatterns)
				healthGroup.GET("/analysis/health-score", healthHandler.GetHealthScoreForDate)
				healthGroup.GET("/analysis/weight-bmi", healthHandler.GetWeightBMIAnalysis)
				healthGroup.GET("/analysis/exercise", healthHandler.GetExerciseAnalysis)
				healthGroup.GET("/analysis/sleep", healthHandler.GetSleepAnalysis)

				// Heart rate analysis
				healthGroup.GET("/analysis/heartrate/zones", healthHandler.GetHeartRateZones)
				healthGroup.GET("/analysis/heartrate/anomalies", healthHandler.GetHeartRateAnomalies)
				healthGroup.GET("/analysis/heartrate/resting", healthHandler.GetRestingHeartRate)

				// Activity patterns
				healthGroup.GET("/analysis/patterns/daily", healthHandler.GetDailyActivityPattern)
				healthGroup.GET("/analysis/patterns/weekly", healthHandler.GetWeeklyActivityPattern)
			}
		}

		// 管理员接口
		admin := api.Group("/admin")
		{
			// Data import management
			tracks := admin.Group("/tracks")
			{
				tracks.POST("/import", importHandler.ImportData)
				tracks.GET("/import", importHandler.ListImportTasks)
				tracks.GET("/import/:id", importHandler.GetImportStatus)
			}

			// Unified pipeline (import + geocoding + analysis)
			pipeline := admin.Group("/pipeline")
			{
				pipeline.POST("/trigger", importHandler.TriggerPipeline)
			}

			// Geocoding tasks management
			geocoding := admin.Group("/geocoding")
			{
				geocoding.POST("/tasks", geocodingHandler.CreateTask)
				geocoding.GET("/tasks", geocodingHandler.ListTasks)
				geocoding.GET("/tasks/:id", geocodingHandler.GetTask)
				geocoding.DELETE("/tasks/:id", geocodingHandler.CancelTask)
			}

			// Analysis tasks management
			analysis := admin.Group("/analysis")
			{
				analysis.POST("/tasks", analysisTaskHandler.CreateTask)
				analysis.GET("/tasks", analysisTaskHandler.ListTasks)
				analysis.GET("/tasks/:id", analysisTaskHandler.GetTask)
				analysis.DELETE("/tasks/:id", analysisTaskHandler.CancelTask)
				analysis.POST("/trigger-chain", analysisTaskHandler.TriggerAnalysisChain)
			}
		}
	}

	return r
}
