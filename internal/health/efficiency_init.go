package health

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/jengzang/records-backend-go/internal/health/efficiency"
)

// NewEfficiencyHandler creates a new efficiency handler with cross-module database connections
func NewEfficiencyHandler(healthDB, keyboardDB, screentimeDB *sql.DB) *efficiency.Handler {
	service := efficiency.NewService(healthDB, keyboardDB, screentimeDB)
	return efficiency.NewHandler(service)
}

// RegisterEfficiencyRoutes registers efficiency routes (convenience function)
func RegisterEfficiencyRoutes(router *gin.RouterGroup, healthDB, keyboardDB, screentimeDB *sql.DB) {
	handler := NewEfficiencyHandler(healthDB, keyboardDB, screentimeDB)
	handler.RegisterRoutes(router)
}
