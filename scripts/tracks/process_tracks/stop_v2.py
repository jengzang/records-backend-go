#!/usr/bin/env python3
"""
Stay detection script - refactored to use database directly.
Detects stays longer than a configurable threshold at the same location.

Usage:
    python stop_v2.py [--threshold 2] [--admin-level town] [--start-time 0] [--end-time 0]

Options:
    --threshold: Minimum stay duration in hours (default: 2)
    --admin-level: Administrative level for location matching (city/county/town, default: town)
    --start-time: Start time as Unix timestamp (0 = all, default: 0)
    --end-time: End time as Unix timestamp (0 = all, default: 0)
"""

import sqlite3
import sys
import argparse
from pathlib import Path
from datetime import datetime
from typing import List, Tuple, Optional


class StayDetector:
    """Detects stays from GPS track points."""

    def __init__(self, db_path: Path, threshold_hours: float = 2.0, admin_level: str = 'town'):
        """
        Initialize stay detector.

        Args:
            db_path: Path to SQLite database
            threshold_hours: Minimum stay duration in hours
            admin_level: Administrative level for location matching (city/county/town)
        """
        self.db_path = db_path
        self.threshold_seconds = threshold_hours * 3600
        self.admin_level = admin_level

        # Map admin level to column names
        self.level_columns = {
            'city': ['city'],
            'county': ['city', 'county'],
            'town': ['city', 'county', 'town'],
        }

        if admin_level not in self.level_columns:
            raise ValueError(f"Invalid admin level: {admin_level}. Must be one of: city, county, town")

    def get_track_points(self, start_time: int = 0, end_time: int = 0) -> List[Tuple]:
        """
        Get track points from database.

        Args:
            start_time: Start time as Unix timestamp (0 = all)
            end_time: End time as Unix timestamp (0 = all)

        Returns:
            List of tuples (id, dataTime, city, county, town)
        """
        conn = sqlite3.connect(str(self.db_path))
        cursor = conn.cursor()

        query = '''
            SELECT id, dataTime, city, county, town
            FROM "一生足迹"
            WHERE city IS NOT NULL AND city != ''
        '''

        params = []
        if start_time > 0:
            query += ' AND dataTime >= ?'
            params.append(start_time)
        if end_time > 0:
            query += ' AND dataTime <= ?'
            params.append(end_time)

        query += ' ORDER BY dataTime ASC'

        cursor.execute(query, params)
        points = cursor.fetchall()
        conn.close()

        return points

    def get_location_key(self, city: str, county: str, town: str) -> Tuple:
        """
        Get location key based on admin level.

        Args:
            city: City name
            county: County name
            town: Town name

        Returns:
            Tuple representing the location at the specified admin level
        """
        if self.admin_level == 'city':
            return (city,)
        elif self.admin_level == 'county':
            return (city, county)
        else:  # town
            return (city, county, town)

    def detect_stays(self, points: List[Tuple]) -> List[dict]:
        """
        Detect stays from track points.

        Args:
            points: List of track points

        Returns:
            List of stay dictionaries
        """
        if not points:
            return []

        stays = []
        stay_group = 0
        stay_start_time = None
        stay_end_time = None
        stay_duration = 0
        previous_location = None
        previous_time = None
        point_count = 0

        print(f"Processing {len(points)} track points...")

        for point_id, data_time, city, county, town in points:
            # Skip points with missing admin data
            if not city or (self.admin_level in ['county', 'town'] and not county) or (self.admin_level == 'town' and not town):
                # If we were tracking a stay, save it if it meets threshold
                if stay_duration >= self.threshold_seconds and stay_start_time:
                    stays.append({
                        'stay_group': stay_group,
                        'start_time': stay_start_time,
                        'end_time': stay_end_time,
                        'duration': stay_duration,
                        'location': previous_location,
                        'point_count': point_count,
                    })
                    stay_group += 1

                # Reset tracking
                stay_start_time = None
                stay_duration = 0
                previous_location = None
                previous_time = None
                point_count = 0
                continue

            current_location = self.get_location_key(city, county, town)
            current_time = data_time

            # Calculate time difference
            if previous_time is not None:
                time_diff = current_time - previous_time
            else:
                time_diff = 0

            # Check if same location
            if previous_location == current_location:
                stay_duration += time_diff
                stay_end_time = current_time
                point_count += 1
            else:
                # Location changed - save previous stay if it meets threshold
                if stay_duration >= self.threshold_seconds and stay_start_time:
                    stays.append({
                        'stay_group': stay_group,
                        'start_time': stay_start_time,
                        'end_time': stay_end_time,
                        'duration': stay_duration,
                        'location': previous_location,
                        'point_count': point_count,
                    })
                    stay_group += 1

                # Start new stay
                stay_start_time = current_time
                stay_end_time = current_time
                stay_duration = 0
                point_count = 1

            # Update tracking variables
            previous_time = current_time
            previous_location = current_location

        # Save last stay if it meets threshold
        if stay_duration >= self.threshold_seconds and stay_start_time:
            stays.append({
                'stay_group': stay_group,
                'start_time': stay_start_time,
                'end_time': stay_end_time,
                'duration': stay_duration,
                'location': previous_location,
                'point_count': point_count,
            })

        return stays

    def save_stays_to_db(self, stays: List[dict]):
        """
        Save detected stays to database.

        Args:
            stays: List of stay dictionaries
        """
        if not stays:
            print("No stays to save.")
            return

        conn = sqlite3.connect(str(self.db_path))
        cursor = conn.cursor()

        # Create stays table if it doesn't exist
        cursor.execute('''
            CREATE TABLE IF NOT EXISTS stays (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                stay_group INTEGER,
                start_time INTEGER,
                end_time INTEGER,
                duration INTEGER,
                city TEXT,
                county TEXT,
                town TEXT,
                point_count INTEGER,
                admin_level TEXT,
                created_at TEXT DEFAULT CURRENT_TIMESTAMP,
                algo_version TEXT DEFAULT '1.0'
            )
        ''')

        # Insert stays
        for stay in stays:
            location = stay['location']
            city = location[0] if len(location) > 0 else None
            county = location[1] if len(location) > 1 else None
            town = location[2] if len(location) > 2 else None

            cursor.execute('''
                INSERT INTO stays (stay_group, start_time, end_time, duration, city, county, town, point_count, admin_level)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
            ''', (
                stay['stay_group'],
                stay['start_time'],
                stay['end_time'],
                stay['duration'],
                city,
                county,
                town,
                stay['point_count'],
                self.admin_level,
            ))

        conn.commit()
        conn.close()

        print(f"Saved {len(stays)} stays to database.")

    def run(self, start_time: int = 0, end_time: int = 0):
        """
        Run stay detection process.

        Args:
            start_time: Start time as Unix timestamp (0 = all)
            end_time: End time as Unix timestamp (0 = all)
        """
        print(f"\nStay Detection")
        print(f"Database: {self.db_path}")
        print(f"Threshold: {self.threshold_seconds / 3600:.1f} hours")
        print(f"Admin level: {self.admin_level}")

        # Get track points
        points = self.get_track_points(start_time, end_time)
        print(f"Loaded {len(points)} track points")

        if not points:
            print("No track points found.")
            return

        # Detect stays
        stays = self.detect_stays(points)
        print(f"Detected {len(stays)} stays")

        # Print summary
        if stays:
            total_duration = sum(s['duration'] for s in stays)
            avg_duration = total_duration / len(stays)
            max_duration = max(s['duration'] for s in stays)

            print(f"\nStay Statistics:")
            print(f"  Total stays: {len(stays)}")
            print(f"  Total duration: {total_duration / 3600:.1f} hours")
            print(f"  Average duration: {avg_duration / 3600:.1f} hours")
            print(f"  Max duration: {max_duration / 3600:.1f} hours")

            # Print top 10 longest stays
            print(f"\nTop 10 longest stays:")
            sorted_stays = sorted(stays, key=lambda x: x['duration'], reverse=True)[:10]
            for i, stay in enumerate(sorted_stays, 1):
                start_dt = datetime.fromtimestamp(stay['start_time']).strftime('%Y-%m-%d %H:%M')
                end_dt = datetime.fromtimestamp(stay['end_time']).strftime('%Y-%m-%d %H:%M')
                duration_hours = stay['duration'] / 3600
                print(f"  {i}. {stay['location']} - {duration_hours:.1f}h ({start_dt} to {end_dt})")

        # Save to database
        self.save_stays_to_db(stays)


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(description='Detect stays from GPS track points')
    parser.add_argument('--threshold', type=float, default=2.0,
                        help='Minimum stay duration in hours (default: 2.0)')
    parser.add_argument('--admin-level', type=str, default='town', choices=['city', 'county', 'town'],
                        help='Administrative level for location matching (default: town)')
    parser.add_argument('--start-time', type=int, default=0,
                        help='Start time as Unix timestamp (0 = all, default: 0)')
    parser.add_argument('--end-time', type=int, default=0,
                        help='End time as Unix timestamp (0 = all, default: 0)')

    args = parser.parse_args()

    # Paths
    script_dir = Path(__file__).parent
    db_path = script_dir.parent.parent / "data" / "tracks" / "tracks.db"

    # Verify database exists
    if not db_path.exists():
        print(f"Error: Database not found: {db_path}")
        sys.exit(1)

    # Run stay detection
    detector = StayDetector(db_path, threshold_hours=args.threshold, admin_level=args.admin_level)
    detector.run(start_time=args.start_time, end_time=args.end_time)


if __name__ == "__main__":
    main()
