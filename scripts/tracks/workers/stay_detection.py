#!/usr/bin/env python3
"""
Skill: stay_detection (Advanced)
Purpose: Detect stays using DBSCAN clustering with temporal-spatial constraints
Algorithm: DBSCAN on GPS points with adaptive epsilon and temporal continuity
"""

import sys
import json
import sqlite3
import numpy as np
from sklearn.cluster import DBSCAN
from datetime import datetime
from math import radians, cos, sin, asin, sqrt

class StayDetectionWorker:
    def __init__(self, db_path, task_id):
        self.db_path = db_path
        self.task_id = task_id
        self.conn = sqlite3.connect(db_path)
        self.conn.row_factory = sqlite3.Row

        # Algorithm parameters
        self.min_duration_s = 30 * 60  # 30 minutes
        self.spatial_eps_m = 200  # 200 meters radius
        self.min_samples = 3  # Minimum points for a cluster
        self.max_time_gap_s = 60 * 60  # 1 hour max gap within stay

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
                started_at = CURRENT_TIMESTAMP,
                progress = 0.0
            WHERE id = ?
        """, (self.task_id,))
        self.conn.commit()

    def update_progress(self, progress, message=""):
        """Update task progress"""
        cursor = self.conn.cursor()
        cursor.execute("""
            UPDATE analysis_tasks
            SET progress = ?,
                progress_message = ?
            WHERE id = ?
        """, (progress, message, self.task_id))
        self.conn.commit()

    def load_data(self):
        """Load track points from database"""
        cursor = self.conn.cursor()
        cursor.execute("""
            SELECT id, dataTime, latitude, longitude, province, city, county
            FROM "一生足迹"
            WHERE latitude IS NOT NULL
              AND longitude IS NOT NULL
              AND outlier_flag = 0
            ORDER BY dataTime
        """)
        rows = cursor.fetchall()
        return rows

    def temporal_spatial_dbscan(self, points):
        """
        Perform DBSCAN clustering with temporal-spatial constraints
        Returns: cluster labels for each point
        """
        if len(points) == 0:
            return np.array([])

        # Convert to numpy array for DBSCAN
        coords = np.array([[p['latitude'], p['longitude']] for p in points])

        # Custom distance metric: haversine distance in meters
        def haversine_metric(a, b):
            return self.haversine_distance(a[0], a[1], b[0], b[1])

        # Perform DBSCAN clustering
        # eps is in meters, converted to approximate degrees
        eps_degrees = self.spatial_eps_m / 111000  # Rough conversion
        db = DBSCAN(eps=eps_degrees, min_samples=self.min_samples, metric=haversine_metric)
        labels = db.fit_predict(coords)

        return labels

    def filter_by_temporal_continuity(self, points, labels):
        """
        Filter clusters to ensure temporal continuity
        Split clusters if time gaps exceed threshold
        """
        filtered_labels = labels.copy()
        unique_labels = set(labels)
        unique_labels.discard(-1)  # Remove noise label

        new_label = labels.max() + 1 if len(labels) > 0 else 0

        for label in unique_labels:
            # Get points in this cluster
            cluster_indices = np.where(labels == label)[0]
            cluster_points = [points[i] for i in cluster_indices]

            # Sort by time
            cluster_points_sorted = sorted(cluster_points, key=lambda p: p['dataTime'])

            # Check for time gaps
            current_group_start = 0
            for i in range(1, len(cluster_points_sorted)):
                time_gap = cluster_points_sorted[i]['dataTime'] - cluster_points_sorted[i-1]['dataTime']

                if time_gap > self.max_time_gap_s:
                    # Split cluster - assign new label to points after gap
                    for j in range(i, len(cluster_points_sorted)):
                        point_id = cluster_points_sorted[j]['id']
                        point_idx = next(idx for idx, p in enumerate(points) if p['id'] == point_id)
                        filtered_labels[point_idx] = new_label
                    new_label += 1
                    break

        return filtered_labels

    def process(self, points):
        """Main processing logic"""
        self.update_progress(0.2, "Performing DBSCAN clustering...")

        # Perform DBSCAN clustering
        labels = self.temporal_spatial_dbscan(points)

        self.update_progress(0.4, "Filtering by temporal continuity...")

        # Filter by temporal continuity
        labels = self.filter_by_temporal_continuity(points, labels)

        self.update_progress(0.6, "Extracting stay segments...")

        # Extract stay segments from clusters
        stays = []
        unique_labels = set(labels)
        unique_labels.discard(-1)  # Remove noise

        for label in unique_labels:
            cluster_indices = np.where(labels == label)[0]
            cluster_points = [points[i] for i in cluster_indices]

            # Calculate stay properties
            start_ts = min(p['dataTime'] for p in cluster_points)
            end_ts = max(p['dataTime'] for p in cluster_points)
            duration_s = end_ts - start_ts

            # Filter by minimum duration
            if duration_s < self.min_duration_s:
                continue

            # Calculate center point (mean of coordinates)
            center_lat = np.mean([p['latitude'] for p in cluster_points])
            center_lon = np.mean([p['longitude'] for p in cluster_points])

            # Calculate confidence based on cluster density
            point_count = len(cluster_points)
            confidence = min(1.0, point_count / 10.0)  # Max confidence at 10+ points

            # Get admin info from first point
            admin_info = cluster_points[0]

            stays.append({
                'start_ts': start_ts,
                'end_ts': end_ts,
                'duration_s': duration_s,
                'center_lat': center_lat,
                'center_lon': center_lon,
                'point_count': point_count,
                'confidence': confidence,
                'province': admin_info.get('province'),
                'city': admin_info.get('city'),
                'county': admin_info.get('county'),
                'cluster_id': int(label)
            })

        self.update_progress(0.8, f"Found {len(stays)} stay segments")

        return stays

    def save_results(self, stays):
        """Save results to database"""
        cursor = self.conn.cursor()

        # Clear existing stay segments
        cursor.execute("DELETE FROM stay_segments")

        # Insert new stay segments
        for stay in stays:
            cursor.execute("""
                INSERT INTO stay_segments (
                    start_ts, end_ts, duration_s,
                    center_lat, center_lon,
                    point_count, confidence,
                    province, city, county,
                    cluster_id, cluster_confidence,
                    algo_version
                ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'v1_dbscan')
            """, (
                stay['start_ts'], stay['end_ts'], stay['duration_s'],
                stay['center_lat'], stay['center_lon'],
                stay['point_count'], stay['confidence'],
                stay['province'], stay['city'], stay['county'],
                stay['cluster_id'], stay['confidence']
            ))

        self.conn.commit()

    def mark_completed(self, summary):
        """Mark task as completed with summary"""
        cursor = self.conn.cursor()
        cursor.execute("""
            UPDATE analysis_tasks
            SET status = 'completed',
                completed_at = CURRENT_TIMESTAMP,
                progress = 1.0,
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
                completed_at = CURRENT_TIMESTAMP,
                error_message = ?
            WHERE id = ?
        """, (error_msg, self.task_id))
        self.conn.commit()

    def run(self):
        """Execute the worker"""
        try:
            self.mark_running()
            points = self.load_data()
            self.update_progress(0.1, f"Loaded {len(points)} track points")

            stays = self.process(points)
            self.save_results(stays)

            summary = {
                'total_points': len(points),
                'stays_detected': len(stays),
                'total_stay_duration_hours': sum(s['duration_s'] for s in stays) / 3600,
                'avg_stay_duration_minutes': np.mean([s['duration_s'] for s in stays]) / 60 if stays else 0
            }

            self.mark_completed(summary)
            print(f"Stay detection completed: {len(stays)} stays detected")
            return 0

        except Exception as e:
            error_msg = f"Stay detection failed: {str(e)}"
            print(error_msg, file=sys.stderr)
            self.mark_failed(error_msg)
            return 1

        finally:
            self.conn.close()

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: python stay_detection.py <db_path> <task_id>")
        sys.exit(1)

    db_path = sys.argv[1]
    task_id = int(sys.argv[2])

    worker = StayDetectionWorker(db_path, task_id)
    sys.exit(worker.run())

