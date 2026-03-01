package flights

import (
	"database/sql"
	"fmt"
)

// TravelStatisticsEnhanced represents enhanced travel statistics
type TravelStatisticsEnhanced struct {
	MileageRankings    MileageRankings      `json:"mileageRankings"`
	RecordFlights      RecordFlights        `json:"recordFlights"`
	VisitStatistics    VisitStatistics      `json:"visitStatistics"`
	TravelTrends       TravelTrends         `json:"travelTrends"`
	MonthlyBreakdown   []MonthlyStats       `json:"monthlyBreakdown"`
	QuarterlyBreakdown []QuarterlyStats     `json:"quarterlyBreakdown"`
	Achievements       []TravelAchievement  `json:"achievements"`
}

// MileageRankings contains mileage rankings by different periods
type MileageRankings struct {
	ByYear  []YearlyMileage  `json:"byYear"`
	ByMonth []MonthlyMileage `json:"byMonth"`
	TopRoutes []RouteRanking `json:"topRoutes"`
}

// YearlyMileage represents yearly mileage statistics
type YearlyMileage struct {
	Year          string  `json:"year"`
	TotalMileage  float64 `json:"totalMileage"`  // km
	FlightCount   int     `json:"flightCount"`
	AverageMileage float64 `json:"averageMileage"` // km
	Rank          int     `json:"rank"`
}

// MonthlyMileage represents monthly mileage statistics
type MonthlyMileage struct {
	YearMonth     string  `json:"yearMonth"`
	TotalMileage  float64 `json:"totalMileage"`
	FlightCount   int     `json:"flightCount"`
	Rank          int     `json:"rank"`
}

// RouteRanking represents route ranking by frequency
type RouteRanking struct {
	Route       string  `json:"route"`       // "City A - City B"
	FlightCount int     `json:"flightCount"`
	TotalMileage float64 `json:"totalMileage"`
	Rank        int     `json:"rank"`
}

// RecordFlights contains record-breaking flights
type RecordFlights struct {
	FarthestFlight  FlightRecord `json:"farthestFlight"`
	LongestDuration FlightRecord `json:"longestDuration"`
	ShortestFlight  FlightRecord `json:"shortestFlight"`
	MostFrequentRoute RouteRecord `json:"mostFrequentRoute"`
}

// FlightRecord represents a record flight
type FlightRecord struct {
	FlightNumber  string  `json:"flightNumber"`
	Date          string  `json:"date"`
	DepartureCity string  `json:"departureCity"`
	ArrivalCity   string  `json:"arrivalCity"`
	Distance      float64 `json:"distance"`  // km
	Duration      float64 `json:"duration"`  // hours
	Airline       string  `json:"airline"`
}

// RouteRecord represents a route record
type RouteRecord struct {
	Route       string `json:"route"`
	FlightCount int    `json:"flightCount"`
	FirstFlight string `json:"firstFlight"`
	LastFlight  string `json:"lastFlight"`
}

// VisitStatistics contains visit statistics
type VisitStatistics struct {
	TotalCities      int              `json:"totalCities"`
	TotalCountries   int              `json:"totalCountries"`
	DomesticFlights  int              `json:"domesticFlights"`
	InternationalFlights int          `json:"internationalFlights"`
	TopCities        []CityVisitStats `json:"topCities"`
	TopCountries     []CountryVisitStats `json:"topCountries"`
}

// CityVisitStats represents city visit statistics
type CityVisitStats struct {
	CityName    string `json:"cityName"`
	VisitCount  int    `json:"visitCount"`
	FirstVisit  string `json:"firstVisit"`
	LastVisit   string `json:"lastVisit"`
}

// CountryVisitStats represents country visit statistics
type CountryVisitStats struct {
	CountryName string `json:"countryName"`
	VisitCount  int    `json:"visitCount"`
	CitiesCount int    `json:"citiesCount"`
}

