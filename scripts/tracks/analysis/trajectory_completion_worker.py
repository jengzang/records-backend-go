"""
Trajectory Completion Worker

Fills in missing trajectory segments for train and flight travel:
- Identifies train/flight segments (high speed, long duration)
- Matches to known routes (KML/JSON data)
- Interpolates intermediate points
- Marks synthetic points

Outputs:
- is_synthetic: BOOLEAN
- synthetic_source: TEXT (TRAIN_INTERPOLATION, FLIGHT_INTERPOLATION)
- synthetic_metadata: TEXT (JSON with route info)
"""

import sys
import json
import math
import argparse
import os
from typing import List, Tuple, Dict, Any, Optional
sys.path.append('../../common')
from incremental_analyzer import IncrementalAnalyzer


class TrajectoryCompletionWorker(IncrementalAnalyzer):
    """Worker for completing missing train/flight trajectory segments"""

    # Thresholds
    HIGH_SPEED_THRESHOLD = 80  # km/h
    MIN_SEGMENT_DURATION = 1800  # 30 minutes
    FLIGHT_ALTITUDE_THRESHOLD = 1000  # meters
    TRAIN_SPEED_RANGE = (80, 350)  # km/h
    FLIGHT_SPEED_RANGE = (200, 1000)  # km/h

    def __init__(self, db_path: str, task_id: int):
        super().__init__(db_path, task_id)
        self.train_routes = []
        self.flight_routes = []
        self.load_thresholds()
        self.load_route_data()

    def load_thresholds(self):
        """Load threshold parameters from task params"""
        task_info = self.get_task_info()
        params = task_info.get('params', {})

        self.HIGH_SPEED_THRESHOLD = params.get('high_speed_threshold', 80)
        self.MIN_SEGMENT_DURATION = params.get('min_segment_duration', 1800)
        self.FLIGHT_ALTITUDE_THRESHOLD = params.get('flight_altitude', 1000)

        self.logger.info(f"Loaded thresholds: speed={self.HIGH_SPEED_THRESHOLD}km/h, "
                        f"duration={self.MIN_SEGMENT_DURATION}s")

    def load_route_data(self):
        """Load train and flight route data"""
        # Note: In production, this would load actual KML/JSON files
        # For now, we'll use a simplified approach
        self.logger.info("Route data loading skipped (simplified implementation)")
        # TODO: Implement actual route loading from TrainPlane/ directory

    def get_unanalyzed_points_query(self) -> str:
        """Override to get points that haven't been analyzed for completion"""
        return """
            SELECT id, dataTime, longitude, latitude, heading,
                   accuracy, speed, distance, altitude,
                   province, city, county, town, village
            FROM "一生足迹"
            WHERE is_synthetic IS NULL
            ORDER BY dataTime
            LIMIT ?
        """

    def calculate_distance(self, lat1: float, lon1: float,
                          lat2: float, lon2: float) -> float:
        """Calculate distance between two GPS points using Haversine formula"""
        R = 6371000  # Earth radius in meters

        lat1_rad = math.radians(lat1)
        lat2_rad = math.radians(lat2)
        delta_lat = math.radians(lat2 - lat1)
        delta_lon = math.radians(lon2 - lon1)

        a = (math.sin(delta_lat / 2) ** 2 +
             math.cos(lat1_rad) * math.cos(lat2_rad) *
             math.sin(delta_lon / 2) ** 2)
        c = 2 * math.atan2(math.sqrt(a), math.sqrt(1 - a))

        return R * c

    def calculate_speed(self, distance: float, time_diff: int) -> float:
        """Calculate speed in km/h"""
        if time_diff == 0:
            return 0
        return (distance / time_diff) * 3.6  # m/s to km/h

    def identify_transport_type(self, points_window: List[Dict]) -> Optional[str]:
        """
        Identify if a segment is train or flight based on characteristics

        Returns:
            'TRAIN', 'FLIGHT', or None
        """
        if len(points_window) < 2:
            return None

        # Calculate average speed and altitude
        speeds = []
        altitudes = []

        for i in range(1, len(points_window)):
            prev = points_window[i-1]
            curr = points_window[i]

            distance = self.calculate_distance(
                prev['latitude'], prev['longitude'],
                curr['latitude'], curr['longitude']
            )
            time_diff = curr['dataTime'] - prev['dataTime']
            speed = self.calculate_speed(distance, time_diff)

            speeds.append(speed)
            altitudes.append(curr['altitude'] if curr['altitude'] else 0)

        avg_speed = sum(speeds) / len(speeds) if speeds else 0
        avg_altitude = sum(altitudes) / len(altitudes) if altitudes else 0

        # Check for flight
        if (avg_altitude > self.FLIGHT_ALTITUDE_THRESHOLD and
            self.FLIGHT_SPEED_RANGE[0] <= avg_speed <= self.FLIGHT_SPEED_RANGE[1]):
            return 'FLIGHT'

        # Check for train
        if self.TRAIN_SPEED_RANGE[0] <= avg_speed <= self.TRAIN_SPEED_RANGE[1]:
            # Additional check: cross-province travel
            provinces = set(p['province'] for p in points_window if p['province'])
            if len(provinces) > 1:
                return 'TRAIN'

        return None

    def detect_gaps(self, points: List[Dict]) -> List[Tuple[int, int]]:
        """
        Detect time gaps that might need interpolation

        Returns:
            List of (start_idx, end_idx) tuples indicating gaps
        """
        gaps = []
        GAP_THRESHOLD = 3600  # 1 hour

        for i in range(1, len(points)):
            time_diff = points[i]['dataTime'] - points[i-1]['dataTime']
            if time_diff > GAP_THRESHOLD:
                # Check if this gap is during high-speed travel
                distance = self.calculate_distance(
                    points[i-1]['latitude'], points[i-1]['longitude'],
                    points[i]['latitude'], points[i]['longitude']
                )
                speed = self.calculate_speed(distance, time_diff)

                if speed > self.HIGH_SPEED_THRESHOLD:
                    gaps.append((i-1, i))

        return gaps

    def interpolate_points(self, start_point: Dict, end_point: Dict,
                          num_points: int, transport_type: str) -> List[Dict]:
        """
        Interpolate intermediate points between start and end

        Args:
            start_point: Starting point
            end_point: Ending point
            num_points: Number of points to interpolate
            transport_type: 'TRAIN' or 'FLIGHT'

        Returns:
            List of interpolated point dictionaries
        """
        interpolated = []

        for i in range(1, num_points + 1):
            ratio = i / (num_points + 1)

            # Linear interpolation for lat/lon
            lat = start_point['latitude'] + ratio * (end_point['latitude'] - start_point['latitude'])
            lon = start_point['longitude'] + ratio * (end_point['longitude'] - start_point['longitude'])

            # Linear interpolation for time
            time = int(start_point['dataTime'] + ratio * (end_point['dataTime'] - start_point['dataTime']))

            # Estimate altitude
            if transport_type == 'FLIGHT':
                # Parabolic altitude profile (climb, cruise, descend)
                if ratio < 0.2:  # Climb
                    alt = 0 + (10000 * (ratio / 0.2))
                elif ratio > 0.8:  # Descend
                    alt = 10000 * ((1 - ratio) / 0.2)
                else:  # Cruise
                    alt = 10000
            else:  # TRAIN
                alt = 0

            interpolated.append({
                'dataTime': time,
                'longitude': lon,
                'latitude': lat,
                'altitude': alt,
                'is_synthetic': True,
                'synthetic_source': f'{transport_type}_INTERPOLATION',
                'synthetic_metadata': json.dumps({
                    'start_id': start_point['id'],
                    'end_id': end_point['id'],
                    'interpolation_ratio': ratio
                })
            })

        return interpolated

    def process_batch(self, points: List[Tuple]) -> int:
        """
        Process a batch of points for trajectory completion

        Returns:
            Number of failed points
        """
        failed = 0

        # Convert tuples to dicts
        point_dicts = []
        for p in points:
            point_dicts.append({
                'id': p[0],
                'dataTime': p[1],
                'longitude': p[2],
                'latitude': p[3],
                'heading': p[4],
                'accuracy': p[5],
                'speed': p[6],
                'distance': p[7],
                'altitude': p[8],
                'province': p[9],
                'city': p[10],
                'county': p[11],
                'town': p[12],
                'village': p[13]
            })

        # Detect gaps
        gaps = self.detect_gaps(point_dicts)

        if not gaps:
            # No gaps, mark all points as non-synthetic
            for point in point_dicts:
                try:
                    self.conn.execute("""
                        UPDATE "一生足迹"
                        SET is_synthetic = 0,
                            updated_at = CURRENT_TIMESTAMP
                        WHERE id = ?
                    """, (point['id'],))
                except Exception as e:
                    self.logger.error(f"Failed to update point {point['id']}: {e}")
                    failed += 1

            self.conn.commit()
            return failed

        # Process each gap
        for start_idx, end_idx in gaps:
            start_point = point_dicts[start_idx]
            end_point = point_dicts[end_idx]

            # Identify transport type
            window = point_dicts[max(0, start_idx-2):min(len(point_dicts), end_idx+3)]
            transport_type = self.identify_transport_type(window)

            if not transport_type:
                # Can't determine type, skip interpolation
                continue

            # Calculate number of points to interpolate
            time_diff = end_point['dataTime'] - start_point['dataTime']
            num_points = max(1, time_diff // 600)  # One point every 10 minutes

            # Interpolate points
            try:
                interpolated = self.interpolate_points(
                    start_point, end_point, num_points, transport_type
                )

                # Insert interpolated points
                for interp_point in interpolated:
                    self.conn.execute("""
                        INSERT INTO "一生足迹"
                        (dataTime, longitude, latitude, altitude,
                         is_synthetic, synthetic_source, synthetic_metadata,
                         created_at, updated_at)
                        VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
                    """, (
                        interp_point['dataTime'],
                        interp_point['longitude'],
                        interp_point['latitude'],
                        interp_point['altitude'],
                        1,
                        interp_point['synthetic_source'],
                        interp_point['synthetic_metadata']
                    ))

                self.logger.info(f"Interpolated {num_points} {transport_type} points "
                               f"between {start_point['id']} and {end_point['id']}")

            except Exception as e:
                self.logger.error(f"Failed to interpolate gap: {e}")
                failed += 1

        # Mark original points as non-synthetic
        for point in point_dicts:
            try:
                self.conn.execute("""
                    UPDATE "一生足迹"
                    SET is_synthetic = 0,
                        updated_at = CURRENT_TIMESTAMP
                    WHERE id = ?
                """, (point['id'],))
            except Exception as e:
                self.logger.error(f"Failed to update point {point['id']}: {e}")
                failed += 1

        self.conn.commit()
        return failed

    def clear_previous_results(self):
        """Clear previous trajectory completion results for full recompute"""
        self.logger.info("Clearing previous trajectory completion results...")
        self.conn.execute("""
            DELETE FROM "一生足迹"
            WHERE is_synthetic = 1
        """)
        self.conn.execute("""
            UPDATE "一生足迹"
            SET is_synthetic = NULL,
                synthetic_source = NULL,
                synthetic_metadata = NULL
        """)
        self.conn.commit()
        self.logger.info("Previous results cleared")


def main():
    parser = argparse.ArgumentParser(description='Trajectory Completion Worker')
    parser.add_argument('--db-path', required=True, help='Path to SQLite database')
    parser.add_argument('--task-id', type=int, required=True, help='Analysis task ID')
    args = parser.parse_args()

    worker = TrajectoryCompletionWorker(args.db_path, args.task_id)

    try:
        worker.connect()
        worker.run()
    except Exception as e:
        worker.logger.error(f"Worker failed: {e}", exc_info=True)
        worker.mark_failed(str(e))
        sys.exit(1)
    finally:
        worker.disconnect()


if __name__ == '__main__':
    main()
