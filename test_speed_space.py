#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""Test script for speed-space coupling analyzer"""

import requests
import json
import time

BASE_URL = "http://localhost:8080/api/v1"

def create_analysis_task(skill_name, mode="full"):
    """Create an analysis task"""
    url = f"{BASE_URL}/admin/analysis/tasks"
    payload = {
        "skill_name": skill_name,
        "mode": mode
    }

    print(f"Creating {skill_name} task...")
    response = requests.post(url, json=payload)
    print(f"Status: {response.status_code}")

    if response.status_code == 200 or response.status_code == 201:
        try:
            data = response.json()
            print(f"Response: {json.dumps(data, indent=2)}")
            return data.get("data", {}).get("id")
        except:
            print(f"Response text: {response.text}")
            return None
    else:
        print(f"Error: {response.text}")
        return None

def get_task_status(task_id):
    """Get task status"""
    url = f"{BASE_URL}/admin/analysis/tasks/{task_id}"
    response = requests.get(url)

    if response.status_code == 200:
        try:
            data = response.json()
            return data.get("data", {})
        except:
            return None
    return None

def wait_for_task(task_id, timeout=60):
    """Wait for task to complete"""
    print(f"Waiting for task {task_id} to complete...")
    start_time = time.time()

    while time.time() - start_time < timeout:
        status = get_task_status(task_id)
        if status:
            print(f"Status: {status.get('status')}, Progress: {status.get('processed_points')}/{status.get('total_points')}")

            if status.get("status") == "completed":
                print("Task completed!")
                print(f"Summary: {status.get('result_summary')}")
                return True
            elif status.get("status") == "failed":
                print(f"Task failed: {status.get('error_message')}")
                return False

        time.sleep(2)

    print("Timeout waiting for task")
    return False

def query_speed_space_stats():
    """Query speed-space statistics"""
    url = f"{BASE_URL}/stats/speed-space"
    params = {"limit": 10}

    print("\nQuerying speed-space stats...")
    response = requests.get(url, params=params)
    print(f"Status: {response.status_code}")

    if response.status_code == 200:
        try:
            data = response.json()
            print(f"Results: {json.dumps(data, indent=2)}")
        except:
            print(f"Response text: {response.text}")
    else:
        print(f"Error: {response.text}")

def query_high_speed_zones():
    """Query high-speed zones"""
    url = f"{BASE_URL}/stats/speed-space/high-speed-zones"
    params = {"limit": 10}

    print("\nQuerying high-speed zones...")
    response = requests.get(url, params=params)
    print(f"Status: {response.status_code}")

    if response.status_code == 200:
        try:
            data = response.json()
            print(f"Results: {json.dumps(data, indent=2)}")
        except:
            print(f"Response text: {response.text}")
    else:
        print(f"Error: {response.text}")

if __name__ == "__main__":
    # Create speed_space_coupling task
    task_id = create_analysis_task("speed_space_coupling", "full")

    if task_id:
        # Wait for completion
        if wait_for_task(task_id):
            # Query results
            query_speed_space_stats()
            query_high_speed_zones()
    else:
        print("Failed to create task")
