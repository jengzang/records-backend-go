"""
Transport Mode Classification Worker

Classifies GPS trajectory points into transport modes:
- WALK: Walking (speed < 10 km/h)
- CAR: Driving (speed 20-120 km/h)
- TRAIN: Train travel (speed 80-350 km/h, crosses provinces)
- FLIGHT: Air travel (altitude > 1000m, speed 200-1000 km/h)
- STAY: Stationary (speed < 1 km/h, duration > 10 min)
- UNKNOWN: Cannot determine

Algorithm:
1. Extract features (speed, acceleration, altitude, admin changes)
2. Apply rule-based classification
3. Segment trajectory by mode changes
4. Calculate confidence scores
5. Update database with results
"""

import sys
import argparse
import json
import math
from typing import List, Tuple, Dict, Any, Optional
from datetime import datetime

# Add parent directory to path for imports
sys.path.append('/app/scripts/common')
from incremental_analyzer import IncrementalAnalyzer


class TransportModeWorker(IncrementalAnalyzer):
    """Worker for transport mode classification"""

    def __init__(self, db_path: str, task_id: int, batch_size: int = 1000):
        super().__init__(db_path, task_id, batch_size)
        self.current_segment = None
        self.segment_buffer = []

    def calculate_speed(self, p1: Tuple, p2: Tuple) -> float:
        """
        Calculate speed between two points in km/h

        Args:
            p1: First point (id, dataTime, lon, lat, ...)
            p2: Second point

        Returns:
            Speed in km/h
        """
        # Extract coordinates
        lon1, lat1 = p1[2], p1[3]
        lon2, lat2 = p2[2], p2[3]
        time1, time2 = p1[1], p2[1]

        # Calculate distance using Haversine formula
        R = 6371  # Earth radius in km
        dlat = math.radians(lat2 - lat1)
        dlon = math.radians(lon2 - lon1)
        a = (math.sin(dlat/2) ** 2 +
             math.cos(math.radians(lat1)) * math.cos(math.radians(lat2)) *
             math.sin(dlon/2) ** 2)
        c = 2 * math.atan2(math.sqrt(a), math.sqrt(1-a))
        distance_km = R * c

        # Calculate time difference in hours
        time_diff_hours = (time2 - time1) / 3600.0

        if time_diff_hours == 0:
            return 0

        return distance_km / time_diff_hours

    def classify_point(self, point: Tuple, prev_point: Optional[Tuple],
                      next_point: Optional[Tuple]) -> Dict[str, Any]:
        """
        Classify a single point's transport mode

        Args:
            point: Current point
            prev_point: Previous point (or None)
            next_point: Next point (or None)

        Returns:
            Dict with mode, confidence, and reason_codes
        """
        # Extract point data
        altitude = point[8] if len(point) > 8 else 0
        speed_recorded = point[6] if len(point) > 6 else 0
        province = point[9] if len(point) > 9 else None
        prev_province = prev_point[9] if prev_point and len(prev_point) > 9 else None

        # Calculate actual speed if we have prev/next points
        speed = speed_recorded
        if prev_point:
            calculated_speed = self.calculate_speed(prev_point, point)
            speed = max(speed, calculated_speed)

        reason_codes = []

        # Rule 1: FLIGHT - High altitude and high speed
        if altitude > 1000 and 200 <= speed <= 1000:
            reason_codes.append("HIGH_ALTITUDE")
            reason_codes.append("FLIGHT_SPEED_RANGE")
            return {
                'mode': 'FLIGHT',
                'confidence': 0.95,
                'reason_codes': reason_codes
            }

        # Rule 2: TRAIN - High speed and crosses provinces
        crosses_province = (prev_province and province and
                           prev_province != province)
        if 80 <= speed <= 350:
            reason_codes.append("TRAIN_SPEED_RANGE")
            if crosses_province:
                reason_codes.append("CROSSES_PROVINCE")
                return {
                    'mode': 'TRAIN',
                    'confidence': 0.85,
                    'reason_codes': reason_codes
                }
            else:
                # Could be train or car
                return {
                    'mode': 'CAR',  # Default to CAR if no province crossing
                    'confidence': 0.70,
                    'reason_codes': ["HIGH_SPEED"]
                }

        # Rule 3: CAR - Moderate speed
        if 20 <= speed <= 120:
            reason_codes.append("CAR_SPEED_RANGE")
            return {
                'mode': 'CAR',
                'confidence': 0.75,
                'reason_codes': reason_codes
            }

        # Rule 4: WALK - Low speed
        if 1 <= speed < 10:
            reason_codes.append("WALKING_SPEED")
            return {
                'mode': 'WALK',
                'confidence': 0.80,
                'reason_codes': reason_codes
            }

        # Rule 5: STAY - Very low speed
        if speed < 1:
            reason_codes.append("STATIONARY")
            return {
                'mode': 'STAY',
                'confidence': 0.90,
                'reason_codes': reason_codes
            }

        # Default: UNKNOWN
        return {
            'mode': 'UNKNOWN',
            'confidence': 0.50,
            'reason_codes': ['NO_MATCHING_RULE']
        }

    def create_segment(self, points: List[Tuple], mode: str,
                      confidence: float, reason_codes: List[str]) -> Dict[str, Any]:
        """
        Create a segment record from a list of points

        Args:
            points: List of points in the segment
            mode: Transport mode
            confidence: Confidence score
            reason_codes: List of reason codes

        Returns:
            Segment dictionary
        """
        if not points:
            return None

        first_point = points[0]
        last_point = points[-1]

        start_time = first_point[1]
        end_time = last_point[1]
        duration_s = end_time - start_time

        # Calculate total distance
        total_distance = 0
        for i in range(len(points) - 1):
            p1, p2 = points[i], points[i+1]
            lon1, lat1 = p1[2], p1[3]
            lon2, lat2 = p2[2], p2[3]

            # Haversine distance
            R = 6371000  # Earth radius in meters
            dlat = math.radians(lat2 - lat1)
            dlon = math.radians(lon2 - lon1)
            a = (math.sin(dlat/2) ** 2 +
                 math.cos(math.radians(lat1)) * math.cos(math.radians(lat2)) *
                 math.sin(dlon/2) ** 2)
            c = 2 * math.atan2(math.sqrt(a), math.sqrt(1-a))
            total_distance += R * c

        # Calculate speeds
        avg_speed_kmh = (total_distance / 1000) / (duration_s / 3600) if duration_s > 0 else 0
        max_speed_kmh = max([p[6] if len(p) > 6 else 0 for p in points])

        return {
            'mode': mode,
            'start_time': start_time,
            'end_time': end_time,
            'start_point_id': first_point[0],
            'end_point_id': last_point[0],
            'point_count': len(points),
            'distance_m': total_distance,
            'duration_s': duration_s,
            'avg_speed_kmh': avg_speed_kmh,
            'max_speed_kmh': max_speed_kmh,
            'confidence': confidence,
            'reason_codes': json.dumps(reason_codes),
            'metadata': json.dumps({})
        }

    def process_batch(self, points: List[Tuple]) -> int:
        """
        Process a batch of points for transport mode classification

        Args:
            points: List of point tuples

        Returns:
            Number of failed points
        """
        failed = 0

        try:
            # Classify each point
            classifications = []
            for i, point in enumerate(points):
                prev_point = points[i-1] if i > 0 else None
                next_point = points[i+1] if i < len(points) - 1 else None

                classification = self.classify_point(point, prev_point, next_point)
                classifications.append((point, classification))

            # Segment by mode changes
            segments = []
            current_segment_points = []
            current_mode = None
            current_confidence = 0
            current_reasons = []

            for point, classification in classifications:
                mode = classification['mode']

                if mode != current_mode:
                    # Mode changed - save previous segment
                    if current_segment_points:
                        segment = self.create_segment(
                            current_segment_points,
                            current_mode,
                            current_confidence,
                            current_reasons
                        )
                        if segment:
                            segments.append(segment)

                    # Start new segment
                    current_segment_points = [point]
                    current_mode = mode
                    current_confidence = classification['confidence']
                    current_reasons = classification['reason_codes']
                else:
                    # Same mode - add to current segment
                    current_segment_points.append(point)

            # Save last segment
            if current_segment_points:
                segment = self.create_segment(
                    current_segment_points,
                    current_mode,
                    current_confidence,
                    current_reasons
                )
                if segment:
                    segments.append(segment)

            # Insert segments into database and update points
            segment_idx = 0
            for point, classification in classifications:
                # Find which segment this point belongs to
                if segment_idx < len(segments):
                    segment = segments[segment_idx]

                    # Check if we need to move to next segment
                    if point[1] > segment['end_time']:
                        segment_idx += 1
                        if segment_idx < len(segments):
                            segment = segments[segment_idx]

            # Now insert segments and update points
            for i, segment in enumerate(segments):
                cursor = self.conn.execute("""
                    INSERT INTO segments (
                        mode, start_time, end_time, start_point_id, end_point_id,
                        point_count, distance_m, duration_s, avg_speed_kmh,
                        max_speed_kmh, confidence, reason_codes, metadata
                    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                """, (
                    segment['mode'], segment['start_time'], segment['end_time'],
                    segment['start_point_id'], segment['end_point_id'],
                    segment['point_count'], segment['distance_m'], segment['duration_s'],
                    segment['avg_speed_kmh'], segment['max_speed_kmh'],
                    segment['confidence'], segment['reason_codes'], segment['metadata']
                ))
                segment_id = cursor.lastrowid

                # Update points in this segment
                self.conn.execute("""
                    UPDATE "一生足迹"
                    SET segment_id = ?,
                        mode = ?,
                        mode_confidence = ?,
                        mode_reason_codes = ?
                    WHERE id >= ? AND id <= ?
                """, (segment_id, segment['mode'], segment['confidence'],
                      segment['reason_codes'], segment['start_point_id'],
                      segment['end_point_id']))

            self.conn.commit()

        except Exception as e:
            self.logger.error(f"Error processing batch: {e}")
            failed = len(points)

        return failed

    def clear_previous_results(self):
        """Clear previous transport mode analysis results"""
        self.logger.info("Clearing previous transport mode results...")

        # Clear segment references in track points
        self.conn.execute("""
            UPDATE "一生足迹"
            SET segment_id = NULL,
                mode = NULL,
                mode_confidence = NULL,
                mode_reason_codes = NULL
        """)

        # Delete all segments
        self.conn.execute("DELETE FROM segments")

        self.conn.commit()
        self.logger.info("Previous results cleared")


def main():
    parser = argparse.ArgumentParser(description='Transport Mode Classification Worker')
    parser.add_argument('--db-path', required=True, help='Path to SQLite database')
    parser.add_argument('--task-id', type=int, required=True, help='Analysis task ID')
    parser.add_argument('--batch-size', type=int, default=1000, help='Batch size')

    args = parser.parse_args()

    worker = TransportModeWorker(args.db_path, args.task_id, args.batch_size)
    worker.run()


if __name__ == '__main__':
    main()
