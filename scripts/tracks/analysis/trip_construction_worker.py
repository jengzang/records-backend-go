"""
Trip Construction Worker

Constructs trips based on stay segments.
A trip is defined as movement between two stay segments.

Algorithm:
1. Query all stay segments ordered by time
2. For each pair of consecutive stays, create a trip
3. Calculate trip statistics (distance, duration, modes used)
4. Classify trip type (COMMUTE, ROUND_TRIP, ONE_WAY, MULTI_STOP)
5. Insert trip records into database
"""

import sys
import argparse
import json
from typing import List, Tuple, Dict, Any
from datetime import datetime

# Add parent directory to path for imports
sys.path.append('/app/scripts/common')
from task_executor import TaskExecutor


class TripConstructionWorker(TaskExecutor):
    """Worker for trip construction"""

    def __init__(self, db_path: str, task_id: int):
        super().__init__(db_path, task_id)

    def get_all_stays(self) -> List[Dict[str, Any]]:
        """
        Get all stay segments ordered by time

        Returns:
            List of stay dictionaries
        """
        cursor = self.conn.execute("""
            SELECT id, stay_type, start_time, end_time, duration_s,
                   center_lat, center_lon, province, city, county, metadata
            FROM stay_segments
            ORDER BY start_time
        """)

        stays = []
        for row in cursor.fetchall():
            stays.append({
                'id': row[0],
                'stay_type': row[1],
                'start_time': row[2],
                'end_time': row[3],
                'duration_s': row[4],
                'center_lat': row[5],
                'center_lon': row[6],
                'province': row[7],
                'city': row[8],
                'county': row[9],
                'metadata': json.loads(row[10]) if row[10] else {}
            })

        return stays

    def get_segments_between_stays(self, start_time: int, end_time: int) -> List[Dict[str, Any]]:
        """
        Get all segments between two stay times

        Args:
            start_time: Start timestamp
            end_time: End timestamp

        Returns:
            List of segment dictionaries
        """
        cursor = self.conn.execute("""
            SELECT id, mode, distance_m, duration_s
            FROM segments
            WHERE start_time >= ? AND end_time <= ?
            ORDER BY start_time
        """, (start_time, end_time))

        segments = []
        for row in cursor.fetchall():
            segments.append({
                'id': row[0],
                'mode': row[1],
                'distance_m': row[2],
                'duration_s': row[3]
            })

        return segments

    def classify_trip_type(self, origin_stay: Dict, dest_stay: Dict,
                          all_stays: List[Dict]) -> str:
        """
        Classify trip type

        Args:
            origin_stay: Origin stay dictionary
            dest_stay: Destination stay dictionary
            all_stays: All stays for context

        Returns:
            Trip type: COMMUTE, ROUND_TRIP, ONE_WAY, MULTI_STOP
        """
        # Check if it's a commute (HOME <-> WORK)
        origin_activity = origin_stay['metadata'].get('activity_type', '')
        dest_activity = dest_stay['metadata'].get('activity_type', '')

        if (origin_activity == 'HOME' and dest_activity == 'WORK') or \
           (origin_activity == 'WORK' and dest_activity == 'HOME'):
            return 'COMMUTE'

        # Check if it's a round trip (same location)
        if origin_stay['county'] == dest_stay['county']:
            return 'ROUND_TRIP'

        # Check if there are multiple stops
        # (This is simplified - could be enhanced to check intermediate stays)
        return 'ONE_WAY'

    def run(self):
        """
        Main execution method for trip construction

        This method:
        1. Connects to database
        2. Gets all stay segments
        3. Constructs trips between consecutive stays
        4. Inserts trip records
        5. Marks task as completed
        """
        try:
            # Connect to database
            self.connect()

            # Get task information
            self.task_info = self.get_task_info()
            self.logger.info(f"Starting trip construction")

            # Mark task as running
            self.mark_running()

            # Get all stays
            stays = self.get_all_stays()
            self.logger.info(f"Found {len(stays)} stay segments")

            if len(stays) < 2:
                self.logger.warning("Not enough stays to construct trips")
                self.mark_completed({
                    'trips_created': 0,
                    'message': 'Not enough stays'
                })
                return

            # Construct trips between consecutive stays
            trips_created = 0
            trips_by_date = {}

            for i in range(len(stays) - 1):
                origin_stay = stays[i]
                dest_stay = stays[i + 1]

                # Get date for trip numbering
                trip_date = datetime.fromtimestamp(origin_stay['end_time']).strftime('%Y-%m-%d')

                # Get trip number for this date
                if trip_date not in trips_by_date:
                    trips_by_date[trip_date] = 0
                trips_by_date[trip_date] += 1
                trip_number = trips_by_date[trip_date]

                # Get segments between stays
                segments = self.get_segments_between_stays(
                    origin_stay['end_time'],
                    dest_stay['start_time']
                )

                # Calculate trip statistics
                total_distance = sum(s['distance_m'] for s in segments)
                total_duration = dest_stay['start_time'] - origin_stay['end_time']
                modes_used = list(set(s['mode'] for s in segments))

                # Classify trip type
                trip_type = self.classify_trip_type(origin_stay, dest_stay, stays)

                # Create trip record
                self.conn.execute("""
                    INSERT INTO trips (
                        date, trip_number, origin_stay_id, dest_stay_id,
                        start_time, end_time, duration_s, distance_m,
                        segment_count, modes, metadata
                    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                """, (
                    trip_date,
                    trip_number,
                    origin_stay['id'],
                    dest_stay['id'],
                    origin_stay['end_time'],
                    dest_stay['start_time'],
                    total_duration,
                    total_distance,
                    len(segments),
                    json.dumps(modes_used),
                    json.dumps({'trip_type': trip_type})
                ))

                trips_created += 1

                # Update progress every 100 trips
                if trips_created % 100 == 0:
                    progress_percent = int((i / (len(stays) - 1)) * 100)
                    self.update_progress(
                        processed=i,
                        progress_percent=progress_percent
                    )

            self.conn.commit()

            # Mark task as completed
            self.mark_completed({
                'trips_created': trips_created,
                'stays_processed': len(stays),
                'unique_dates': len(trips_by_date)
            })

        except Exception as e:
            self.logger.exception(f"Trip construction failed: {e}")
            self.mark_failed(str(e))
            raise

        finally:
            # Always disconnect
            self.disconnect()


def main():
    parser = argparse.ArgumentParser(description='Trip Construction Worker')
    parser.add_argument('--db-path', required=True, help='Path to SQLite database')
    parser.add_argument('--task-id', type=int, required=True, help='Analysis task ID')

    args = parser.parse_args()

    worker = TripConstructionWorker(args.db_path, args.task_id)
    worker.run()


if __name__ == '__main__':
    main()
