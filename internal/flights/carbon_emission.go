package flights

import (
	"database/sql"
	"math"
)

// CarbonEmissionAnalysis 碳排放分析结果
type CarbonEmissionAnalysis struct {
	TotalEmission      float64                    `json:"total_emission"`       // 总碳排放(kg CO2)
	FlightEmission     float64                    `json:"flight_emission"`      // 航班碳排放
	RailwayEmission    float64                    `json:"railway_emission"`     // 铁路碳排放
	EmissionByYear     []YearEmission             `json:"emission_by_year"`     // 年度碳排放
	EmissionByAirline  []AirlineEmission          `json:"emission_by_airline"`  // 航司碳排放
	EmissionComparison EmissionComparison         `json:"emission_comparison"`  // 排放对比
	CarbonFootprint    CarbonFootprint            `json:"carbon_footprint"`     // 碳足迹报告
	Recommendations    []string                   `json:"recommendations"`      // 碳中和建议
}

// YearEmission 年度碳排放
type YearEmission struct {
	Year            int     `json:"year"`
	FlightEmission  float64 `json:"flight_emission"`
	RailwayEmission float64 `json:"railway_emission"`
	TotalEmission   float64 `json:"total_emission"`
	FlightCount     int     `json:"flight_count"`
	RailwayCount    int     `json:"railway_count"`
}

// AirlineEmission 航司碳排放
type AirlineEmission struct {
	Airline  string  `json:"airline"`
	Emission float64 `json:"emission"`
	Count    int     `json:"count"`
}

// EmissionComparison 排放对比
type EmissionComparison struct {
	FlightVsRailway    float64 `json:"flight_vs_railway"`     // 航空vs铁路比例
	AvgFlightEmission  float64 `json:"avg_flight_emission"`   // 平均航班排放
	AvgRailwayEmission float64 `json:"avg_railway_emission"`  // 平均铁路排放
	EmissionSaved      float64 `json:"emission_saved"`        // 如果全用铁路节省的排放
}

// CarbonFootprint 碳足迹报告
type CarbonFootprint struct {
	AnnualAverage    float64 `json:"annual_average"`     // 年均碳排放
	TreesNeeded      int     `json:"trees_needed"`       // 需要种植的树木数量(抵消)
	EquivalentCars   float64 `json:"equivalent_cars"`    // 相当于汽车行驶公里数
	GlobalPercentile float64 `json:"global_percentile"`  // 全球百分位
}

// 碳排放系数(kg CO2/km)
const (
	FlightEmissionFactor  = 0.115 // 航空: 115g CO2/km
	RailwayEmissionFactor = 0.041 // 铁路: 41g CO2/km (高铁)
	TreeAbsorption        = 21.77  // 一棵树年均吸收CO2(kg)
	CarEmissionFactor     = 0.192  // 汽车: 192g CO2/km
)

