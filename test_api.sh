#!/bin/bash
echo "=== Testing Backend APIs on port 18080 ==="
echo ""

# Test 1: Devices
echo "Test 1: List Devices"
wget -q -O - http://localhost:18080/api/v1/screentime/devices 2>&1 | head -20
echo ""
echo "---"

# Test 2: Cross-device comparison
echo "Test 2: Cross-Device Comparison"
wget -q -O - http://localhost:18080/api/v1/screentime/cross-device/comparison 2>&1 | head -30
echo ""
echo "---"

# Test 3: Work-life balance
echo "Test 3: Work-Life Balance"
wget -q -O - http://localhost:18080/api/v1/screentime/cross-device/work-life-balance 2>&1 | head -30
echo ""
