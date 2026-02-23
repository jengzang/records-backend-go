#!/usr/bin/env python3
"""
Test keyboard API endpoints
"""

import requests
import json

BASE_URL = "http://localhost:8080/api/keyboard"

def test_endpoint(name, url, params=None):
    """Test an API endpoint"""
    print(f"\n{'='*60}")
    print(f"Testing: {name}")
    print(f"URL: {url}")
    if params:
        print(f"Params: {params}")
    print('='*60)

    try:
        response = requests.get(url, params=params, timeout=10)
        print(f"Status: {response.status_code}")

        if response.status_code == 200:
            data = response.json()
            print(f"Response: {json.dumps(data, indent=2, ensure_ascii=False)[:500]}...")
            return True
        else:
            print(f"Error: {response.text}")
            return False

    except Exception as e:
        print(f"Exception: {e}")
        return False

def main():
    """Run all tests"""
    print("Keyboard API Test Suite")
    print("="*60)

    tests = [
        ("Health Check", "http://localhost:8080/health", None),
        ("Summary Stats", f"{BASE_URL}/statistics/summary", None),
        ("Daily Stats (Last 10)", f"{BASE_URL}/daily", {"limit": "10"}),
        ("Daily Stats (Date Range)", f"{BASE_URL}/daily", {"start": "20250901", "end": "20250930", "limit": "30"}),
        ("Top 10 Keys", f"{BASE_URL}/top-keys", {"limit": "10"}),
        ("Scancode Stats", f"{BASE_URL}/scancodes", {"date": "20250926"}),
        ("Trends (Daily)", f"{BASE_URL}/statistics/trends", {"start": "20250901", "end": "20250930"}),
        ("Trends (Weekly)", f"{BASE_URL}/statistics/trends", {"granularity": "weekly", "start": "20250801", "end": "20250930"}),
        ("Keyboard Heatmap", f"{BASE_URL}/heatmap/keyboard", {"start": "20250901", "end": "20250930"}),
    ]

    results = []
    for name, url, params in tests:
        success = test_endpoint(name, url, params)
        results.append((name, success))

    # Print summary
    print("\n" + "="*60)
    print("TEST SUMMARY")
    print("="*60)
    passed = sum(1 for _, success in results if success)
    total = len(results)
    print(f"Passed: {passed}/{total}")

    for name, success in results:
        status = "✓ PASS" if success else "✗ FAIL"
        print(f"  {status}: {name}")

if __name__ == '__main__':
    main()