// GetCarbonEmissionAnalysis 获取碳排放分析
func GetCarbonEmissionAnalysis() (*CarbonEmissionAnalysis, error) {
	db, err := sql.Open("sqlite3", "data/flights.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	analysis := &CarbonEmissionAnalysis{}

	// 计算航班碳排放
	flightEmission, flightsByYear, flightsByAirline, err := calculateFlightEmission(db)
	if err != nil {
		return nil, err
	}
	analysis.FlightEmission = flightEmission

	// 计算铁路碳排放
	railwayEmission, railwaysByYear, err := calculateRailwayEmission(db)
	if err != nil {
		return nil, err
	}
	analysis.RailwayEmission = railwayEmission

	// 总排放
	analysis.TotalEmission = flightEmission + railwayEmission

	// 年度排放
	analysis.EmissionByYear = mergeYearEmissions(flightsByYear, railwaysByYear)

	// 航司排放
	analysis.EmissionByAirline = flightsByAirline

	// 排放对比
	analysis.EmissionComparison = calculateEmissionComparison(
		flightEmission, railwayEmission,
		len(flightsByYear), len(railwaysByYear),
	)

	// 碳足迹报告
	analysis.CarbonFootprint = calculateCarbonFootprint(analysis.TotalEmission, len(analysis.EmissionByYear))

	// 碳中和建议
	analysis.Recommendations = generateRecommendations(analysis)

	return analysis, nil
}

// calculateFlightEmission 计算航班碳排放
func calculateFlightEmission(db *sql.DB) (float64, map[int]YearEmission, []AirlineEmission, error) {
	query := `
		SELECT
			strftime('%Y', date) as year,
			airline,
			distance
		FROM flights
		WHERE distance > 0
		ORDER BY date
	`

	rows, err := db.Query(query)
	if err != nil {
		return 0, nil, nil, err
	}
	defer rows.Close()

	totalEmission := 0.0
	yearMap := make(map[int]YearEmission)
	airlineMap := make(map[string]*AirlineEmission)

	for rows.Next() {
		var year int
		var airline string
		var distance float64

		if err := rows.Scan(&year, &airline, &distance); err != nil {
			continue
		}

		emission := distance * FlightEmissionFactor

		totalEmission += emission

		// 年度统计
		ye := yearMap[year]
		ye.Year = year
		ye.FlightEmission += emission
		ye.FlightCount++
		yearMap[year] = ye

		// 航司统计
		if airlineMap[airline] == nil {
			airlineMap[airline] = &AirlineEmission{Airline: airline}
		}
		airlineMap[airline].Emission += emission
		airlineMap[airline].Count++
	}

	// 转换为切片
	airlines := make([]AirlineEmission, 0, len(airlineMap))
	for _, ae := range airlineMap {
		airlines = append(airlines, *ae)
	}

	return totalEmission, yearMap, airlines, nil
}

// calculateRailwayEmission 计算铁路碳排放
func calculateRailwayEmission(db *sql.DB) (float64, map[int]YearEmission, error) {
	query := `
		SELECT
			strftime('%Y', date) as year,
			distance
		FROM railway_trips
		WHERE distance > 0
		ORDER BY date
	`

	rows, err := db.Query(query)
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()

	totalEmission := 0.0
	yearMap := make(map[int]YearEmission)

	for rows.Next() {
		var year int
		var distance float64

		if err := rows.Scan(&year, &distance); err != nil {
			continue
		}

		emission := distance * RailwayEmissionFactor

		totalEmission += emission

		// 年度统计
		ye := yearMap[year]
		ye.Year = year
		ye.RailwayEmission += emission
		ye.RailwayCount++
		yearMap[year] = ye
	}

	return totalEmission, yearMap, nil
}

// mergeYearEmissions 合并年度排放数据
func mergeYearEmissions(flights, railways map[int]YearEmission) []YearEmission {
	yearMap := make(map[int]YearEmission)

	for year, fe := range flights {
		ye := yearMap[year]
		ye.Year = year
		ye.FlightEmission = fe.FlightEmission
		ye.FlightCount = fe.FlightCount
		yearMap[year] = ye
	}

	for year, re := range railways {
		ye := yearMap[year]
		ye.Year = year
		ye.RailwayEmission = re.RailwayEmission
		ye.RailwayCount = re.RailwayCount
		yearMap[year] = ye
	}

	// 计算总排放
	result := make([]YearEmission, 0, len(yearMap))
	for _, ye := range yearMap {
		ye.TotalEmission = ye.FlightEmission + ye.RailwayEmission
		result = append(result, ye)
	}

	return result
}

// calculateEmissionComparison 计算排放对比
func calculateEmissionComparison(flightEmission, railwayEmission float64, flightCount, railwayCount int) EmissionComparison {
	comp := EmissionComparison{}

	if railwayEmission > 0 {
		comp.FlightVsRailway = flightEmission / railwayEmission
	}

	if flightCount > 0 {
		comp.AvgFlightEmission = flightEmission / float64(flightCount)
	}

	if railwayCount > 0 {
		comp.AvgRailwayEmission = railwayEmission / float64(railwayCount)
	}

	// 如果全用铁路,节省的排放
	if flightCount > 0 && comp.AvgFlightEmission > 0 {
		avgDistance := flightEmission / FlightEmissionFactor
		railwayEquivalent := avgDistance * RailwayEmissionFactor
		comp.EmissionSaved = flightEmission - railwayEquivalent
	}

	return comp
}

// calculateCarbonFootprint 计算碳足迹报告
func calculateCarbonFootprint(totalEmission float64, years int) CarbonFootprint {
	footprint := CarbonFootprint{}

	if years > 0 {
		footprint.AnnualAverage = totalEmission / float64(years)
	}

	// 需要种植的树木数量
	footprint.TreesNeeded = int(math.Ceil(totalEmission / TreeAbsorption))

	// 相当于汽车行驶公里数
	footprint.EquivalentCars = totalEmission / CarEmissionFactor

	// 全球百分位(假设全球人均年碳排放4000kg)
	if years > 0 {
		globalAverage := 4000.0
		footprint.GlobalPercentile = (footprint.AnnualAverage / globalAverage) * 100
	}

	return footprint
}

// generateRecommendations 生成碳中和建议
func generateRecommendations(analysis *CarbonEmissionAnalysis) []string {
	recommendations := []string{}

	// 基于航空vs铁路比例
	if analysis.EmissionComparison.FlightVsRailway > 2.0 {
		recommendations = append(recommendations,
			"优先选择高铁出行,短途航班(< 1000km)可改为高铁,碳排放可减少60%以上")
	}

	// 基于年均排放
	if analysis.CarbonFootprint.AnnualAverage > 2000 {
		recommendations = append(recommendations,
			"年均旅行碳排放较高,建议减少非必要航班,或购买碳抵消额度")
	}

	// 基于树木数量
	if analysis.CarbonFootprint.TreesNeeded > 100 {
		recommendations = append(recommendations,
			"考虑参与植树造林项目,抵消旅行碳足迹")
	}

	// 基于全球百分位
	if analysis.CarbonFootprint.GlobalPercentile > 150 {
		recommendations = append(recommendations,
			"旅行碳排放超过全球人均水平50%以上,建议采取碳中和措施")
	}

	// 通用建议
	recommendations = append(recommendations,
		"选择直飞航班,减少中转次数可降低碳排放",
		"优先选择新型节能飞机(如A350, 787)的航班",
		"考虑购买碳抵消额度,支持可再生能源项目",
	)

	return recommendations
}
