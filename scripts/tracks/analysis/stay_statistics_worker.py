"""
Stay Statistics Worker

Aggregates stay segment data by administrative regions to generate stay rankings.
Supports multiple statistical types, time ranges, and activity types.

Statistical Types:
- PROVINCE: Province-level statistics
- CITY: City-level statistics
- COUNTY: County-level statistics
- TOWN: Town-level statistics
- ACTIVITY_TYPE: Activity type statistics (HOME/WORK/TRANSIT/VISIT)

Time Ranges:
- all: All time data
- year: Aggregated by year (YYYY)
- month: Aggregated by month (YYYY-MM)
- day: Aggregated by day (YYYY-MM-DD)

Metrics:
- stay_count: Number of stay segments
- total_duration_s: Total stay duration (seconds)
- avg_duration_s: Average stay duration (seconds)
- max_duration_s: Maximum stay duration (seconds)
- visit_count: Number of unique visit dates
- first_visit: First visit timestamp
- last_visit: Last visit timestamp
"""

import sys
import argparse
import json
from typing import List, Tuple, Dict, Any
from datetime import datetime, date
from collections import defaultdict

# Add parent directory to path for imports
sys.path.append('/app/scripts/common')
from task_executor import TaskExecutor


class StayStatisticsWorker(TaskExecutor):
    """Worker for stay statistics aggregation"""

    STAT_TYPES = ['PROVINCE', 'CITY', 'COUNTY', 'TOWN', 'ACTIVITY_TYPE']
    TIME_RANGES = {
        'all': None,
        'year': '%Y',
        'month': '%Y-%m',
        'day': '%Y-%m-%d'
    }

    def __init__(self, db_path: str, task_id: int, batch_size: int = 1000):
        super().__init__(db_path, task_id)
        self.batch_size = batch_size
        self.stats_buffer = defaultdict(lambda: {
            'stay_count': 0,
            'total_duration_s': 0,
            'max_duration_s': 0,
            'visit_dates': set(),
            'first_visit': None,
            'last_visit': None
        })

    def count_visit_days(self, start_time: int, end_time: int) -> int:
        """
        Count number of days a stay spans

        Args:
            start_time: Start timestamp (Unix seconds)
            end_time: End timestamp (Unix seconds)

        Returns:
            Number of days (minimum 1)
        """
        start_date = date.fromtimestamp(start_time)
        end_date = date.fromtimestamp(end_time)
        return (end_date - start_date).days + 1

    def get_stay_segments(self, limit: int = None) -> List[Tuple]:
        """
        Get stay segments from database

        Args:
            limit: Maximum number of segments to fetch

        Returns:
            List of stay segment tuples
        """
        if limit is None:
            limit = self.batch_size

        # Get task info to determine mode
        task_info = self.get_task_info()

        # Query stay segments with high confidence
        query = '''
            SELECT id, start_time, end_time, duration_s,
                   center_lat, center_lon,
                   province, city, county, town,
                   stay_type, confidence, metadata
            FROM stay_segments
            WHERE confidence > 0.7
            ORDER BY start_time
            LIMIT ?
        '''

        cursor = self.conn.execute(query, (limit,))
        return cursor.fetchall()

    def get_stat_key(self, segment: Tuple, stat_type: str) -> str:
        """
        Get statistical key for a segment based on stat_type

        Args:
            segment: Stay segment tuple
            stat_type: Statistical type

        Returns:
            Statistical key
        """
        if stat_type == 'PROVINCE':
            return segment[6] if len(segment) > 6 else None  # province
        elif stat_type == 'CITY':
            return segment[7] if len(segment) > 7 else None  # city
        elif stat_type == 'COUNTY':
            return segment[8] if len(segment) > 8 else None  # county
        elif stat_type == 'TOWN':
            return segment[9] if len(segment) > 9 else None  # town
        elif stat_type == 'ACTIVITY_TYPE':
            # Extract activity_type from metadata
            metadata_json = segment[12] if len(segment) > 12 else None
            if metadata_json:
                try:
                    metadata = json.loads(metadata_json)
                    return metadata.get('activity_type', 'UNKNOWN')
                except:
                    return 'UNKNOWN'
            return 'UNKNOWN'
        return None

    def process_stays(self) -> Dict[str, Any]:
        """
        Process all stay segments and aggregate statistics

        Returns:
            Result summary dict
        """
        processed = 0
        failed = 0
        offset = 0

        while True:
            # Fetch next batch of stay segments
            segments = self.get_stay_segments(self.batch_size)
            if not segments:
                self.logger.info("No more stay segments to process")
                break

            for segment in segments:
                try:
                    # Extract segment data
                    segment_id = segment[0]
                    start_time = segment[1]
                    end_time = segment[2]
                    duration_s = segment[3]

                    # Calculate visit days
                    visit_days = self.count_visit_days(start_time, end_time)

                    # Generate visit dates
                    visit_dates = set()
                    for day_offset in range(visit_days):
                        visit_date = datetime.fromtimestamp(
                            start_time + day_offset * 86400
                        ).strftime('%Y-%m-%d')
                        visit_dates.add(visit_date)

                    # Aggregate for each stat type and time range
                    for stat_type in self.STAT_TYPES:
                        stat_key = self.get_stat_key(segment, stat_type)
                        if not stat_key:
                            continue

                        for time_range_type, time_format in self.TIME_RANGES.items():
                            # Calculate time key
                            if time_format:
                                time_key = datetime.fromtimestamp(start_time).strftime(time_format)
                            else:
                                time_key = 'all'

                            # Create composite key
                            composite_key = (stat_type, stat_key, time_range)

                            # Aggregate statistics
                            stats = self.stats_buffer[composite_key]
                            stats['stay_count'] += 1
                            stats['total_duration_s'] += duration_s
                            stats['max_duration_s'] = max(stats['max_duration_s'], duration_s)
                            stats['visit_dates'].update(visit_dates)

                            # Update time range
                            if stats['first_visit'] is None or start_time < stats['first_visit']:
                                stats['first_visit'] = start_time
                            if stats['last_visit'] is None or end_time > stats['last_visit']:
                                stats['last_visit'] = end_time

                    processed += 1

                except Exception as e:
                    self.logger.error(f"Failed to process segment {segment[0]}: {e}")
                    failed += 1

            # Flush statistics to database every batch
            self.flush_statistics()

            # Update progress
            self.update_progress(
                processed=processed,
                failed=failed,
                progress_percent=int((processed / (processed + len(segments))) * 100)
            )

            offset += len(segments)

        return {
            'processed': processed,
            'failed': failed,
            'success_rate': (processed - failed) / processed if processed > 0 else 0
        }

    def flush_statistics(self):
        """Flush aggregated statistics to database"""
        if not self.stats_buffer:
            return

        for composite_key, stats in self.stats_buffer.items():
            stat_type, stat_key, time_range = composite_key

            # Calculate metrics
            visit_count = len(stats['visit_dates'])
            avg_duration_s = (stats['total_duration_s'] / stats['stay_count']
                            if stats['stay_count'] > 0 else 0)

            # UPSERT stay statistics
            self.conn.execute(
                '''
                INSERT INTO stay_statistics (
                    stat_type, stat_key, time_range,
                    stay_count, total_duration_s, avg_duration_s, max_duration_s,
                    visit_count, first_visit, last_visit,
                    created_at, updated_at
                )
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
                ON CONFLICT(stat_type, stat_key, time_range) DO UPDATE SET
                    stay_count = stay_count + excluded.stay_count,
                    total_duration_s = total_duration_s + excluded.total_duration_s,
                    avg_duration_s = (total_duration_s + excluded.total_duration_s) /
                                    (stay_count + excluded.stay_count),
                    max_duration_s = MAX(max_duration_s, excluded.max_duration_s),
                    visit_count = excluded.visit_count,
                    first_visit = MIN(first_visit, excluded.first_visit),
                    last_visit = MAX(last_visit, excluded.last_visit),
                    updated_at = CURRENT_TIMESTAMP
                ''',
                (
                    stat_type, stat_key, time_range,
                    stats['stay_count'], stats['total_duration_s'],
                    int(avg_duration_s), stats['max_duration_s'],
                    visit_count, stats['first_visit'], stats['last_visit']
                )
            )

        # Commit and clear buffer
        self.conn.commit()
        self.stats_buffer.clear()

    def run(self):
        """Main execution loop"""
        try:
            # Connect to database
            self.connect()

            # Mark task as running
            self.mark_running()

            # Process all stay segments
            result = self.process_stays()

            # Mark task as completed
            self.mark_completed(result)

        except Exception as e:
            self.logger.exception(f"Task execution failed: {e}")
            self.mark_failed(str(e))
            raise

        finally:
            # Always disconnect
            self.disconnect()

    def clear_previous_results(self):
        """Clear previous stay statistics results"""
        self.logger.info("Clearing previous stay statistics...")

        # Delete all stay statistics
        self.conn.execute('DELETE FROM stay_statistics')

        self.conn.commit()
        self.logger.info("Previous results cleared")


def main():
    """Main entry point"""
    parser = argparse.ArgumentParser(description='Stay Statistics Worker')
    parser.add_argument('--task-id', type=int, required=True,
                       help='Analysis task ID')
    parser.add_argument('--db-path', type=str,
                       default='/data/tracks/tracks.db',
                       help='Path to SQLite database')
    parser.add_argument('--batch-size', type=int, default=1000,
                       help='Batch size for processing')

    args = parser.parse_args()

    # Create and run worker
    worker = StayStatisticsWorker(
        db_path=args.db_path,
        task_id=args.task_id,
        batch_size=args.batch_size
    )

    worker.run()


if __name__ == '__main__':
    main()
