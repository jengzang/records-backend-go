"""
Streak Detection Worker

Detects continuous movement streaks:
1. High-speed streaks: Continuous high-speed movement (>60 km/h, >1 hour, >50km)
2. Walking streaks: Continuous walking (>30 min, >2km)

Algorithm:
1. Query all segments
2. Identify high-speed segments (CAR/TRAIN)
3. Identify walking segments
4. Check duration and distance thresholds
5. Insert extreme events into database
"""

import sys
import argparse
import json
from typing import List, Dict, Any

# Add parent directory to path for imports
sys.path.append('/app/scripts/common')
from task_executor import TaskExecutor


class StreakDetectionWorker(TaskExecutor):
    """Worker for streak detection"""

    def __init__(self, db_path: str, task_id: int):
        super().__init__(db_path, task_id)

        # Default thresholds
        self.high_speed_min_speed = 60  # km/h
        self.high_speed_min_duration = 3600  # 1 hour
        self.high_speed_min_distance = 50000  # 50 km

        self.walking_min_duration = 1800  # 30 minutes
        self.walking_min_distance = 2000  # 2 km

    def get_all_segments(self) -> List[Dict[str, Any]]:
        """
        Get all segments ordered by time

        Returns:
            List of segment dictionaries
        """
        cursor = self.conn.execute("""
            SELECT id, mode, start_time, end_time, duration_s,
                   distance_m, avg_speed_kmh, max_speed_kmh
            FROM segments
            ORDER BY start_time
        """)

        segments = []
        for row in cursor.fetchall():
            segments.append({
                'id': row[0],
                'mode': row[1],
                'start_time': row[2],
                'end_time': row[3],
                'duration_s': row[4],
                'distance_m': row[5],
                'avg_speed_kmh': row[6],
                'max_speed_kmh': row[7]
            })

        return segments

    def detect_high_speed_streaks(self, segments: List[Dict]) -> List[Dict[str, Any]]:
        """
        Detect high-speed streaks

        Args:
            segments: List of segment dictionaries

        Returns:
            List of high-speed streak events
        """
        streaks = []

        for segment in segments:
            # Check if segment qualifies as high-speed streak
            if segment['mode'] in ['CAR', 'TRAIN'] and \
               segment['avg_speed_kmh'] > self.high_speed_min_speed and \
               segment['duration_s'] > self.high_speed_min_duration and \
               segment['distance_m'] > self.high_speed_min_distance:

                streaks.append({
                    'event_type': 'HIGH_SPEED_STREAK',
                    'event_time': segment['start_time'],
                    'value': segment['avg_speed_kmh'],
                    'unit': 'km/h',
                    'metadata': json.dumps({
                        'segment_id': segment['id'],
                        'mode': segment['mode'],
                        'duration_s': segment['duration_s'],
                        'distance_m': segment['distance_m'],
                        'max_speed_kmh': segment['max_speed_kmh']
                    })
                })

        return streaks

    def detect_walking_streaks(self, segments: List[Dict]) -> List[Dict[str, Any]]:
        """
        Detect walking streaks

        Args:
            segments: List of segment dictionaries

        Returns:
            List of walking streak events
        """
        streaks = []

        for segment in segments:
            # Check if segment qualifies as walking streak
            if segment['mode'] == 'WALK' and \
               segment['duration_s'] > self.walking_min_duration and \
               segment['distance_m'] > self.walking_min_distance:

                streaks.append({
                    'event_type': 'WALKING_STREAK',
                    'event_time': segment['start_time'],
                    'value': segment['distance_m'],
                    'unit': 'meters',
                    'metadata': json.dumps({
                        'segment_id': segment['id'],
                        'duration_s': segment['duration_s'],
                        'avg_speed_kmh': segment['avg_speed_kmh']
                    })
                })

        return streaks

    def run(self):
        """
        Main execution method for streak detection

        This method:
        1. Connects to database
        2. Gets all segments
        3. Detects high-speed and walking streaks
        4. Inserts extreme events
        5. Marks task as completed
        """
        try:
            # Connect to database
            self.connect()

            # Get task information
            self.task_info = self.get_task_info()
            params = self.task_info.get('params', {})

            # Override thresholds if provided
            self.high_speed_min_speed = params.get('high_speed_min_speed', 60)
            self.high_speed_min_duration = params.get('high_speed_min_duration', 3600)
            self.high_speed_min_distance = params.get('high_speed_min_distance', 50000)
            self.walking_min_duration = params.get('walking_min_duration', 1800)
            self.walking_min_distance = params.get('walking_min_distance', 2000)

            self.logger.info(f"Starting streak detection")

            # Mark task as running
            self.mark_running()

            # Get all segments
            segments = self.get_all_segments()
            self.logger.info(f"Found {len(segments)} segments")

            # Detect streaks
            high_speed_streaks = self.detect_high_speed_streaks(segments)
            walking_streaks = self.detect_walking_streaks(segments)

            all_streaks = high_speed_streaks + walking_streaks
            self.logger.info(
                f"Detected {len(high_speed_streaks)} high-speed streaks, "
                f"{len(walking_streaks)} walking streaks"
            )

            # Insert extreme events
            for streak in all_streaks:
                self.conn.execute("""
                    INSERT INTO extreme_events (
                        event_type, event_time, value, unit, metadata
                    ) VALUES (?, ?, ?, ?, ?)
                """, (
                    streak['event_type'],
                    streak['event_time'],
                    streak['value'],
                    streak['unit'],
                    streak['metadata']
                ))

            self.conn.commit()

            # Mark task as completed
            self.mark_completed({
                'high_speed_streaks': len(high_speed_streaks),
                'walking_streaks': len(walking_streaks),
                'total_streaks': len(all_streaks)
            })

        except Exception as e:
            self.logger.exception(f"Streak detection failed: {e}")
            self.mark_failed(str(e))
            raise

        finally:
            # Always disconnect
            self.disconnect()


def main():
    parser = argparse.ArgumentParser(description='Streak Detection Worker')
    parser.add_argument('--db-path', required=True, help='Path to SQLite database')
    parser.add_argument('--task-id', type=int, required=True, help='Analysis task ID')

    args = parser.parse_args()

    worker = StreakDetectionWorker(args.db_path, args.task_id)
    worker.run()


if __name__ == '__main__':
    main()