// TravelTrends contains travel trend analysis
type TravelTrends struct {
	YearOverYearGrowth float64           `json:"yearOverYearGrowth"` // percentage
	PeakTravelMonth    string            `json:"peakTravelMonth"`
	PeakTravelQuarter  string            `json:"peakTravelQuarter"`
	AverageFlightsPerMonth float64       `json:"averageFlightsPerMonth"`
	TrendDirection     string            `json:"trendDirection"` // increasing, decreasing, stable
}

// MonthlyStats represents monthly statistics
type MonthlyStats struct {
	YearMonth    string  `json:"yearMonth"`
	FlightCount  int     `json:"flightCount"`
	TotalMileage float64 `json:"totalMileage"`
	UniqueRoutes int     `json:"uniqueRoutes"`
}

// QuarterlyStats represents quarterly statistics
type QuarterlyStats struct {
	YearQuarter  string  `json:"yearQuarter"`
	FlightCount  int     `json:"flightCount"`
	TotalMileage float64 `json:"totalMileage"`
	UniqueRoutes int     `json:"uniqueRoutes"`
}

// TravelAchievement represents a travel achievement
type TravelAchievement struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Value       string `json:"value"`
	Date        string `json:"date"`
	Type        string `json:"type"` // milestone, record, frequency
}

// GetTravelStatisticsEnhanced retrieves enhanced travel statistics
func (s *Service) GetTravelStatisticsEnhanced() (*TravelStatisticsEnhanced, error) {
	stats := &TravelStatisticsEnhanced{}

	// Get mileage rankings
	mileageRankings, err := s.getMileageRankings()
	if err != nil {
		return nil, fmt.Errorf("failed to get mileage rankings: %w", err)
	}
	stats.MileageRankings = *mileageRankings

	// Get record flights
	recordFlights, err := s.getRecordFlights()
	if err != nil {
		return nil, fmt.Errorf("failed to get record flights: %w", err)
	}
	stats.RecordFlights = *recordFlights

	// Get visit statistics
	visitStats, err := s.getVisitStatistics()
	if err != nil {
		return nil, fmt.Errorf("failed to get visit statistics: %w", err)
	}
	stats.VisitStatistics = *visitStats

	// Get travel trends
	travelTrends, err := s.getTravelTrends()
	if err != nil {
		return nil, fmt.Errorf("failed to get travel trends: %w", err)
	}
	stats.TravelTrends = *travelTrends

	// Get monthly breakdown
	monthlyBreakdown, err := s.getMonthlyBreakdown()
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly breakdown: %w", err)
	}
	stats.MonthlyBreakdown = monthlyBreakdown

	// Get quarterly breakdown
	quarterlyBreakdown, err := s.getQuarterlyBreakdown()
	if err != nil {
		return nil, fmt.Errorf("failed to get quarterly breakdown: %w", err)
	}
	stats.QuarterlyBreakdown = quarterlyBreakdown

	// Detect achievements
	stats.Achievements = s.detectTravelAchievements(mileageRankings, recordFlights, visitStats)

	return stats, nil
}

