package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/jengzang/records-backend-go/internal/health/efficiency"
	_ "modernc.org/sqlite"
)

func main() {
	// Open databases
	healthDB, err := sql.Open("sqlite", "data/applehealth/health.db")
	if err != nil {
		log.Fatalf("Failed to open health database: %v", err)
	}
	defer healthDB.Close()

	keyboardDB, err := sql.Open("sqlite", "data/keyboard/kmcounter.db")
	if err != nil {
		log.Fatalf("Failed to open keyboard database: %v", err)
	}
	defer keyboardDB.Close()

	screentimeDB, err := sql.Open("sqlite", "data/screentime/screentime.db")
	if err != nil {
		log.Fatalf("Failed to open screentime database: %v", err)
	}
	defer screentimeDB.Close()

	// Create service
	service := efficiency.NewService(healthDB, keyboardDB, screentimeDB)

	// Test date and hour
	testDate := "2026-01-26"
	testHour := 10

	fmt.Println("========================================")
	fmt.Printf("Testing data fetchers for %s, hour %d\n", testDate, testHour)
	fmt.Println("========================================\n")

	// Test keyboard data
	fmt.Println("1. Testing Keyboard Data Fetcher:")
	keyboardData, err := service.FetchKeyboardDataTest(testDate, testHour)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else if keyboardData == nil {
		fmt.Println("   No keyboard data available")
	} else {
		fmt.Printf("   Typing Speed: %.2f keystrokes/hour\n", keyboardData.TypingSpeed)
		fmt.Printf("   Normalized Score: %.2f/100\n", keyboardData.TypingSpeedNormalized)
	}
	fmt.Println()

	// Test screentime data
	fmt.Println("2. Testing ScreenTime Data Fetcher:")
	screentimeData, err := service.FetchScreenTimeDataTest(testDate, testHour)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else if screentimeData == nil {
		fmt.Println("   No screentime data available")
	} else {
		fmt.Printf("   Work App Ratio: %.2f%%\n", screentimeData.WorkAppRatio*100)
		fmt.Printf("   Entertainment Ratio: %.2f%%\n", screentimeData.EntertainmentAppRatio*100)
		fmt.Printf("   Focus Sessions: %d\n", screentimeData.FocusSessionCount)
		fmt.Printf("   App Switch Frequency: %.2f switches/hour\n", screentimeData.AppSwitchFrequency)
		fmt.Printf("   Work Ratio Normalized: %.2f/100\n", screentimeData.WorkAppRatioNormalized)
		fmt.Printf("   Focus Normalized: %.2f/100\n", screentimeData.FocusNormalized)
	}
	fmt.Println()

	// Test health data
	fmt.Println("3. Testing Health Data Fetcher:")
	healthData, err := service.FetchHealthDataTest(testDate, testHour)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else if healthData == nil {
		fmt.Println("   No health data available")
	} else {
		fmt.Printf("   Avg Heart Rate: %.2f bpm\n", healthData.AvgHeartRate)
		fmt.Printf("   Heart Rate Variability: %.2f ms\n", healthData.HeartRateVariability)
		fmt.Printf("   Step Count: %d steps\n", healthData.StepCount)
		fmt.Printf("   HRV Normalized: %.2f/100\n", healthData.HRVNormalized)
		fmt.Printf("   Activity Normalized: %.2f/100\n", healthData.ActivityNormalized)
	}
	fmt.Println()

	fmt.Println("========================================")
	fmt.Println("Test completed!")
	fmt.Println("========================================")
}
