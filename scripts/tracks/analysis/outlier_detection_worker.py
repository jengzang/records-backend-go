"""
Outlier Detection Worker

Identifies GPS anomalies in trajectory data:
- GPS drift (low accuracy)
- Jump points (unrealistic speed)
- Backtrack patterns (A→B→A)
- Static drift (same location, varying coordinates)

Outputs:
- outlier_flag: BOOLEAN
- outlier_reason_codes: TEXT (JSON array)
- qa_status: TEXT (PASS/WARNING/FAIL)
"""

import sys
import json
import math
import argparse
from typing import List, Tuple, Dict, Any
sys.path.append('../../common')
from incremental_analyzer import IncrementalAnalyzer


class OutlierDetectionWorker(IncrementalAnalyzer):
    """Worker for detecting GPS outliers in trajectory data"""

    # Thresholds (can be overridden via params)
    LOW_ACCURACY_THRESHOLD = 100  # meters
    MAX_SPEED_THRESHOLD = 300  # km/h (non-flight)
    BACKTRACK_DISTANCE_THRESHOLD = 50  # meters
    BACKTRACK_TIME_THRESHOLD = 300  # seconds (5 minutes)
    STATIC_DRIFT_DISTANCE = 20  # meters
    STATIC_DRIFT_MIN_POINTS = 3  # minimum points to detect drift

    def __init__(self, db_path: str, task_id: int):
        super().__init__(db_path, task_id)
        self.load_thresholds()

    def load_thresholds(self):
        """Load threshold parameters from task params"""
        task_info = self.get_task_info()
        params = task_info.get('params', {})

        self.LOW_ACCURACY_THRESHOLD = params.get('low_accuracy_threshold', 100)
        self.MAX_SPEED_THRESHOLD = params.get('max_speed_threshold', 300)
        self.BACKTRACK_DISTANCE_THRESHOLD = params.get('backtrack_distance', 50)
        self.BACKTRACK_TIME_THRESHOLD = params.get('backtrack_time', 300)
        self.STATIC_DRIFT_DISTANCE = params.get('static_drift_distance', 20)

        self.logger.info(f"Loaded thresholds: accuracy={self.LOW_ACCURACY_THRESHOLD}m, "
                        f"speed={self.MAX_SPEED_THRESHOLD}km/h")

    def get_unanalyzed_points_query(self) -> str:
        """Override to get points that haven't been analyzed for outliers"""
        return """
            SELECT id, dataTime, longitude, latitude, heading,
                   accuracy, speed, distance, altitude,
                   province, city, county, town, village
            FROM "一生足迹"
            WHERE outlier_flag IS NULL
            ORDER BY dataTime
            LIMIT ?
        """

    def calculate_distance(self, lat1: float, lon1: float,
                          lat2: float, lon2: float) -> float:
        """
        Calculate distance between two GPS points using Haversine formula

        Returns:
            Distance in meters
        """
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
        """
        Calculate speed in km/h

        Args:
            distance: Distance in meters
            time_diff: Time difference in seconds

        Returns:
            Speed in km/h
        """
        if time_diff == 0:
            return 0
        return (distance / time_diff) * 3.6  # m/s to km/h

    def detect_low_accuracy(self, point: Dict) -> bool:
        """Detect if point has low GPS accuracy"""
        accuracy = point.get('accuracy', 0)
        return accuracy > self.LOW_ACCURACY_THRESHOLD

    def detect_jump(self, prev_point: Dict, curr_point: Dict) -> bool:
        """Detect unrealistic speed jump between points"""
        if not prev_point:
            return False

        # Calculate distance and time difference
        distance = self.calculate_distance(
            prev_point['latitude'], prev_point['longitude'],
            curr_point['latitude'], curr_point['longitude']
        )
        time_diff = curr_point['dataTime'] - prev_point['dataTime']

        # Skip if time gap is too large (>1 hour)
        if time_diff > 3600:
            return False

        # Calculate speed
        speed = self.calculate_speed(distance, time_diff)

        # Check if speed exceeds threshold
        return speed > self.MAX_SPEED_THRESHOLD

    def detect_backtrack(self, points_window: List[Dict]) -> bool:
        """
        Detect A→B→A backtrack pattern

        Args:
            points_window: List of 3 consecutive points [A, B, A']

        Returns:
            True if backtrack detected
        """
        if len(points_window) < 3:
            return False

        point_a = points_window[0]
        point_b = points_window[1]
        point_a_prime = points_window[2]

        # Calculate distances
        dist_ab = self.calculate_distance(
            point_a['latitude'], point_a['longitude'],
            point_b['latitude'], point_b['longitude']
        )
        dist_ba = self.calculate_distance(
            point_b['latitude'], point_b['longitude'],
            point_a_prime['latitude'], point_a_prime['longitude']
        )
        dist_aa = self.calculate_distance(
            point_a['latitude'], point_a['longitude'],
            point_a_prime['latitude'], point_a_prime['longitude']
        )

        # Check if A' is close to A
        if dist_aa > self.BACKTRACK_DISTANCE_THRESHOLD:
            return False

        # Check if time is short
        time_diff = point_a_prime['dataTime'] - point_a['dataTime']
        if time_diff > self.BACKTRACK_TIME_THRESHOLD:
            return False

        # Check if B is not too far
        if dist_ab < 10 or dist_ba < 10:  # Too close, not a real backtrack
            return False

        return True

    def detect_static_drift(self, points_window: List[Dict]) -> bool:
        """
        Detect static drift (same location, varying coordinates)

        Args:
            points_window: List of consecutive points

        Returns:
            True if static drift detected
        """
        if len(points_window) < self.STATIC_DRIFT_MIN_POINTS:
            return False

        # Calculate centroid
        avg_lat = sum(p['latitude'] for p in points_window) / len(points_window)
        avg_lon = sum(p['longitude'] for p in points_window) / len(points_window)

        # Check if all points are within threshold of centroid
        all_close = all(
            self.calculate_distance(p['latitude'], p['longitude'], avg_lat, avg_lon)
            < self.STATIC_DRIFT_DISTANCE
            for p in points_window
        )

        # Check if coordinates are varying (not exactly the same)
        coords_vary = len(set((p['latitude'], p['longitude']) for p in points_window)) > 1

        return all_close and coords_vary

    def process_batch(self, points: List[Tuple]) -> int:
        """
        Process a batch of points for outlier detection

        Returns:
            Number of failed points
        """
        failed = 0

        # Convert tuples to dicts for easier access
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

        # Process each point
        for i, point in enumerate(point_dicts):
            try:
                reason_codes = []

                # 1. Check low accuracy
                if self.detect_low_accuracy(point):
                    reason_codes.append('LOW_ACCURACY')

                # 2. Check jump (need previous point)
                if i > 0:
                    if self.detect_jump(point_dicts[i-1], point):
                        reason_codes.append('JUMP')

                # 3. Check backtrack (need 3-point window)
                if i >= 2:
                    if self.detect_backtrack(point_dicts[i-2:i+1]):
                        reason_codes.append('BACKTRACK')

                # 4. Check static drift (need 5-point window)
                if i >= 4:
                    if self.detect_static_drift(point_dicts[i-4:i+1]):
                        reason_codes.append('STATIC_DRIFT')

                # Determine QA status
                outlier_flag = len(reason_codes) > 0
                if not outlier_flag:
                    qa_status = 'PASS'
                elif 'LOW_ACCURACY' in reason_codes and len(reason_codes) == 1:
                    qa_status = 'WARNING'
                else:
                    qa_status = 'FAIL'

                # Update database
                self.conn.execute("""
                    UPDATE "一生足迹"
                    SET outlier_flag = ?,
                        outlier_reason_codes = ?,
                        qa_status = ?,
                        updated_at = CURRENT_TIMESTAMP
                    WHERE id = ?
                """, (
                    outlier_flag,
                    json.dumps(reason_codes) if reason_codes else None,
                    qa_status,
                    point['id']
                ))

            except Exception as e:
                self.logger.error(f"Failed to process point {point['id']}: {e}")
                failed += 1

        self.conn.commit()
        return failed

    def clear_previous_results(self):
        """Clear previous outlier detection results for full recompute"""
        self.logger.info("Clearing previous outlier detection results...")
        self.conn.execute("""
            UPDATE "一生足迹"
            SET outlier_flag = NULL,
                outlier_reason_codes = NULL,
                qa_status = NULL
        """)
        self.conn.commit()
        self.logger.info("Previous results cleared")


def main():
    parser = argparse.ArgumentParser(description='Outlier Detection Worker')
    parser.add_argument('--db-path', required=True, help='Path to SQLite database')
    parser.add_argument('--task-id', type=int, required=True, help='Analysis task ID')
    args = parser.parse_args()

    worker = OutlierDetectionWorker(args.db_path, args.task_id)

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
