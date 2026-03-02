#!/bin/bash
# Efficiency Curve API Test Script
# Tests all 5 API endpoints

BASE_URL="http://localhost:18080/api/v1/cross-module/efficiency-curve"

echo "========================================="
echo "Efficiency Curve API Test"
echo "========================================="
echo ""

# Test 1: Health check
echo "Test 1: Health Check"
curl -s http://localhost:18080/health | jq .
echo ""
echo ""

# Test 2: Get hourly curve (should return empty initially)
echo "Test 2: GET /hourly (last 7 days)"
curl -s "${BASE_URL}/hourly?start_date=$(date -d '7 days ago' +%Y-%m-%d)&end_date=$(date +%Y-%m-%d)" | jq .
echo ""
echo ""

# Test 3: Get workday profile (should return 404 initially)
echo "Test 3: GET /profile?profile_type=workday"
curl -s "${BASE_URL}/profile?profile_type=workday" | jq .
echo ""
echo ""

# Test 4: Get weekend profile (should return 404 initially)
echo "Test 4: GET /profile?profile_type=weekend"
curl -s "${BASE_URL}/profile?profile_type=weekend" | jq .
echo ""
echo ""

# Test 5: Get comparison (should return 404 initially)
echo "Test 5: GET /comparison"
curl -s "${BASE_URL}/comparison" | jq .
echo ""
echo ""

# Test 6: Get insights (should return empty array initially)
echo "Test 6: GET /insights"
curl -s "${BASE_URL}/insights" | jq .
echo ""
echo ""

# Test 7: Trigger analysis (will fail until data fetch methods are implemented)
echo "Test 7: POST /analyze (last 30 days)"
curl -s -X POST "${BASE_URL}/analyze?start_date=$(date -d '30 days ago' +%Y-%m-%d)&end_date=$(date +%Y-%m-%d)" | jq .
echo ""
echo ""

echo "========================================="
echo "Test completed!"
echo "========================================="
echo ""
echo "Expected results:"
echo "- Test 1: Should return {\"status\": \"ok\"}"
echo "- Test 2-6: Should return empty data or 404 (no data yet)"
echo "- Test 7: Will fail until data fetch methods are implemented"
echo ""
echo "Next steps:"
echo "1. Implement fetchKeyboardData(), fetchScreenTimeData(), fetchHealthData()"
echo "2. Run analysis: POST /analyze"
echo "3. Verify data: GET /hourly, /profile, /comparison"