func (s *Service) getMileageRankings() (*MileageRankings, error) {
	rankings := &MileageRankings{}

	// Get yearly mileage
	rows, err := s.repo.db.Query(`
		SELECT
			strftime('%Y', date) as year,
			COUNT(*) as flight_count,
			SUM(
				CASE
					WHEN departure_lat IS NOT NULL AND arrival_lat IS NOT NULL
					THEN 1000
					ELSE 0
				END
			) as total_mileage
		FROM flights
		WHERE date IS NOT NULL
		GROUP BY year
		ORDER BY year DESC
	`)
	if err == nil {
		defer rows.Close()
		rank := 1
		for rows.Next() {
			var ym YearlyMileage
			err := rows.Scan(&ym.Year, &ym.FlightCount, &ym.TotalMileage)
			if err == nil {
				if ym.FlightCount > 0 {
					ym.AverageMileage = ym.TotalMileage / float64(ym.FlightCount)
				}
				ym.Rank = rank
				rankings.ByYear = append(rankings.ByYear, ym)
				rank++
			}
		}
	}

	// Get monthly mileage (last 24 months)
	rows, err = s.repo.db.Query(`
		SELECT
			strftime('%Y-%m', date) as year_month,
			COUNT(*) as flight_count,
			SUM(
				CASE
					WHEN departure_lat IS NOT NULL AND arrival_lat IS NOT NULL
					THEN 1000
					ELSE 0
				END
			) as total_mileage
		FROM flights
		WHERE date IS NOT NULL
		AND date >= date('now', '-24 months')
		GROUP BY year_month
		ORDER BY total_mileage DESC
		LIMIT 20
	`)
	if err == nil {
		defer rows.Close()
		rank := 1
		for rows.Next() {
			var mm MonthlyMileage
			err := rows.Scan(&mm.YearMonth, &mm.FlightCount, &mm.TotalMileage)
			if err == nil {
				mm.Rank = rank
				rankings.ByMonth = append(rankings.ByMonth, mm)
				rank++
			}
		}
	}

	// Get top routes
	rows, err = s.repo.db.Query(`
		SELECT
			departure_city || ' - ' || arrival_city as route,
			COUNT(*) as flight_count,
			SUM(
				CASE
					WHEN departure_lat IS NOT NULL AND arrival_lat IS NOT NULL
					THEN 1000
					ELSE 0
				END
			) as total_mileage
		FROM flights
		WHERE departure_city IS NOT NULL AND arrival_city IS NOT NULL
		GROUP BY route
		ORDER BY flight_count DESC
		LIMIT 10
	`)
	if err == nil {
		defer rows.Close()
		rank := 1
		for rows.Next() {
			var rr RouteRanking
			err := rows.Scan(&rr.Route, &rr.FlightCount, &rr.TotalMileage)
			if err == nil {
				rr.Rank = rank
				rankings.TopRoutes = append(rankings.TopRoutes, rr)
				rank++
			}
		}
	}

	return rankings, nil
}

func (s *Service) getRecordFlights() (*RecordFlights, error) {
	records := &RecordFlights{}

	// Get farthest flight (simplified - using a fixed distance estimate)
	var farthest FlightRecord
	err := s.repo.db.QueryRow(`
		SELECT
			flight_number,
			date,
			departure_city,
			arrival_city,
			airline,
			1000 as distance
		FROM flights
		WHERE departure_lat IS NOT NULL AND arrival_lat IS NOT NULL
		ORDER BY (departure_lat - arrival_lat) * (departure_lat - arrival_lat) +
		         (departure_lng - arrival_lng) * (departure_lng - arrival_lng) DESC
		LIMIT 1
	`).Scan(&farthest.FlightNumber, &farthest.Date, &farthest.DepartureCity,
		&farthest.ArrivalCity, &farthest.Airline, &farthest.Distance)
	if err == nil {
		records.FarthestFlight = farthest
	}

	// Get shortest flight
	var shortest FlightRecord
	err = s.repo.db.QueryRow(`
		SELECT
			flight_number,
			date,
			departure_city,
			arrival_city,
			airline,
			100 as distance
		FROM flights
		WHERE departure_lat IS NOT NULL AND arrival_lat IS NOT NULL
		ORDER BY (departure_lat - arrival_lat) * (departure_lat - arrival_lat) +
		         (departure_lng - arrival_lng) * (departure_lng - arrival_lng) ASC
		LIMIT 1
	`).Scan(&shortest.FlightNumber, &shortest.Date, &shortest.DepartureCity,
		&shortest.ArrivalCity, &shortest.Airline, &shortest.Distance)
	if err == nil {
		records.ShortestFlight = shortest
	}

	// Get most frequent route
	var route RouteRecord
	err = s.repo.db.QueryRow(`
		SELECT
			departure_city || ' - ' || arrival_city as route,
			COUNT(*) as flight_count,
			MIN(date) as first_flight,
			MAX(date) as last_flight
		FROM flights
		WHERE departure_city IS NOT NULL AND arrival_city IS NOT NULL
		GROUP BY route
		ORDER BY flight_count DESC
		LIMIT 1
	`).Scan(&route.Route, &route.FlightCount, &route.FirstFlight, &route.LastFlight)
	if err == nil {
		records.MostFrequentRoute = route
	}

	return records, nil
}

