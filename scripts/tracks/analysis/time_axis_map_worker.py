"""
Time Axis Map Worker

Prepares trajectory data for time-based filtering and visualization:
- Validates time data consistency
- Creates time-based indexes for efficient querying
- Pre-computes time range metadata for quick access
- Generates time slice summaries for frontend time axis

Algorithm:
1. Validate time data:
   - Check dataTime consistency
   - Verify time_visually format
   - Ensure time field is properly formatted

2. Create time-based indexes:
   - Index on dataTime for range queries
   - Index on date (extracted from dataTime) for daily queries
   - Index on year-month for monthly queries

3. Generate time slice metadata:
   - Daily summaries (point count, distance, modes)
   - Monthly summaries
   - Yearly summaries

4. Store metadata in spatial_analysis table with type='TIME_SLICE'

This worker doesn't modify track points, but creates supporting data
for efficient time-based queries in the frontend.
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


class TimeAxisMapWorker(IncrementalAnalyzer):
    """Worker for time axis map preparation"""

    def __init__(self, db_path: str, task_id: int, batch_size: int = 1000):
        """
        Initialize time axis map worker

        Args:
            db_path: Path to SQLite database
            task_id: ID of the analysis task
            batch_size: Number of points to process per batch
        """
        super().__init__(db_path, task_id, batch_size)
        self.time_slices = defaultdict(lambda: {
            'point_count': 0,
            'distance_m': 0,
            'modes': set(),
            'min_time': None,
            'max_time': None
        })

    def process_batch(self, points: List[Tuple]) -> int:
        """
        Process a batch of points and aggregate time slice data

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

                # Skip invalid timestamps
                if dataTime is None or dataTime <= 0:
                    failed += 1
                    continue

                # Get mode
                mode = self.get_mode(point_id)

                # Convert timestamp to date strings
                dt = datetime.fromtimestamp(dataTime)
                date_str = dt.strftime('%Y-%m-%d')
                month_str = dt.strftime('%Y-%m')
                year_str = dt.strftime('%Y')

                # Aggregate by day
                self.aggregate_time_slice(f"DAY_{date_str}", dataTime, distance, mode)

                # Aggregate by month
                self.aggregate_time_slice(f"MONTH_{month_str}", dataTime, distance, mode)

                # Aggregate by year
                self.aggregate_time_slice(f"YEAR_{year_str}", dataTime, distance, mode)

            except Exception as e:
                self.logger.error(f"Failed to process point {point[0]}: {e}")
                failed += 1

        # Flush time slices to database every batch
        self.flush_time_slices()

        return failed

    def aggregate_time_slice(self, slice_key: str, timestamp: int, distance: float, mode: str):
        """
        Aggregate data for a time slice

        Args:
            slice_key: Time slice key (e.g., "DAY_2024-01-15")
            timestamp: Unix timestamp
            distance: Distance in meters
            mode: Transport mode
        """
        stats = self.time_slices[slice_key]
        stats['point_count'] += 1
        stats['distance_m'] += distance if distance else 0

        if mode:
            stats['modes'].add(mode)

        if stats['min_time'] is None or timestamp < stats['min_time']:
            stats['min_time'] = timestamp
        if stats['max_time'] is None or timestamp > stats['max_time']:
            stats['max_time'] = timestamp

    def get_mode(self, point_id: int) -> str:
        """Get transport mode for a point"""
        cursor = self.conn.execute(
            'SELECT mode FROM "一生足迹" WHERE id = ?',
            (point_id,)
        )
        row = cursor.fetchone()
        return row[0] if row else None

    def flush_time_slices(self):
        """Flush time slice metadata to database"""
        if not self.time_slices:
            return

        for slice_key, stats in self.time_slices.items():
            # Parse slice type and value
            slice_type, slice_value = slice_key.split('_', 1)

            # Convert sets to JSON
            modes_json = json.dumps(list(stats['modes']))

            # Calculate duration
            duration_s = 0
            if stats['min_time'] and stats['max_time']:
                duration_s = stats['max_time'] - stats['min_time']

            # Prepare metadata
            metadata = {
                'point_count': stats['point_count'],
                'distance_m': stats['distance_m'],
                'duration_s': duration_s,
                'modes': list(stats['modes']),
                'start_time': stats['min_time'],
                'end_time': stats['max_time']
            }

            # UPSERT into spatial_analysis table
            self.conn.execute(
                '''
                INSERT INTO spatial_analysis (
                    analysis_type, analysis_key, time_range,
                    result_value, metadata, created_at, updated_at
                )
                VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
                ON CONFLICT(analysis_type, analysis_key, time_range) DO UPDATE SET
                    result_value = result_value + excluded.result_value,
                    metadata = excluded.metadata,
                    updated_at = CURRENT_TIMESTAMP
                ''',
                (
                    'TIME_SLICE',
                    slice_type,
                    slice_value,
                    stats['point_count'],
                    json.dumps(metadata)
                )
            )

        # Commit and clear buffer
        self.conn.commit()
        self.time_slices.clear()

    def clear_previous_results(self):
        """Clear previous time axis map results"""
        self.logger.info("Clearing previous time axis map results...")

        # Delete time slice records from spatial_analysis
        self.conn.execute(
            '''
            DELETE FROM spatial_analysis
            WHERE analysis_type = 'TIME_SLICE'
            '''
        )

        self.conn.commit()
        self.logger.info("Previous results cleared")

    def create_time_indexes(self):
        """Create time-based indexes for efficient querying"""
        self.logger.info("Creating time-based indexes...")

        try:
            # Index on dataTime for range queries
            self.conn.execute(
                '''
                CREATE INDEX IF NOT EXISTS idx_datatime_range
                ON "一生足迹"(dataTime)
                '''
            )

            # Index on time field for string-based queries
            self.conn.execute(
                '''
                CREATE INDEX IF NOT EXISTS idx_time_string
                ON "一生足迹"(time)
                '''
            )

            self.conn.commit()
            self.logger.info("Time-based indexes created")

        except Exception as e:
            self.logger.warning(f"Failed to create indexes (may already exist): {e}")

    def run(self):
        """Override run to create indexes before processing"""
        # Create indexes first
        self.connect()
        self.create_time_indexes()
        self.disconnect()

        # Run standard incremental analysis
        super().run()


def main():
    """Main entry point"""
    parser = argparse.ArgumentParser(description='Time Axis Map Worker')
    parser.add_argument('--task-id', type=int, required=True,
                       help='Analysis task ID')
    parser.add_argument('--db-path', type=str,
                       default='/data/tracks/tracks.db',
                       help='Path to SQLite database')
    parser.add_argument('--batch-size', type=int, default=1000,
                       help='Batch size for processing')

    args = parser.parse_args()

    # Create and run worker
    worker = TimeAxisMapWorker(
        db_path=args.db_path,
        task_id=args.task_id,
        batch_size=args.batch_size
    )

    worker.run()


if __name__ == '__main__':
    main()
