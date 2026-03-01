package screentime

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/jengzang/records-backend-go/internal/logger"
	"github.com/sirupsen/logrus"
)

// DeviceManager manages multiple screentime data sources
type DeviceManager struct {
	devicesDB   *sql.DB
	connections map[string]*DeviceConnection
	mu          sync.RWMutex
	dataDir     string
}

// DeviceConnection represents a connection to a device's database
type DeviceConnection struct {
	ID         string
	Name       string
	Type       string // phone, computer
	DBPath     string
	DataFormat string // phone_txt, manictime_excel
	DB         *sql.DB
	IsActive   bool
}

// Device represents device metadata
type Device struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	DBPath         string `json:"dbPath"`
	DataFormat     string `json:"dataFormat"`
	IsActive       bool   `json:"isActive"`
	CreatedAt      string `json:"createdAt"`
	LastSync       string `json:"lastSync,omitempty"`
	TotalRecords   int    `json:"totalRecords"`
	DateRangeStart string `json:"dateRangeStart,omitempty"`
	DateRangeEnd   string `json:"dateRangeEnd,omitempty"`
	Metadata       string `json:"metadata,omitempty"`
}

// NewDeviceManager creates a new device manager
func NewDeviceManager(devicesDBPath, dataDir string) (*DeviceManager, error) {
	logger.Info("Initializing device manager", logrus.Fields{
		"devices_db": devicesDBPath,
		"data_dir":   dataDir,
	})

	db, err := sql.Open("sqlite", devicesDBPath)
	if err != nil {
		logger.Error("Failed to open devices database", err, logrus.Fields{
			"db_path": devicesDBPath,
		})
		return nil, fmt.Errorf("failed to open devices database: %w", err)
	}

	dm := &DeviceManager{
		devicesDB:   db,
		connections: make(map[string]*DeviceConnection),
		dataDir:     dataDir,
	}

	// Load all active devices
	if err := dm.loadDevices(); err != nil {
		logger.Error("Failed to load devices", err, nil)
		return nil, fmt.Errorf("failed to load devices: %w", err)
	}

	logger.Info("Device manager initialized successfully", logrus.Fields{
		"device_count": len(dm.connections),
	})

	return dm, nil
}

// loadDevices loads all active devices from the devices database
func (dm *DeviceManager) loadDevices() error {
	query := `SELECT id, name, type, db_path, data_format, is_active FROM devices WHERE is_active = 1`
	rows, err := dm.devicesDB.Query(query)
	if err != nil {
		logger.Error("Failed to query devices", err, nil)
		return err
	}
	defer rows.Close()

	loadedCount := 0
	for rows.Next() {
		var device Device
		err := rows.Scan(&device.ID, &device.Name, &device.Type, &device.DBPath, &device.DataFormat, &device.IsActive)
		if err != nil {
			logger.Warn("Failed to scan device row", logrus.Fields{
				"error": err.Error(),
			})
			continue
		}

		// Open connection to device database
		dbPath := filepath.Join(dm.dataDir, device.DBPath)
		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			logger.Error("Failed to open device database", err, logrus.Fields{
				"device_id": device.ID,
				"db_path":   dbPath,
			})
			continue
		}

		dm.mu.Lock()
		dm.connections[device.ID] = &DeviceConnection{
			ID:         device.ID,
			Name:       device.Name,
			Type:       device.Type,
			DBPath:     device.DBPath,
			DataFormat: device.DataFormat,
			DB:         db,
			IsActive:   device.IsActive,
		}
		dm.mu.Unlock()

		logger.Info("Device loaded successfully", logrus.Fields{
			"device_id":   device.ID,
			"device_name": device.Name,
			"device_type": device.Type,
		})
		loadedCount++
	}

	logger.Info("Devices loaded", logrus.Fields{
		"count": loadedCount,
	})

	return rows.Err()
}

// GetDevice returns a device connection by ID
func (dm *DeviceManager) GetDevice(id string) (*DeviceConnection, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	conn, exists := dm.connections[id]
	if !exists {
		logger.Error("Device not found", nil, logrus.Fields{
			"device_id":         id,
			"available_devices": dm.getDeviceIDs(),
		})
		return nil, fmt.Errorf("device not found: %s", id)
	}

	logger.Debug("Device retrieved", logrus.Fields{
		"device_id": id,
	})

	return conn, nil
}

// getDeviceIDs returns a list of all device IDs (must be called with lock held)
func (dm *DeviceManager) getDeviceIDs() []string {
	ids := make([]string, 0, len(dm.connections))
	for id := range dm.connections {
		ids = append(ids, id)
	}
	return ids
}

// ListDevices returns all registered devices
func (dm *DeviceManager) ListDevices() ([]*Device, error) {
	query := `
	SELECT id, name, type, db_path, data_format, is_active, 
	       created_at, last_sync, total_records, date_range_start, date_range_end, metadata
	FROM devices
	ORDER BY created_at DESC
	`

	rows, err := dm.devicesDB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []*Device
	for rows.Next() {
		var device Device
		var lastSync, dateStart, dateEnd, metadata sql.NullString

		err := rows.Scan(
			&device.ID, &device.Name, &device.Type, &device.DBPath, &device.DataFormat,
			&device.IsActive, &device.CreatedAt, &lastSync, &device.TotalRecords,
			&dateStart, &dateEnd, &metadata,
		)
		if err != nil {
			continue
		}

		if lastSync.Valid {
			device.LastSync = lastSync.String
		}
		if dateStart.Valid {
			device.DateRangeStart = dateStart.String
		}
		if dateEnd.Valid {
			device.DateRangeEnd = dateEnd.String
		}
		if metadata.Valid {
			device.Metadata = metadata.String
		}

		devices = append(devices, &device)
	}

	return devices, rows.Err()
}

// RegisterDevice registers a new device
func (dm *DeviceManager) RegisterDevice(device *Device) error {
	logger.Info("Registering device", logrus.Fields{
		"device_id":   device.ID,
		"device_name": device.Name,
		"device_type": device.Type,
		"db_path":     device.DBPath,
	})

	query := `
	INSERT INTO devices (id, name, type, db_path, data_format, is_active)
	VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := dm.devicesDB.Exec(query, device.ID, device.Name, device.Type, device.DBPath, device.DataFormat, device.IsActive)
	if err != nil {
		logger.Error("Failed to register device", err, logrus.Fields{
			"device_id": device.ID,
		})
		return fmt.Errorf("failed to register device: %w", err)
	}

	// Reload devices to include the new one
	if err := dm.loadDevices(); err != nil {
		logger.Error("Failed to reload devices after registration", err, nil)
		return err
	}

	logger.Info("Device registered successfully", logrus.Fields{
		"device_id": device.ID,
	})

	return nil
}

// GetAllActiveConnections returns all active device connections
func (dm *DeviceManager) GetAllActiveConnections() []*DeviceConnection {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	var connections []*DeviceConnection
	for _, conn := range dm.connections {
		if conn.IsActive {
			connections = append(connections, conn)
		}
	}

	return connections
}

// Close closes all database connections
func (dm *DeviceManager) Close() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	for _, conn := range dm.connections {
		if conn.DB != nil {
			conn.DB.Close()
		}
	}

	if dm.devicesDB != nil {
		return dm.devicesDB.Close()
	}

	return nil
}