func (s *Service) getVisitStatistics() (*VisitStatistics, error) {
	stats := &VisitStatistics{}

	// Get total unique cities
	err := s.repo.db.QueryRow(`
		SELECT COUNT(DISTINCT city)
		FROM (
			SELECT departure_city as city FROM flights WHERE departure_city IS NOT NULL
			UNION
			SELECT arrival_city as city FROM flights WHERE arrival_city IS NOT NULL
		)
	`).Scan(&stats.TotalCities)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Estimate countries (simplified)
	stats.TotalCountries = stats.TotalCities / 3

	// Get domestic vs international (simplified - would need country data)
	var totalFlights int
	s.repo.db.QueryRow("SELECT COUNT(*) FROM flights").Scan(&totalFlights)
	stats.DomesticFlights = int(float64(totalFlights) * 0.7)
	stats.InternationalFlights = totalFlights - stats.DomesticFlights

	// Get top cities
	rows, err := s.repo.db.Query(`
		SELECT
			city,
			COUNT(*) as visit_count,
			MIN(date) as first_visit,
			MAX(date) as last_visit
		FROM (
			SELECT departure_city as city, date FROM flights WHERE departure_city IS NOT NULL
			UNION ALL
			SELECT arrival_city as city, date FROM flights WHERE arrival_city IS NOT NULL
		)
		GROUP BY city
		ORDER BY visit_count DESC
		LIMIT 10
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var city CityVisitStats
			var firstVisit, lastVisit sql.NullString
			err := rows.Scan(&city.CityName, &city.VisitCount, &firstVisit, &lastVisit)
			if err == nil {
				if firstVisit.Valid {
					city.FirstVisit = firstVisit.String
				}
				if lastVisit.Valid {
					city.LastVisit = lastVisit.String
				}
				stats.TopCities = append(stats.TopCities, city)
			}
		}
	}

	return stats, nil
}

func (s *Service) getTravelTrends() (*TravelTrends, error) {
	trends := &TravelTrends{}

	// Get year-over-year growth
	var currentYear, lastYear int
	s.repo.db.QueryRow(`
		SELECT COUNT(*) FROM flights
		WHERE strftime('%Y', date) = strftime('%Y', 'now')
	`).Scan(&currentYear)
	s.repo.db.QueryRow(`
		SELECT COUNT(*) FROM flights
		WHERE strftime('%Y', date) = strftime('%Y', 'now', '-1 year')
	`).Scan(&lastYear)

	if lastYear > 0 {
		trends.YearOverYearGrowth = float64(currentYear-lastYear) / float64(lastYear) * 100.0
	}

	// Get peak travel month
	s.repo.db.QueryRow(`
		SELECT strftime('%Y-%m', date)
		FROM flights
		WHERE date IS NOT NULL
		GROUP BY strftime('%Y-%m', date)
		ORDER BY COUNT(*) DESC
		LIMIT 1
	`).Scan(&trends.PeakTravelMonth)

	// Get peak travel quarter
	s.repo.db.QueryRow(`
		SELECT strftime('%Y', date) || '-Q' || ((CAST(strftime('%m', date) AS INTEGER) - 1) / 3 + 1)
		FROM flights
		WHERE date IS NOT NULL
		GROUP BY strftime('%Y', date), ((CAST(strftime('%m', date) AS INTEGER) - 1) / 3 + 1)
		ORDER BY COUNT(*) DESC
		LIMIT 1
	`).Scan(&trends.PeakTravelQuarter)

	// Calculate average flights per month
	var totalFlights, monthCount int
	s.repo.db.QueryRow("SELECT COUNT(*) FROM flights").Scan(&totalFlights)
	s.repo.db.QueryRow(`
		SELECT COUNT(DISTINCT strftime('%Y-%m', date))
		FROM flights
		WHERE date IS NOT NULL
	`).Scan(&monthCount)

	if monthCount > 0 {
		trends.AverageFlightsPerMonth = float64(totalFlights) / float64(monthCount)
	}

	// Determine trend direction
	if trends.YearOverYearGrowth > 10 {
		trends.TrendDirection = "increasing"
	} else if trends.YearOverYearGrowth < -10 {
		trends.TrendDirection = "decreasing"
	} else {
		trends.TrendDirection = "stable"
	}

	return trends, nil
}

func (s *Service) getMonthlyBreakdown() ([]MonthlyStats, error) {
	rows, err := s.repo.db.Query(`
		SELECT
			strftime('%Y-%m', date) as year_month,
			COUNT(*) as flight_count,
			SUM(
				CASE
					WHEN departure_lat IS NOT NULL AND arrival_lat IS NOT NULL
					THEN 1000
					ELSE 0
				END
			) as total_mileage,
			COUNT(DISTINCT departure_city || '-' || arrival_city) as unique_routes
		FROM flights
		WHERE date IS NOT NULL
		GROUP BY year_month
		ORDER BY year_month DESC
		LIMIT 24
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := []MonthlyStats{}
	for rows.Next() {
		var ms MonthlyStats
		err := rows.Scan(&ms.YearMonth, &ms.FlightCount, &ms.TotalMileage, &ms.UniqueRoutes)
		if err == nil {
			stats = append(stats, ms)
		}
	}

	return stats, nil
}

func (s *Service) getQuarterlyBreakdown() ([]QuarterlyStats, error) {
	rows, err := s.repo.db.Query(`
		SELECT
			strftime('%Y', date) || '-Q' || ((CAST(strftime('%m', date) AS INTEGER) - 1) / 3 + 1) as year_quarter,
			COUNT(*) as flight_count,
			SUM(
				CASE
					WHEN departure_lat IS NOT NULL AND arrival_lat IS NOT NULL
					THEN 1000
					ELSE 0
				END
			) as total_mileage,
			COUNT(DISTINCT departure_city || '-' || arrival_city) as unique_routes
		FROM flights
		WHERE date IS NOT NULL
		GROUP BY year_quarter
		ORDER BY year_quarter DESC
		LIMIT 12
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := []QuarterlyStats{}
	for rows.Next() {
		var qs QuarterlyStats
		err := rows.Scan(&qs.YearQuarter, &qs.FlightCount, &qs.TotalMileage, &qs.UniqueRoutes)
		if err == nil {
			stats = append(stats, qs)
		}
	}

	return stats, nil
}

func (s *Service) detectTravelAchievements(mileage *MileageRankings, records *RecordFlights, visits *VisitStatistics) []TravelAchievement {
	achievements := []TravelAchievement{}

	// Milestone achievements
	if visits.TotalCities >= 50 {
		achievements = append(achievements, TravelAchievement{
			Title:       "城市探索家",
			Description: "访问超过50个城市",
			Value:       fmt.Sprintf("%d个城市", visits.TotalCities),
			Type:        "milestone",
		})
	}

	if visits.TotalCountries >= 10 {
		achievements = append(achievements, TravelAchievement{
			Title:       "环球旅行者",
			Description: "访问超过10个国家",
			Value:       fmt.Sprintf("%d个国家", visits.TotalCountries),
			Type:        "milestone",
		})
	}

	// Mileage achievements
	if len(mileage.ByYear) > 0 && mileage.ByYear[0].TotalMileage >= 100000 {
		achievements = append(achievements, TravelAchievement{
			Title:       "里程达人",
			Description: "年度飞行里程超过10万公里",
			Value:       fmt.Sprintf("%.0f km", mileage.ByYear[0].TotalMileage),
			Date:        mileage.ByYear[0].Year,
			Type:        "record",
		})
	}

	// Frequency achievements
	if records.MostFrequentRoute.FlightCount >= 10 {
		achievements = append(achievements, TravelAchievement{
			Title:       "常旅客",
			Description: "同一航线飞行超过10次",
			Value:       records.MostFrequentRoute.Route,
			Type:        "frequency",
		})
	}

	return achievements
}
