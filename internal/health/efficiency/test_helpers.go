package efficiency

// Test helper methods (exported for testing)

// FetchKeyboardDataTest is a test helper to expose fetchKeyboardData
func (s *Service) FetchKeyboardDataTest(date string, hour int) (*KeyboardMetrics, error) {
	return s.fetchKeyboardData(date, hour)
}

// FetchScreenTimeDataTest is a test helper to expose fetchScreenTimeData
func (s *Service) FetchScreenTimeDataTest(date string, hour int) (*ScreenTimeMetrics, error) {
	return s.fetchScreenTimeData(date, hour)
}

// FetchHealthDataTest is a test helper to expose fetchHealthData
func (s *Service) FetchHealthDataTest(date string, hour int) (*HealthMetrics, error) {
	return s.fetchHealthData(date, hour)
}
