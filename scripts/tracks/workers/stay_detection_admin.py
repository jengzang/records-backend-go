#!/usr/bin/env python3
"""
Skill: stay_detection_admin (Advanced)
Purpose: Detect stays based on administrative area (RLE algorithm)
Algorithm: Run-Length Encoding with boundary smoothing
"""

import sys
import json
import sqlite3
import numpy as np
from datetime import datetime
from math import radians, cos, sin, asin, sqrt

class AdminAreaStayDetectionWorker:
    def __init__(self, db_path, task_id, profile_id=None):
        self.db_path = db_path
        self.task_id = task_id
        self.conn = sqlite3.connect(db_path)
        self.conn.row_factory = sqlite3.Row

        # Algorithm parameters
        if profile_id:
            self.load_profile(profile_id)
        else:
            self.use_default_params()

    def use_default_params(self):
        """Use default parameters"""
        self.admin_level = 'city'  # Options: province/city/county/town
        self.min_duration_s = 4 * 60 * 60  # 4 hours (admin stays require longer duration)
        self.merge_gap_s = 30 * 60  # 30 minutes (merge brief exits)
        self.min_points = 3  # Minimum points for a stay

    def load_profile(self, profile_id):
        """Load parameters from threshold_profiles table"""
        cursor = self.conn.cursor()
        cursor.execute('SELECT params_json FROM threshold_profiles WHERE id = ?', (profile_id,))
        row = cursor.fetchone()
        if row:
            params = json.loads(row[0])
            self.admin_level = params.get('admin_level', 'city')
            self.min_duration_s = params.get('min_duration_s', 4 * 60 * 60)
            self.merge_gap_s = params.get('merge_gap_s', 30 * 60)
            self.min_points = params.get('min_points', 3)
        else:
            self.use_default_params()

    def haversine_distance(self, lat1, lon1, lat2, lon2):
        """Calculate haversine distance in meters"""
        R = 6371000  # Earth radius in meters
        lat1, lon1, lat2, lon2 = map(radians, [lat1, lon1, lat2, lon2])
        dlat = lat2 - lat1
        dlon = lon2 - lon1
        a = sin(dlat/2)**2 + cos(lat1) * cos(lat2) * sin(dlon/2)**2
        c = 2 * asin(sqrt(a))
        return R * c

    def mark_running(self):
        """Mark task as running"""
        cursor = self.conn.cursor()
        cursor.execute("""
            UPDATE analysis_tasks
            SET status = 'running',
                start_time = CAST(strftime('%s', 'now') AS INTEGER),
                progress_percent = 0
            WHERE id = ?
        """, (self.task_id,))
        self.conn.commit()

    def update_progress(self, progress, message=""):
        """Update task progress"""
        cursor = self.conn.cursor()
        cursor.execute("""
            UPDATE analysis_tasks
            SET progress_percent = ?,
                result_summary = ?
            WHERE id = ?
        """, (int(progress * 100), message, self.task_id))
        self.conn.commit()

    def load_data(self):
        """Load track points from database"""
        cursor = self.conn.cursor()
        cursor.execute(f"""
            SELECT id, dataTime, latitude, longitude, province, city, county, town, village
            FROM "一生足迹"
            WHERE latitude IS NOT NULL
              AND longitude IS NOT NULL
              AND outlier_flag = 0
              AND {self.admin_level} IS NOT NULL
            ORDER BY dataTime
        """)
        rows = cursor.fetchall()
        return [dict(row) for row in rows]

    def detect_admin_stays(self, points):
        """Detect stays using Run-Length Encoding on administrative areas"""
        stays = []

        if len(points) == 0:
            return stays

        # Run-Length Encoding: group consecutive points by admin area
        current_admin = points[0][self.admin_level]
        segment_start = 0

        for i in range(1, len(points)):
            if points[i][self.admin_level] != current_admin:
                # Admin area changed, process current segment
                segment = self.create_admin_stay_segment(
                    points[segment_start:i],
                    current_admin
                )
                if segment:
                    stays.append(segment)

                # Start new segment
                current_admin = points[i][self.admin_level]
                segment_start = i

        # Process last segment
        segment = self.create_admin_stay_segment(
            points[segment_start:],
            current_admin
        )
        if segment:
            stays.append(segment)

        # Merge nearby stays (handle brief exits)
        stays = self.merge_nearby_stays(stays)

        return stays

    def create_admin_stay_segment(self, points, admin_area):
        """Create an admin area stay segment from points"""
        if len(points) < self.min_points:
            return None

        start_time = points[0]['dataTime']
        end_time = points[-1]['dataTime']
        duration_s = end_time - start_time

        # Filter by minimum duration
        if duration_s < self.min_duration_s:
            return None

        # Calculate center point (median of coordinates for robustness)
        lats = [p['latitude'] for p in points]
        lons = [p['longitude'] for p in points]
        center_lat = np.median(lats)
        center_lon = np.median(lons)

        # Calculate radius (max distance from center)
        radius_m = max(
            self.haversine_distance(center_lat, center_lon, p['latitude'], p['longitude'])
            for p in points
        )

        # Calculate confidence (based on point count and duration)
        point_count = len(points)
        confidence = min(1.0, point_count / 20.0)

        # Create reason codes
        reason_codes = ['admin_area_rle', 'min_duration_met']
        if point_count >= 20:
            reason_codes.append('high_point_count')

        # Create metadata
        metadata = {
            'algorithm': 'rle',
            'admin_level': self.admin_level,
            'admin_area': admin_area
        }

        return {
            'stay_type': 'ADMIN_AREA',
            'start_time': start_time,
            'end_time': end_time,
            'duration_s': duration_s,
            'center_lat': center_lat,
            'center_lon': center_lon,
            'radius_m': radius_m,
            'point_count': point_count,
            'confidence': confidence,
            'province': points[0].get('province'),
            'city': points[0].get('city'),
            'county': points[0].get('county'),
            'town': points[0].get('town'),
            'village': points[0].get('village'),
            'reason_codes': json.dumps(reason_codes),
            'metadata': json.dumps(metadata)
        }

    def merge_nearby_stays(self, stays):
        """Merge stays separated by brief exits from the same admin area"""
        if len(stays) <= 1:
            return stays

        merged = []
        current = stays[0].copy()

        for next_stay in stays[1:]:
            # Check if same admin area and gap is small
            current_admin = json.loads(current['metadata'])['admin_area']
            next_admin = json.loads(next_stay['metadata'])['admin_area']

            if (current_admin == next_admin and
                next_stay['start_time'] - current['end_time'] <= self.merge_gap_s):
                # Merge: extend current stay
                current['end_time'] = next_stay['end_time']
                current['duration_s'] = current['end_time'] - current['start_time']
                current['point_count'] += next_stay['point_count']
                # Recalculate confidence
                current['confidence'] = min(1.0, current['point_count'] / 20.0)
            else:
                # Cannot merge, save current and start new
                merged.append(current)
                current = next_stay.copy()

        merged.append(current)
        return merged

    def save_results(self, stays):
        """Save results to database (append to existing stays)"""
        cursor = self.conn.cursor()

        # Delete existing ADMIN_AREA stays to avoid duplicates
        cursor.execute("DELETE FROM stay_segments WHERE stay_type = 'ADMIN_AREA'")

        # Insert new stay segments
        for stay in stays:
            cursor.execute("""
                INSERT INTO stay_segments (
                    stay_type, start_time, end_time, duration_s,
                    center_lat, center_lon, radius_m,
                    province, city, county, town, village,
                    point_count, confidence,
                    reason_codes, metadata,
                    algo_version, created_at, updated_at
                ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'v1_rle',
                          CAST(strftime('%s', 'now') AS INTEGER),
                          CAST(strftime('%s', 'now') AS INTEGER))
            """, (
                stay['stay_type'], stay['start_time'], stay['end_time'], stay['duration_s'],
                stay['center_lat'], stay['center_lon'], stay['radius_m'],
                stay['province'], stay['city'], stay['county'], stay['town'], stay['village'],
                stay['point_count'], stay['confidence'],
                stay['reason_codes'], stay['metadata']
            ))

        self.conn.commit()

    def mark_completed(self, summary):
        """Mark task as completed with summary"""
        cursor = self.conn.cursor()
        cursor.execute("""
            UPDATE analysis_tasks
            SET status = 'completed',
                end_time = CAST(strftime('%s', 'now') AS INTEGER),
                progress_percent = 100,
                result_summary = ?
            WHERE id = ?
        """, (json.dumps(summary), self.task_id))
        self.conn.commit()

    def mark_failed(self, error_msg):
        """Mark task as failed"""
        cursor = self.conn.cursor()
        cursor.execute("""
            UPDATE analysis_tasks
            SET status = 'failed',
                end_time = CAST(strftime('%s', 'now') AS INTEGER),
                result_summary = ?
            WHERE id = ?
        """, (error_msg, self.task_id))
        self.conn.commit()

    def run(self):
        """Execute the worker"""
        try:
            self.mark_running()
            points = self.load_data()
            self.update_progress(0.1, f"Loaded {len(points)} track points")

            self.update_progress(0.3, "Detecting admin area stays...")
            stays = self.detect_admin_stays(points)

            self.update_progress(0.8, f"Saving {len(stays)} stays...")
            self.save_results(stays)

            summary = {
                'total_points': len(points),
                'stays_detected': len(stays),
                'admin_level': self.admin_level,
                'total_stay_duration_hours': sum(s['duration_s'] for s in stays) / 3600,
                'avg_stay_duration_hours': np.mean([s['duration_s'] for s in stays]) / 3600 if stays else 0
            }

            self.mark_completed(summary)
            print(f"Admin area stay detection completed: {len(stays)} stays detected")
            return 0

        except Exception as e:
            error_msg = f"Admin area stay detection failed: {str(e)}"
            print(error_msg, file=sys.stderr)
            import traceback
            traceback.print_exc()
            self.mark_failed(error_msg)
            return 1

        finally:
            self.conn.close()

if __name__ == "__main__":
    if len(sys.argv) < 3:
        print("Usage: python stay_detection_admin.py <db_path> <task_id> [profile_id]")
        sys.exit(1)

    db_path = sys.argv[1]
    task_id = int(sys.argv[2])
    profile_id = int(sys.argv[3]) if len(sys.argv) > 3 else None

    worker = AdminAreaStayDetectionWorker(db_path, task_id, profile_id)
    sys.exit(worker.run())

