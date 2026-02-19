"""
Stay Detection Worker

Detects stay segments using two criteria:
1. Spatial Stay: Points within radius (default 100m) for duration (default 2 hours)
2. Administrative Stay: Points in same admin area for duration (default 2 hours)

Algorithm:
1. Group consecutive points by spatial proximity or admin area
2. Calculate stay duration
3. Filter stays by minimum duration threshold
4. Classify stay type (HOME/WORK/TRANSIT/VISIT)
5. Calculate stay center (weighted average)
6. Update database with results
"""

import sys
import argparse
import json
import math
from typing import List, Tuple, Dict, Any, Optional
from datetime import datetime
from collections import defaultdict

# Add parent directory to path for imports
sys.path.append('/app/scripts/common')
from incremental_analyzer import IncrementalAnalyzer


class StayDetectionWorker(IncrementalAnalyzer):
    """Worker for stay detection"""

    def __init__(self, db_path: str, task_id: int, batch_size: int = 1000):
        super().__init__(db_path, task_id, batch_size)

        # Default parameters (can be overridden by task params)
        self.spatial_radius_m = 100
        self.min_duration_s = 7200  # 2 hours
        self.admin_level = 'county'  # county level for admin stays

    def calculate_distance(self, lat1: float, lon1: float,
                          lat2: float, lon2: float) -> float:
        """
        Calculate distance between two points in meters using Haversine formula

        Args:
            lat1, lon1: First point coordinates
            lat2, lon2: Second point coordinates

        Returns:
            Distance in meters
        """
        R = 6371000  # Earth radius in meters
        dlat = math.radians(lat2 - lat1)
        dlon = math.radians(lon2 - lon1)
        a = (math.sin(dlat/2) ** 2 +
             math.cos(math.radians(lat1)) * math.cos(math.radians(lat2)) *
             math.sin(dlon/2) ** 2)
        c = 2 * math.atan2(math.sqrt(a), math.sqrt(1-a))
        return R * c

    def detect_spatial_stays(self, points: List[Tuple]) -> List[Dict[str, Any]]:
        """
        Detect stays based on spatial proximity

        Args:
            points: List of point tuples

        Returns:
            List of stay segment dictionaries
        """
        stays = []
        current_stay_points = []

        for point in points:
            if not current_stay_points:
                # Start new stay
                current_stay_points.append(point)
                continue

            # Calculate distance to stay center
            center_lat = sum(p[3] for p in current_stay_points) / len(current_stay_points)
            center_lon = sum(p[2] for p in current_stay_points) / len(current_stay_points)
            distance = self.calculate_distance(point[3], point[2], center_lat, center_lon)

            if distance <= self.spatial_radius_m:
                # Within radius - add to current stay
                current_stay_points.append(point)
            else:
                # Outside radius - check if current stay meets duration threshold
                if len(current_stay_points) >= 2:
                    duration = current_stay_points[-1][1] - current_stay_points[0][1]
                    if duration >= self.min_duration_s:
                        stay = self.create_stay_segment(
                            current_stay_points,
                            'SPATIAL',
                            center_lat,
                            center_lon
                        )
                        stays.append(stay)

                # Start new stay
                current_stay_points = [point]

        # Check last stay
        if len(current_stay_points) >= 2:
            duration = current_stay_points[-1][1] - current_stay_points[0][1]
            if duration >= self.min_duration_s:
                center_lat = sum(p[3] for p in current_stay_points) / len(current_stay_points)
                center_lon = sum(p[2] for p in current_stay_points) / len(current_stay_points)
                stay = self.create_stay_segment(
                    current_stay_points,
                    'SPATIAL',
                    center_lat,
                    center_lon
                )
                stays.append(stay)

        return stays

    def detect_admin_stays(self, points: List[Tuple]) -> List[Dict[str, Any]]:
        """
        Detect stays based on administrative area

        Args:
            points: List of point tuples

        Returns:
            List of stay segment dictionaries
        """
        stays = []
        current_stay_points = []
        current_admin = None

        # Determine admin field index based on level
        admin_idx_map = {
            'province': 9,
            'city': 10,
            'county': 11,
            'town': 12,
            'village': 13
        }
        admin_idx = admin_idx_map.get(self.admin_level, 11)

        for point in points:
            admin_value = point[admin_idx] if len(point) > admin_idx else None

            if not current_stay_points:
                # Start new stay
                current_stay_points.append(point)
                current_admin = admin_value
                continue

            if admin_value == current_admin:
                # Same admin area - add to current stay
                current_stay_points.append(point)
            else:
                # Different admin area - check if current stay meets duration threshold
                if len(current_stay_points) >= 2:
                    duration = current_stay_points[-1][1] - current_stay_points[0][1]
                    if duration >= self.min_duration_s:
                        center_lat = sum(p[3] for p in current_stay_points) / len(current_stay_points)
                        center_lon = sum(p[2] for p in current_stay_points) / len(current_stay_points)
                        stay = self.create_stay_segment(
                            current_stay_points,
                            f'ADMIN_{self.admin_level.upper()}',
                            center_lat,
                            center_lon
                        )
                        stays.append(stay)

                # Start new stay
                current_stay_points = [point]
                current_admin = admin_value

        # Check last stay
        if len(current_stay_points) >= 2:
            duration = current_stay_points[-1][1] - current_stay_points[0][1]
            if duration >= self.min_duration_s:
                center_lat = sum(p[3] for p in current_stay_points) / len(current_stay_points)
                center_lon = sum(p[2] for p in current_stay_points) / len(current_stay_points)
                stay = self.create_stay_segment(
                    current_stay_points,
                    f'ADMIN_{self.admin_level.upper()}',
                    center_lat,
                    center_lon
                )
                stays.append(stay)

        return stays

    def classify_stay_type(self, stay: Dict[str, Any]) -> str:
        """
        Classify stay type based on time and frequency

        Args:
            stay: Stay segment dictionary

        Returns:
            Stay type: HOME, WORK, TRANSIT, VISIT
        """
        start_time = stay['start_time']
        duration_s = stay['duration_s']

        # Convert Unix timestamp to hour of day
        start_hour = datetime.fromtimestamp(start_time).hour

        # HOME: Night time stay (22:00-06:00)
        if 22 <= start_hour or start_hour < 6:
            return 'HOME'

        # WORK: Weekday daytime stay (09:00-18:00), long duration
        if 9 <= start_hour < 18 and duration_s > 14400:  # > 4 hours
            return 'WORK'

        # TRANSIT: Short stay (< 1 hour)
        if duration_s < 3600:
            return 'TRANSIT'

        # Default: VISIT
        return 'VISIT'

    def create_stay_segment(self, points: List[Tuple], stay_type: str,
                           center_lat: float, center_lon: float) -> Dict[str, Any]:
        """
        Create a stay segment record

        Args:
            points: List of points in the stay
            stay_type: Type of stay (SPATIAL, ADMIN_*)
            center_lat: Center latitude
            center_lon: Center longitude

        Returns:
            Stay segment dictionary
        """
        if not points:
            return None

        first_point = points[0]
        last_point = points[-1]

        start_time = first_point[1]
        end_time = last_point[1]
        duration_s = end_time - start_time

        # Calculate radius (max distance from center)
        max_radius = 0
        for point in points:
            distance = self.calculate_distance(point[3], point[2], center_lat, center_lon)
            max_radius = max(max_radius, distance)

        # Extract admin info from first point
        province = first_point[9] if len(first_point) > 9 else None
        city = first_point[10] if len(first_point) > 10 else None
        county = first_point[11] if len(first_point) > 11 else None
        town = first_point[12] if len(first_point) > 12 else None
        village = first_point[13] if len(first_point) > 13 else None

        stay = {
            'stay_type': stay_type,
            'start_time': start_time,
            'end_time': end_time,
            'duration_s': duration_s,
            'center_lat': center_lat,
            'center_lon': center_lon,
            'radius_m': max_radius,
            'province': province,
            'city': city,
            'county': county,
            'town': town,
            'village': village,
            'point_count': len(points),
            'confidence': 0.85,
            'reason_codes': json.dumps([stay_type]),
            'metadata': json.dumps({})
        }

        # Classify stay type (HOME/WORK/TRANSIT/VISIT)
        activity_type = self.classify_stay_type(stay)
        metadata = json.loads(stay['metadata'])
        metadata['activity_type'] = activity_type
        stay['metadata'] = json.dumps(metadata)

        return stay

    def process_batch(self, points: List[Tuple]) -> int:
        """
        Process a batch of points for stay detection

        Args:
            points: List of point tuples

        Returns:
            Number of failed points
        """
        failed = 0

        try:
            # Load task parameters
            params = self.task_info.get('params', {})
            self.spatial_radius_m = params.get('spatial_radius_m', 100)
            self.min_duration_s = params.get('min_duration_s', 7200)
            self.admin_level = params.get('admin_level', 'county')
            stay_detection_mode = params.get('mode', 'spatial')  # 'spatial' or 'admin'

            # Detect stays based on mode
            if stay_detection_mode == 'spatial':
                stays = self.detect_spatial_stays(points)
            elif stay_detection_mode == 'admin':
                stays = self.detect_admin_stays(points)
            else:
                # Both modes
                spatial_stays = self.detect_spatial_stays(points)
                admin_stays = self.detect_admin_stays(points)
                stays = spatial_stays + admin_stays

            # Insert stays into database
            for stay in stays:
                cursor = self.conn.execute("""
                    INSERT INTO stay_segments (
                        stay_type, start_time, end_time, duration_s,
                        center_lat, center_lon, radius_m,
                        province, city, county, town, village,
                        point_count, confidence, reason_codes, metadata
                    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                """, (
                    stay['stay_type'], stay['start_time'], stay['end_time'],
                    stay['duration_s'], stay['center_lat'], stay['center_lon'],
                    stay['radius_m'], stay['province'], stay['city'],
                    stay['county'], stay['town'], stay['village'],
                    stay['point_count'], stay['confidence'],
                    stay['reason_codes'], stay['metadata']
                ))
                stay_id = cursor.lastrowid

                # Update points with stay_id
                self.conn.execute("""
                    UPDATE "一生足迹"
                    SET stay_id = ?,
                        is_stay_point = 1
                    WHERE dataTime >= ? AND dataTime <= ?
                """, (stay_id, stay['start_time'], stay['end_time']))

            self.conn.commit()

        except Exception as e:
            self.logger.error(f"Error processing batch: {e}")
            failed = len(points)

        return failed

    def clear_previous_results(self):
        """Clear previous stay detection results"""
        self.logger.info("Clearing previous stay detection results...")

        # Clear stay references in track points
        self.conn.execute("""
            UPDATE "一生足迹"
            SET stay_id = NULL,
                is_stay_point = 0
        """)

        # Delete all stay segments
        self.conn.execute("DELETE FROM stay_segments")

        self.conn.commit()
        self.logger.info("Previous results cleared")


def main():
    parser = argparse.ArgumentParser(description='Stay Detection Worker')
    parser.add_argument('--db-path', required=True, help='Path to SQLite database')
    parser.add_argument('--task-id', type=int, required=True, help='Analysis task ID')
    parser.add_argument('--batch-size', type=int, default=1000, help='Batch size')

    args = parser.parse_args()

    worker = StayDetectionWorker(args.db_path, args.task_id, args.batch_size)
    worker.run()


if __name__ == '__main__':
    main()
