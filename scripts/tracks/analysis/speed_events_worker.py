"""
Speed Events Worker

Detects extreme speed events and generates speed rankings:
1. Maximum speed events (by transport mode)
2. Longest distance segments
3. Longest duration segments

Algorithm:
1. Query all segments
2. Find top N segments by speed (per mode)
3. Find top N segments by distance
4. Find top N segments by duration
5. Insert extreme events into database
"""

import sys
import argparse
import json
from typing import List, Dict, Any

# Add parent directory to path for imports
sys.path.append('/app/scripts/common')
from task_executor import TaskExecutor


class SpeedEventsWorker(TaskExecutor):
    """Worker for speed events detection"""

    def __init__(self, db_path: str, task_id: int):
        super().__init__(db_path, task_id)

        # Default parameters
        self.top_n = 10  # Top N events to record

        # Speed thresholds by mode (km/h)
        self.speed_thresholds = {
            'WALK': 8,
            'CAR': 120,
            'TRAIN': 300,
            'FLIGHT': 800
        }

    def get_all_segments(self) -> List[Dict[str, Any]]:
        """
        Get all segments ordered by time

        Returns:
            List of segment dictionaries
        """
        cursor = self.conn.execute("""
            SELECT id, mode, start_time, end_time, duration_s,
                   distance_m, avg_speed_kmh, max_speed_kmh,
                   start_point_id, end_point_id
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
                'max_speed_kmh': row[7],
                'start_point_id': row[8],
                'end_point_id': row[9]
            })

        return segments

    def detect_max_speed_events(self, segments: List[Dict]) -> List[Dict[str, Any]]:
        """
        Detect maximum speed events (top N per mode)

        Args:
            segments: List of segment dictionaries

        Returns:
            List of max speed events
        """
        events = []

        # Group segments by mode
        segments_by_mode = {}
        for segment in segments:
            mode = segment['mode']
            if mode not in segments_by_mode:
                segments_by_mode[mode] = []
            segments_by_mode[mode].append(segment)

        # Find top N by speed for each mode
        for mode, mode_segments in segments_by_mode.items():
            # Sort by max_speed_kmh descending
            sorted_segments = sorted(
                mode_segments,
                key=lambda s: s['max_speed_kmh'],
                reverse=True
            )

            # Take top N
            top_segments = sorted_segments[:self.top_n]

            # Check if speed exceeds threshold
            threshold = self.speed_thresholds.get(mode, 0)

            for rank, segment in enumerate(top_segments, 1):
                if segment['max_speed_kmh'] > threshold:
                    events.append({
                        'event_type': 'MAX_SPEED',
                        'event_time': segment['start_time'],
                        'value': segment['max_speed_kmh'],
                        'unit': 'km/h',
                        'metadata': json.dumps({
                            'segment_id': segment['id'],
                            'mode': mode,
                            'rank': rank,
                            'threshold': threshold,
                            'distance_m': segment['distance_m'],
                            'duration_s': segment['duration_s']
                        })
                    })

        return events

    def detect_long_distance_events(self, segments: List[Dict]) -> List[Dict[str, Any]]:
        """
        Detect longest distance segments (top N overall)

        Args:
            segments: List of segment dictionaries

        Returns:
            List of long distance events
        """
        events = []

        # Sort by distance descending
        sorted_segments = sorted(
            segments,
            key=lambda s: s['distance_m'],
            reverse=True
        )

        # Take top N
        top_segments = sorted_segments[:self.top_n]

        for rank, segment in enumerate(top_segments, 1):
            events.append({
                'event_type': 'LONG_DISTANCE',
                'event_time': segment['start_time'],
                'value': segment['distance_m'],
                'unit': 'meters',
                'metadata': json.dumps({
                    'segment_id': segment['id'],
                    'mode': segment['mode'],
                    'rank': rank,
                    'duration_s': segment['duration_s'],
                    'avg_speed_kmh': segment['avg_speed_kmh']
                })
            })

        return events

    def detect_long_duration_events(self, segments: List[Dict]) -> List[Dict[str, Any]]:
        """
        Detect longest duration segments (top N overall)

        Args:
            segments: List of segment dictionaries

        Returns:
            List of long duration events
        """
        events = []

        # Sort by duration descending
        sorted_segments = sorted(
            segments,
            key=lambda s: s['duration_s'],
            reverse=True
        )

        # Take top N
        top_segments = sorted_segments[:self.top_n]

        for rank, segment in enumerate(top_segments, 1):
            events.append({
                'event_type': 'LONG_DURATION',
                'event_time': segment['start_time'],
                'value': segment['duration_s'],
                'unit': 'seconds',
                'metadata': json.dumps({
                    'segment_id': segment['id'],
                    'mode': segment['mode'],
                    'rank': rank,
                    'distance_m': segment['distance_m'],
                    'avg_speed_kmh': segment['avg_speed_kmh']
                })
            })

        return events

    def run(self):
        """
        Main execution method for speed events detection

        This method:
        1. Connects to database
        2. Gets all segments
        3. Detects max speed, long distance, and long duration events
        4. Inserts extreme events
        5. Marks task as completed
        """
        try:
            # Connect to database
            self.connect()

            # Get task information
            self.task_info = self.get_task_info()
            params = self.task_info.get('params', {})

            # Override parameters if provided
            self.top_n = params.get('top_n', 10)

            self.logger.info(f"Starting speed events detection (top {self.top_n})")

            # Mark task as running
            self.mark_running()

            # Get all segments
            segments = self.get_all_segments()
            self.logger.info(f"Found {len(segments)} segments")

            # Detect events
            max_speed_events = self.detect_max_speed_events(segments)
            long_distance_events = self.detect_long_distance_events(segments)
            long_duration_events = self.detect_long_duration_events(segments)

            all_events = max_speed_events + long_distance_events + long_duration_events
            self.logger.info(
                f"Detected {len(max_speed_events)} max speed events, "
                f"{len(long_distance_events)} long distance events, "
                f"{len(long_duration_events)} long duration events"
            )

            # Insert extreme events
            for event in all_events:
                self.conn.execute("""
                    INSERT INTO extreme_events (
                        event_type, event_time, value, unit, metadata
                    ) VALUES (?, ?, ?, ?, ?)
                """, (
                    event['event_type'],
                    event['event_time'],
                    event['value'],
                    event['unit'],
                    event['metadata']
                ))

            self.conn.commit()

            # Mark task as completed
            self.mark_completed({
                'max_speed_events': len(max_speed_events),
                'long_distance_events': len(long_distance_events),
                'long_duration_events': len(long_duration_events),
                'total_events': len(all_events)
            })

        except Exception as e:
            self.logger.exception(f"Speed events detection failed: {e}")
            self.mark_failed(str(e))
            raise

        finally:
            # Always disconnect
            self.disconnect()


def main():
    parser = argparse.ArgumentParser(description='Speed Events Worker')
    parser.add_argument('--db-path', required=True, help='Path to SQLite database')
    parser.add_argument('--task-id', type=int, required=True, help='Analysis task ID')

    args = parser.parse_args()

    worker = SpeedEventsWorker(args.db_path, args.task_id)
    worker.run()


if __name__ == '__main__':
    main()
