#!/usr/bin/env python3
"""
Test Efficiency Analysis End-to-End
"""
import requests
import json
from datetime import datetime, timedelta

BASE_URL = "http://localhost:18080/api/v1/cross-module/efficiency-curve"

def test_health_check():
    """Test server health"""
    print("=" * 60)
    print("Test 1: Health Check")
    print("=" * 60)
    response = requests.get("http://localhost:18080/health")
    print(f"Status: {response.status_code}")
    print(f"Response: {response.json()}")
    print()

def test_trigger_analysis():
    """Trigger efficiency analysis"""
    print("=" * 60)
    print("Test 2: Trigger Analysis")
    print("=" * 60)

    # Analyze last 7 days
    end_date = datetime.now().strftime("%Y-%m-%d")
    start_date = (datetime.now() - timedelta(days=7)).strftime("%Y-%m-%d")

    url = f"{BASE_URL}/analyze?start_date={start_date}&end_date={end_date}"
    print(f"POST {url}")

    response = requests.post(url)
    print(f"Status: {response.status_code}")
    print(f"Response: {json.dumps(response.json(), indent=2)}")
    print()

    return response.status_code == 200

def test_get_hourly_curve():
    """Get hourly efficiency curve"""
    print("=" * 60)
    print("Test 3: Get Hourly Curve")
    print("=" * 60)

    end_date = datetime.now().strftime("%Y-%m-%d")
    start_date = (datetime.now() - timedelta(days=7)).strftime("%Y-%m-%d")

    url = f"{BASE_URL}/hourly?start_date={start_date}&end_date={end_date}"
    print(f"GET {url}")

    response = requests.get(url)
    print(f"Status: {response.status_code}")

    if response.status_code == 200:
        data = response.json()
        print(f"Total hours: {data['stats']['total_hours']}")
        print(f"Avg efficiency: {data['stats']['avg_efficiency']:.2f}")
        print(f"Max efficiency: {data['stats']['max_efficiency']:.2f}")
        print(f"Min efficiency: {data['stats']['min_efficiency']:.2f}")
        print(f"Data completeness: {data['stats']['data_completeness']:.2%}")

        if data['scores']:
            print(f"\nSample scores (first 3):")
            for score in data['scores'][:3]:
                print(f"  {score['date']} {score['hour']:02d}:00 - Efficiency: {score['efficiency_score']:.2f}, Completeness: {score['data_completeness']:.2%}")
    else:
        print(f"Error: {response.text}")
    print()

def test_get_profiles():
    """Get efficiency profiles"""
    print("=" * 60)
    print("Test 4: Get Profiles")
    print("=" * 60)

    for profile_type in ['workday', 'weekend']:
        url = f"{BASE_URL}/profile?profile_type={profile_type}"
        print(f"GET {url}")

        response = requests.get(url)
        print(f"Status: {response.status_code}")

        if response.status_code == 200:
            data = response.json()
            print(f"Profile Type: {data['profile_type']}")
            print(f"Peak Hour: {data['peak_hour']}:00 (Score: {data['peak_score']:.2f})")
            print(f"Peak Period: {data['peak_start_hour']}:00 - {data['peak_end_hour']}:00")
            print(f"Chronotype: {data['chronotype']} (Confidence: {data['chronotype_confidence']:.2%})")
            print(f"Avg Efficiency: {data['avg_efficiency']:.2f}")
            print(f"Total Samples: {data['total_samples']}")
        else:
            print(f"Error: {response.text}")
        print()

def test_get_comparison():
    """Get workday vs weekend comparison"""
    print("=" * 60)
    print("Test 5: Get Comparison")
    print("=" * 60)

    url = f"{BASE_URL}/comparison"
    print(f"GET {url}")

    response = requests.get(url)
    print(f"Status: {response.status_code}")

    if response.status_code == 200:
        data = response.json()
        print(f"Workday Avg: {data['workday']['avg_efficiency']:.2f}")
        print(f"Weekend Avg: {data['weekend']['avg_efficiency']:.2f}")
        print(f"Difference: {data['diff']['avg_diff']:.2f}")
        print(f"Peak Hour Diff: {data['diff']['peak_hour_diff']} hours")
        print(f"Interpretation: {data['diff']['interpretation']}")
    else:
        print(f"Error: {response.text}")
    print()

def test_get_insights():
    """Get efficiency insights"""
    print("=" * 60)
    print("Test 6: Get Insights")
    print("=" * 60)

    url = f"{BASE_URL}/insights"
    print(f"GET {url}")

    response = requests.get(url)
    print(f"Status: {response.status_code}")

    if response.status_code == 200:
        insights = response.json()
        print(f"Total insights: {len(insights)}")
        for insight in insights:
            print(f"\n[{insight['insight_type']}] Priority: {insight['priority']}")
            print(f"  Title: {insight['title']}")
            print(f"  Description: {insight['description']}")
            if insight.get('recommendation'):
                print(f"  Recommendation: {insight['recommendation']}")
            print(f"  Confidence: {insight['confidence']:.2%}")
    else:
        print(f"Error: {response.text}")
    print()

def main():
    print("\n" + "=" * 60)
    print("Efficiency Curve End-to-End Test")
    print("=" * 60 + "\n")

    try:
        # Test 1: Health check
        test_health_check()

        # Test 2: Trigger analysis
        analysis_success = test_trigger_analysis()

        if analysis_success:
            # Test 3: Get hourly curve
            test_get_hourly_curve()

            # Test 4: Get profiles
            test_get_profiles()

            # Test 5: Get comparison
            test_get_comparison()

            # Test 6: Get insights
            test_get_insights()
        else:
            print("⚠️  Analysis failed, skipping subsequent tests")

        print("=" * 60)
        print("✅ All tests completed!")
        print("=" * 60)

    except requests.exceptions.ConnectionError:
        print("❌ Error: Cannot connect to server at http://localhost:18080")
        print("   Please make sure the server is running:")
        print("   cd go-backend && go run cmd/server/main.go")
    except Exception as e:
        print(f"❌ Error: {e}")

if __name__ == "__main__":
    main()
