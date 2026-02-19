#!/usr/bin/env python3
"""
Footprint statistics script - refactored to use database directly.
Generates statistics about visited locations at different administrative levels.

Usage:
    python tracks_v2.py [--start-time 0] [--end-time 0] [--output stats.json]

Options:
    --start-time: Start time as Unix timestamp (0 = all, default: 0)
    --end-time: End time as Unix timestamp (0 = all, default: 0)
    --output: Output file path (default: footprint_stats.json)
"""

import sqlite3
import sys
import argparse
import json
from pathlib import Path
from datetime import datetime
from typing import Dict, List, Tuple
from collections import Counter


class FootprintAnalyzer:
    """Analyzes footprint statistics from GPS track points."""

    def __init__(self, db_path: Path):
        """
        Initialize footprint analyzer.

        Args:
            db_path: Path to SQLite database
        """
        self.db_path = db_path

    def get_track_points(self, start_time: int = 0, end_time: int = 0) -> List[Tuple]:
        """
        Get track points from database.

        Args:
            start_time: Start time as Unix timestamp (0 = all)
            end_time: End time as Unix timestamp (0 = all)

        Returns:
            List of tuples (id, dataTime, province, city, county, town, village)
        """
        conn = sqlite3.connect(str(self.db_path))
        cursor = conn.cursor()

        query = '''
            SELECT id, dataTime, province, city, county, town, village, time
            FROM "一生足迹"
            WHERE 1=1
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

    def analyze_footprint(self, points: List[Tuple]) -> Dict:
        """
        Analyze footprint statistics.

        Args:
            points: List of track points

        Returns:
            Dictionary with statistics
        """
        if not points:
            return {}

        print(f"Analyzing {len(points)} track points...")

        # Initialize counters
        province_counter = Counter()
        city_counter = Counter()
        county_counter = Counter()
        town_counter = Counter()
        village_counter = Counter()

        # Count occurrences
        for point_id, data_time, province, city, county, town, village, time_str in points:
            if province:
                province_counter[province] += 1
            if city:
                city_counter[city] += 1
            if county:
                county_counter[county] += 1
            if town:
                town_counter[town] += 1
            if village:
                village_counter[village] += 1

        # Get unique counts
        unique_provinces = len([p for p in province_counter if p])
        unique_cities = len([c for c in city_counter if c])
        unique_counties = len([c for c in county_counter if c])
        unique_towns = len([t for t in town_counter if t])
        unique_villages = len([v for v in village_counter if v])

        # Get top locations
        top_provinces = province_counter.most_common(20)
        top_cities = city_counter.most_common(30)
        top_counties = county_counter.most_common(20)
        top_towns = town_counter.most_common(20)

        # Time-based statistics
        time_stats = self._analyze_time_distribution(points)

        # Build result
        result = {
            'summary': {
                'total_points': len(points),
                'unique_provinces': unique_provinces,
                'unique_cities': unique_cities,
                'unique_counties': unique_counties,
                'unique_towns': unique_towns,
                'unique_villages': unique_villages,
            },
            'top_provinces': [{'name': name, 'count': count} for name, count in top_provinces],
            'top_cities': [{'name': name, 'count': count} for name, count in top_cities],
            'top_counties': [{'name': name, 'count': count} for name, count in top_counties],
            'top_towns': [{'name': name, 'count': count} for name, count in top_towns],
            'time_distribution': time_stats,
            'generated_at': datetime.now().isoformat(),
        }

        return result

    def _analyze_time_distribution(self, points: List[Tuple]) -> Dict:
        """
        Analyze time distribution statistics.

        Args:
            points: List of track points

        Returns:
            Dictionary with time distribution statistics
        """
        hour_counter = Counter()
        year_counter = Counter()
        month_counter = Counter()
        year_month_counter = Counter()

        for point_id, data_time, province, city, county, town, village, time_str in points:
            if time_str and len(time_str) >= 10:
                # Extract time components from time string (format: YYYYMMDDHHMMSS)
                year = time_str[:4]
                month = time_str[4:6]
                hour = time_str[8:10]
                year_month = time_str[:6]

                year_counter[year] += 1
                month_counter[month] += 1
                hour_counter[hour] += 1
                year_month_counter[year_month] += 1

        return {
            'by_hour': dict(sorted(hour_counter.items())),
            'by_year': dict(sorted(year_counter.items())),
            'by_month': dict(sorted(month_counter.items())),
            'by_year_month': dict(sorted(year_month_counter.items())),
        }

    def save_stats_to_db(self, stats: Dict):
        """
        Save statistics to database.

        Args:
            stats: Statistics dictionary
        """
        conn = sqlite3.connect(str(self.db_path))
        cursor = conn.cursor()

        # Create footprint_stats table if it doesn't exist
        cursor.execute('''
            CREATE TABLE IF NOT EXISTS footprint_stats (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                start_time INTEGER,
                end_time INTEGER,
                total_points INTEGER,
                unique_provinces INTEGER,
                unique_cities INTEGER,
                unique_counties INTEGER,
                unique_towns INTEGER,
                unique_villages INTEGER,
                stats_json TEXT,
                created_at TEXT DEFAULT CURRENT_TIMESTAMP,
                algo_version TEXT DEFAULT '1.0'
            )
        ''')

        # Insert statistics
        cursor.execute('''
            INSERT INTO footprint_stats (
                start_time, end_time, total_points,
                unique_provinces, unique_cities, unique_counties, unique_towns, unique_villages,
                stats_json
            )
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
        ''', (
            0,  # start_time (0 = all)
            0,  # end_time (0 = all)
            stats['summary']['total_points'],
            stats['summary']['unique_provinces'],
            stats['summary']['unique_cities'],
            stats['summary']['unique_counties'],
            stats['summary']['unique_towns'],
            stats['summary']['unique_villages'],
            json.dumps(stats, ensure_ascii=False),
        ))

        conn.commit()
        conn.close()

        print("Statistics saved to database.")

    def print_summary(self, stats: Dict):
        """
        Print statistics summary.

        Args:
            stats: Statistics dictionary
        """
        print("\n" + "=" * 60)
        print("FOOTPRINT STATISTICS SUMMARY")
        print("=" * 60)

        summary = stats['summary']
        print(f"\nTotal Points: {summary['total_points']:,}")
        print(f"Unique Provinces: {summary['unique_provinces']}")
        print(f"Unique Cities: {summary['unique_cities']}")
        print(f"Unique Counties: {summary['unique_counties']}")
        print(f"Unique Towns: {summary['unique_towns']}")
        print(f"Unique Villages: {summary['unique_villages']}")

        print(f"\nTop 10 Provinces:")
        for i, item in enumerate(stats['top_provinces'][:10], 1):
            print(f"  {i}. {item['name']}: {item['count']:,} points")

        print(f"\nTop 10 Cities:")
        for i, item in enumerate(stats['top_cities'][:10], 1):
            print(f"  {i}. {item['name']}: {item['count']:,} points")

        print(f"\nTop 10 Counties:")
        for i, item in enumerate(stats['top_counties'][:10], 1):
            print(f"  {i}. {item['name']}: {item['count']:,} points")

        # Time distribution
        time_dist = stats['time_distribution']
        print(f"\nYearly Distribution:")
        for year, count in sorted(time_dist['by_year'].items()):
            print(f"  {year}: {count:,} points")

        print("\n" + "=" * 60)

    def run(self, start_time: int = 0, end_time: int = 0, output_file: str = None):
        """
        Run footprint analysis.

        Args:
            start_time: Start time as Unix timestamp (0 = all)
            end_time: End time as Unix timestamp (0 = all)
            output_file: Output file path (optional)
        """
        print(f"\nFootprint Analysis")
        print(f"Database: {self.db_path}")

        # Get track points
        points = self.get_track_points(start_time, end_time)
        print(f"Loaded {len(points)} track points")

        if not points:
            print("No track points found.")
            return

        # Analyze footprint
        stats = self.analyze_footprint(points)

        # Print summary
        self.print_summary(stats)

        # Save to database
        self.save_stats_to_db(stats)

        # Save to file if specified
        if output_file:
            output_path = Path(output_file)
            with open(output_path, 'w', encoding='utf-8') as f:
                json.dump(stats, f, ensure_ascii=False, indent=2)
            print(f"\nStatistics saved to: {output_path}")


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(description='Analyze footprint statistics from GPS track points')
    parser.add_argument('--start-time', type=int, default=0,
                        help='Start time as Unix timestamp (0 = all, default: 0)')
    parser.add_argument('--end-time', type=int, default=0,
                        help='End time as Unix timestamp (0 = all, default: 0)')
    parser.add_argument('--output', type=str, default=None,
                        help='Output file path (default: None)')

    args = parser.parse_args()

    # Paths
    script_dir = Path(__file__).parent
    db_path = script_dir.parent.parent / "data" / "tracks" / "tracks.db"

    # Verify database exists
    if not db_path.exists():
        print(f"Error: Database not found: {db_path}")
        sys.exit(1)

    # Run footprint analysis
    analyzer = FootprintAnalyzer(db_path)
    analyzer.run(start_time=args.start_time, end_time=args.end_time, output_file=args.output)


if __name__ == "__main__":
    main()
