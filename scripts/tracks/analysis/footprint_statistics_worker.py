"""
Footprint Statistics Worker

Aggregates trajectory data by administrative regions to generate footprint rankings.
Supports multiple statistical types and time ranges.

Statistical Types:
- PROVINCE: Province-level statistics
- CITY: City-level statistics
- COUNTY: County-level statistics
- TOWN: Town-level statistics
- GRID: Grid-level statistics (based on grid_id)

Time Ranges:
- all: All time data
- year: Aggregated by year (YYYY)
- month: Aggregated by month (YYYY-MM)
- day: Aggregated by day (YYYY-MM-DD)

Metrics:
- point_count: Number of trajectory points
- visit_count: Number of unique visit dates
- total_distance_m: Total distance traveled (meters)
- total_duration_s: Total time spent (seconds)
- first_visit: First visit timestamp
- last_visit: Last visit timestamp
"""

import sys
import argparse
import json
from typing import List, Tuple, Dict, Any
from datetime import datetime
from collections import defaultdict

# Add parent directory to path for imports
sys.path.append('/app/scripts/common')
from incremental_analyzer import IncrementalAnalyzer


class FootprintStatisticsWorker(IncrementalAnalyzer):
    """Worker for footprint statistics aggregation"""

    STAT_TYPES = ['PROVINCE', 'CITY', 'COUNTY', 'TOWN', 'GRID']
    TIME_RANGES = {
        'all': None,
        'year': '%Y',
        'month': '%Y-%m',
        'day': '%Y-%m-%d'
    }

    def __init__(self, db_path: str, task_id: int, batch_size: int = 1000):
        super().__init__(db_path, task_id, batch_size)
        self.stats_buffer = defaultdict(lambda: {
            'point_count': 0,
            'visit_dates': set(),
            'total_distance_m': 0,
            'total_duration_s': 0,
            'first_visit': None,
            'last_visit': None
        })

    def get_stat_key(self, point: Tuple, stat_type: str) -> str:
        """
        Get statistical key for a point based on stat_type

        Args:
            point: Point tuple
            stat_type: Statistical type (PROVINCE/CITY/COUNTY/TOWN/GRID)

        Returns:
            Statistical key (e.g., "广东省", "深圳市", "L3_ws10g")
        """
        if stat_type == 'PROVINCE':
            return point[9] if len(point) > 9 else None  # province
        elif stat_type == 'CITY':
            return point[10] if len(point) > 10 else None  # city
        elif stat_type == 'COUNTY':
            return point[11] if len(point) > 11 else None  # county
        elif stat_type == 'TOWN':
            return point[12] if len(point) > 12 else None  # town
        elif stat_type == 'GRID':
            # Get grid_id from database
            return self.get_grid_id(point[0])
        return None

    def get_grid_id(self, point_id: int) -> str:
        """Get grid_id for a point"""
        cursor = self.conn.execute(
            'SELECT grid_id FROM "一生足迹" WHERE id = ?',
            (point_id,)
        )
        row = cursor.fetchone()
        return row[0] if row else None

    def get_segment_duration(self, point_id: int) -> int:
        """
        Get duration for a point from segments table

        Args:
            point_id: Point ID

        Returns:
            Duration in seconds (0 if not found)
        """
        cursor = self.conn.execute(
            '''
            SELECT SUM(duration_s)
            FROM segments
            WHERE start_point_id <= ? AND end_point_id >= ?
            ''',
            (point_id, point_id)
        )
        row = cursor.fetchone()
        return int(row[0]) if row and row[0] else 0

    def process_batch(self, points: List[Tuple]) -> int:
        """
        Process a batch of points and aggregate statistics

        Args:
            points: List of point tuples

        Returns:
            Number of points that failed processing
        """
        failed = 0

        for point in points:
            try:
                # Extract point data
                point_id = point[0]
                dataTime = point[1]
                distance = point[7] if len(point) > 7 else 0

                # Skip outlier points
                outlier_flag = self.get_outlier_flag(point_id)
                if outlier_flag:
                    continue

                # Get visit date
                visit_date = datetime.fromtimestamp(dataTime).strftime('%Y-%m-%d')

                # Get segment duration
                duration_s = self.get_segment_duration(point_id)

                # Aggregate for each stat type and time range
                for stat_type in self.STAT_TYPES:
                    stat_key = self.get_stat_key(point, stat_type)
                    if not stat_key:
                        continue

                    for time_range_type, time_format in self.TIME_RANGES.items():
                        # Calculate time key
                        if time_format:
                            time_key = datetime.fromtimestamp(dataTime).strftime(time_format)
                        else:
                            time_key = 'all'

                        # Create composite key
                        composite_key = (stat_type, stat_key, time_key)

                        # Aggregate statistics
                        stats = self.stats_buffer[composite_key]
                        stats['point_count'] += 1
                        stats['visit_dates'].add(visit_date)
                        stats['total_distance_m'] += distance if distance else 0
                        stats['total_duration_s'] += duration_s

                        # Update time range
                        if stats['first_visit'] is None or dataTime < stats['first_visit']:
                            stats['first_visit'] = dataTime
                        if stats['last_visit'] is None or dataTime > stats['last_visit']:
                            stats['last_visit'] = dataTime

            except Exception as e:
                self.logger.error(f"Failed to process point {point[0]}: {e}")
                failed += 1

        # Flush statistics to database every batch
        self.flush_statistics()

        return failed

    def get_outlier_flag(self, point_id: int) -> bool:
        """Get outlier flag for a point"""
        cursor = self.conn.execute(
            'SELECT outlier_flag FROM "一生足迹" WHERE id = ?',
            (point_id,)
        )
        row = cursor.fetchone()
        return row[0] if row and row[0] is not None else False

    def flush_statistics(self):
        """Flush aggregated statistics to database"""
        if not self.stats_buffer:
            return

        for composite_key, stats in self.stats_buffer.items():
            stat_type, stat_key, time_range = composite_key

            # Calculate visit count (unique dates)
            visit_count = len(stats['visit_dates'])

            # UPSERT footprint statistics
            self.conn.execute(
                '''
                INSERT INTO footprint_statistics (
                    stat_type, stat_key, time_range,
                    point_count, visit_count,
                    total_distance_m, total_duration_s,
                    first_visit, last_visit,
                    created_at, updated_at
                )
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
                ON CONFLICT(stat_type, stat_key, time_range) DO UPDATE SET
                    point_count = point_count + excluded.point_count,
                    visit_count = excluded.visit_count,
                    total_distance_m = total_distance_m + excluded.total_distance_m,
                    total_duration_s = total_duration_s + excluded.total_duration_s,
                    first_visit = MIN(first_visit, excluded.first_visit),
                    last_visit = MAX(last_visit, excluded.last_visit),
                    updated_at = CURRENT_TIMESTAMP
                ''',
                (
                    stat_type, stat_key, time_range,
                    stats['point_count'], visit_count,
                    stats['total_distance_m'], stats['total_duration_s'],
                    stats['first_visit'], stats['last_visit']
                )
            )

        # Commit and clear buffer
        self.conn.commit()
        self.stats_buffer.clear()

    def clear_previous_results(self):
        """Clear previous footprint statistics results"""
        self.logger.info("Clearing previous footprint statistics...")

        # Delete all footprint statistics
        self.conn.execute('DELETE FROM footprint_statistics')

        self.conn.commit()
        self.logger.info("Previous results cleared")


def main():
    """Main entry point"""
    parser = argparse.ArgumentParser(description='Footprint Statistics Worker')
    parser.add_argument('--task-id', type=int, required=True,
                       help='Analysis task ID')
    parser.add_argument('--db-path', type=str,
                       default='/data/tracks/tracks.db',
                       help='Path to SQLite database')
    parser.add_argument('--batch-size', type=int, default=1000,
                       help='Batch size for processing')

    args = parser.parse_args()

    # Create and run worker
    worker = FootprintStatisticsWorker(
        db_path=args.db_path,
        task_id=args.task_id,
        batch_size=args.batch_size
    )

    worker.run()


if __name__ == '__main__':
    main()
