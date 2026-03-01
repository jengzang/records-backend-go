#!/usr/bin/env python3
"""
Import Variflight JSON data into flights database

Usage:
    python import_variflight.py <json_file>

Example:
    python import_variflight.py Variflight_CA1332_20250104.json
"""

import json
import sqlite3
import sys
import re
from datetime import datetime
from pathlib import Path


def parse_flight_number(filename):
    """Extract flight number and date from filename"""
    # Pattern: Variflight_CA1332_20250104.json
    match = re.search(r'Variflight_([A-Z0-9]+)_(\d{8})\.json', filename)
    if match:
        return match.group(1), match.group(2)
    return None, None


def import_variflight_data(json_file, db_path):
    """Import Variflight JSON data into database"""

    # Parse filename
    flight_number, flight_date = parse_flight_number(Path(json_file).name)
    if not flight_number:
        print(f"Error: Could not parse flight number from filename: {json_file}")
        return False

    print(f"Importing flight {flight_number} on {flight_date}")

    # Load JSON data
    try:
        with open(json_file, 'r', encoding='utf-8') as f:
            data = json.load(f)
    except Exception as e:
        print(f"Error loading JSON file: {e}")
        return False

    # Connect to database
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()

    try:
        # Extract flight info
        aircraft_number = data.get('anum', '')
        points = data.get('points', [])

        if not points:
            print("Warning: No tracking points found in JSON")
            return False

        # Determine airline from flight number (first 2 letters)
        airline_code = flight_number[:2]
        airline_map = {
            'CA': 'Air China',
            'CZ': 'China Southern',
            'MU': 'China Eastern',
            'HU': 'Hainan Airlines',
            'KN': 'China United Airlines',
            '3U': 'Sichuan Airlines',
            'ZH': 'Shenzhen Airlines',
        }
        airline = airline_map.get(airline_code, airline_code)

        # Insert flight record
        cursor.execute('''
            INSERT INTO flights (
                flight_number, aircraft_number, airline,
                flight_date, data_source
            ) VALUES (?, ?, ?, ?, ?)
            ON CONFLICT(flight_number, flight_date) DO UPDATE SET
                aircraft_number = excluded.aircraft_number,
                airline = excluded.airline
        ''', (flight_number, aircraft_number, airline, flight_date, 'variflight'))

        flight_id = cursor.lastrowid
        if flight_id == 0:
            # Get existing flight ID
            cursor.execute(
                'SELECT id FROM flights WHERE flight_number = ? AND flight_date = ?',
                (flight_number, flight_date)
            )
            flight_id = cursor.fetchone()[0]

        print(f"Flight ID: {flight_id}")

        # Delete existing points for this flight
        cursor.execute('DELETE FROM flight_points WHERE flight_id = ?', (flight_id,))

        # Insert tracking points
        point_count = 0
        for point in points:
            update_time = point.get('updatetime', 0)
            utc_time = point.get('UTC Time', '')
            longitude = point.get('longitude', 0)
            latitude = point.get('latitude', 0)
            altitude = point.get('height', 0)  # meters
            speed = point.get('speed', 0)      # km/h
            heading = point.get('angle', 0)    # degrees

            if longitude and latitude:
                cursor.execute('''
                    INSERT INTO flight_points (
                        flight_id, update_time, utc_time,
                        longitude, latitude, altitude, speed, heading
                    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
                ''', (flight_id, update_time, utc_time, longitude, latitude,
                      altitude, speed, heading))
                point_count += 1

        print(f"Imported {point_count} tracking points")

        # Calculate statistics
        if point_count > 0:
            cursor.execute('''
                SELECT
                    MAX(altitude) as max_altitude,
                    MAX(speed) as max_speed,
                    AVG(speed) as avg_speed,
                    COUNT(*) as point_count
                FROM flight_points
                WHERE flight_id = ?
            ''', (flight_id,))

            max_alt, max_speed, avg_speed, count = cursor.fetchone()

            # Calculate total distance (simplified - sum of point-to-point distances)
            cursor.execute('''
                SELECT longitude, latitude, update_time
                FROM flight_points
                WHERE flight_id = ?
                ORDER BY update_time ASC
            ''', (flight_id,))

            points_data = cursor.fetchall()
            total_distance = 0
            for i in range(1, len(points_data)):
                lon1, lat1, _ = points_data[i-1]
                lon2, lat2, _ = points_data[i]
                dist = haversine_distance(lat1, lon1, lat2, lon2)
                total_distance += dist

            # Calculate duration
            first_time = points_data[0][2]
            last_time = points_data[-1][2]
            duration_minutes = (last_time - first_time) // 60

            # Insert/update statistics
            cursor.execute('''
                INSERT INTO flight_statistics (
                    flight_id, total_distance, max_altitude, max_speed,
                    avg_speed, duration_minutes, point_count, updated_at
                ) VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
                ON CONFLICT(flight_id) DO UPDATE SET
                    total_distance = excluded.total_distance,
                    max_altitude = excluded.max_altitude,
                    max_speed = excluded.max_speed,
                    avg_speed = excluded.avg_speed,
                    duration_minutes = excluded.duration_minutes,
                    point_count = excluded.point_count,
                    updated_at = CURRENT_TIMESTAMP
            ''', (flight_id, total_distance, max_alt, max_speed,
                  avg_speed, duration_minutes, count))

            print(f"Statistics: {total_distance:.1f}km, {max_alt:.0f}m altitude, "
                  f"{max_speed:.0f}km/h max speed, {duration_minutes}min duration")

        conn.commit()
        print("Import successful!")
        return True

    except Exception as e:
        conn.rollback()
        print(f"Error importing data: {e}")
        import traceback
        traceback.print_exc()
        return False
    finally:
        conn.close()


def haversine_distance(lat1, lon1, lat2, lon2):
    """Calculate distance between two GPS coordinates in kilometers"""
    import math

    R = 6371.0  # Earth radius in km

    lat1_rad = math.radians(lat1)
    lat2_rad = math.radians(lat2)
    delta_lat = math.radians(lat2 - lat1)
    delta_lon = math.radians(lon2 - lon1)

    a = (math.sin(delta_lat / 2) ** 2 +
         math.cos(lat1_rad) * math.cos(lat2_rad) *
         math.sin(delta_lon / 2) ** 2)

    c = 2 * math.atan2(math.sqrt(a), math.sqrt(1 - a))

    return R * c


def main():
    if len(sys.argv) < 2:
        print("Usage: python import_variflight.py <json_file>")
        print("Example: python import_variflight.py Variflight_CA1332_20250104.json")
        sys.exit(1)

    json_file = sys.argv[1]
    if not Path(json_file).exists():
        print(f"Error: File not found: {json_file}")
        sys.exit(1)

    # Database path
    db_path = Path(__file__).parent.parent.parent / 'data' / 'flights' / 'flights.db'
    db_path.parent.mkdir(parents=True, exist_ok=True)

    # Create tables if they don't exist
    conn = sqlite3.connect(str(db_path))
    with open(Path(__file__).parent / 'migrations' / '001_create_flight_tables.sql', 'r') as f:
        conn.executescript(f.read())
    conn.close()

    # Import data
    success = import_variflight_data(json_file, str(db_path))
    sys.exit(0 if success else 1)


if __name__ == '__main__':
    main()
