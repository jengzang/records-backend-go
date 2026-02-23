#!/usr/bin/env python3
"""
Skill: stay_detection_optimized (Advanced)
Purpose: Detect stays using DBSCAN with time-windowed batch processing
Algorithm: DBSCAN on GPS points with temporal batching for performance
Optimization: Process data in monthly windows to avoid O(n²) on full dataset
"""

import sys
import json
import sqlite3
import numpy as np
from sklearn.cluster import DBSCAN
from datetime import datetime
from math import radians, cos, sin, asin, sqrt
from collections import defaultdict

class OptimizedStayDetectionWorker:
    def __init__(self, db_path, task_id, profile_id=None):
        self.db_path = db_path
        self.task_id = task_id
        self.conn = sqlite3.connect(db_path)
        self.conn.row_factory = sqlite3.Row

        # Algorithm parameters (optimized 2026-02-23)
        if profile_id:
            self.load_profile(profile_id)
        else:
            self.use_default_params()

        # Batch processing parameters
        self.batch_window_days = 30  # Process 30 days at a time

    def use_default_params(self):
        """Use default parameters"""
        self.min_duration_s = 2 * 60 * 60  # 2 hours
        self.spatial_eps_m = 100  # 100 meters
        self.min_samples = 10  # 10 points minimum
        self.max_time_gap_s = 15 * 60  # 15 minutes

    def load_profile(self, profile_id):
        """Load parameters from threshold_profiles table"""
        cursor = self.conn.cursor()
        cursor.execute('SELECT params_json FROM threshold_profiles WHERE id = ?', (profile_id,))
        row = cursor.fetchone()
        if row:
            params = json.loads(row[0])
            self.min_duration_s = params.get('min_duration_s', 2 * 60 * 60)
            self.spatial_eps_m = params.get('spatial_eps_m', 100)
            self.min_samples = params.get('min_samples', 10)
            self.max_time_gap_s = params.get('max_time_gap_s', 15 * 60)
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

    def get_time_windows(self):
        """Get time windows for batch processing"""
        cursor = self.conn.cursor()
        cursor.execute("""
            SELECT MIN(dataTime) as min_time, MAX(dataTime) as max_time
            FROM "一生足迹"
            WHERE latitude IS NOT NULL
              AND longitude IS NOT NULL
              AND outlier_flag = 0
        """)
        row = cursor.fetchone()
        min_time, max_time = row['min_time'], row['max_time']

        # Create windows
        window_size_s = self.batch_window_days * 24 * 60 * 60
        windows = []
        current_start = min_time
        while current_start < max_time:
            current_end = min(current_start + window_size_s, max_time)
            windows.append((current_start, current_end))
            current_start = current_end

        return windows

    def load_window_data(self, start_time, end_time):
        """Load track points for a specific time window"""
        cursor = self.conn.cursor()
        cursor.execute("""
            SELECT id, dataTime, latitude, longitude, province, city, county, town, village
            FROM "一生足迹"
            WHERE latitude IS NOT NULL
              AND longitude IS NOT NULL
              AND outlier_flag = 0
              AND dataTime >= ?
              AND dataTime < ?
            ORDER BY dataTime
        """, (start_time, end_time))
        rows = cursor.fetchall()
        return [dict(row) for row in rows]

    def temporal_spatial_dbscan(self, points):
        """Perform DBSCAN clustering with temporal-spatial constraints"""
        if len(points) == 0:
            return np.array([])

        # Convert to numpy array for DBSCAN
        coords = np.array([[p['latitude'], p['longitude']] for p in points])

        # Custom distance metric: haversine distance in meters
        def haversine_metric(a, b):
            return self.haversine_distance(a[0], a[1], b[0], b[1])

        # Perform DBSCAN clustering
        eps_degrees = self.spatial_eps_m / 111000  # Rough conversion
        db = DBSCAN(eps=eps_degrees, min_samples=self.min_samples, metric=haversine_metric)
        labels = db.fit_predict(coords)

        return labels

    def filter_by_temporal_continuity(self, points, labels):
        """
        Filter clusters to ensure temporal continuity
        Split clusters if time gaps exceed threshold (recursive splitting)
        Fixed 2026-02-23: Now recursively splits all gaps
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

            # Find all split positions
            split_positions = []
            for i in range(1, len(cluster_points_sorted)):
                time_gap = cluster_points_sorted[i]['dataTime'] - cluster_points_sorted[i-1]['dataTime']
                if time_gap > self.max_time_gap_s:
                    split_positions.append(i)

            # If there are gaps, split into sub-clusters
            if split_positions:
                sub_clusters = []
                start = 0
                for pos in split_positions:
                    sub_clusters.append(cluster_points_sorted[start:pos])
                    start = pos
                sub_clusters.append(cluster_points_sorted[start:])

                # Assign new labels to each sub-cluster
                for sub_cluster in sub_clusters:
                    if len(sub_cluster) >= self.min_samples:
                        for point in sub_cluster:
                            point_idx = next(idx for idx, p in enumerate(points) if p['id'] == point['id'])
                            filtered_labels[point_idx] = new_label
                        new_label += 1
                    else:
                        for point in sub_cluster:
                            point_idx = next(idx for idx, p in enumerate(points) if p['id'] == point['id'])
                            filtered_labels[point_idx] = -1

        return filtered_labels

    def process_window(self, points):
        """Process a single time window"""
        if len(points) == 0:
            return []

        # Perform DBSCAN clustering
        labels = self.temporal_spatial_dbscan(points)

        # Filter by temporal continuity
        labels = self.filter_by_temporal_continuity(points, labels)

        # Extract stay segments from clusters
        stays = []
        unique_labels = set(labels)
        unique_labels.discard(-1)

        for label in unique_labels:
            cluster_indices = np.where(labels == label)[0]
            cluster_points = [points[i] for i in cluster_indices]

            # Calculate stay properties
            start_time = min(p['dataTime'] for p in cluster_points)
            end_time = max(p['dataTime'] for p in cluster_points)
            duration_s = end_time - start_time

            # Filter by minimum duration
            if duration_s < self.min_duration_s:
                continue

            # Calculate point density
            point_count = len(cluster_points)
            point_density_per_hour = point_count / (duration_s / 3600) if duration_s > 0 else 0

            # Filter by minimum point density
            if point_density_per_hour < 2.0:
                continue

            # Calculate center point
            center_lat = np.mean([p['latitude'] for p in cluster_points])
            center_lon = np.mean([p['longitude'] for p in cluster_points])

            # Calculate radius
            radius_m = max(
                self.haversine_distance(center_lat, center_lon, p['latitude'], p['longitude'])
                for p in cluster_points
            )

            # Calculate confidence
            confidence = min(1.0, point_count / 10.0)

            # Get admin info
            admin_info = dict(cluster_points[0])

            # Create reason codes
            reason_codes = ['dbscan_cluster', 'min_duration_met', 'temporal_continuity_verified']
            if point_count >= 10:
                reason_codes.append('high_point_density')
            if point_density_per_hour >= 5.0:
                reason_codes.append('very_high_density')

            # Create metadata
            metadata = {
                'algorithm': 'dbscan_windowed',
                'eps_m': self.spatial_eps_m,
                'min_samples': self.min_samples,
                'cluster_label': int(label),
                'point_density_per_hour': round(point_density_per_hour, 2)
            }

            stays.append({
                'stay_type': 'SPATIAL',
                'start_time': start_time,
                'end_time': end_time,
                'duration_s': duration_s,
                'center_lat': center_lat,
                'center_lon': center_lon,
                'radius_m': radius_m,
                'point_count': point_count,
                'confidence': confidence,
                'province': admin_info.get('province'),
                'city': admin_info.get('city'),
                'county': admin_info.get('county'),
                'town': admin_info.get('town'),
                'village': admin_info.get('village'),
                'reason_codes': json.dumps(reason_codes),
                'metadata': json.dumps(metadata)
            })

        return stays

    def merge_cross_window_stays(self, all_stays):
        """Merge stays that span across window boundaries"""
        if len(all_stays) == 0:
            return []

        # Sort by start time
        all_stays.sort(key=lambda s: s['start_time'])

        merged = []
        current = all_stays[0]

        for next_stay in all_stays[1:]:
            # Check if stays are close in time and space
            time_gap = next_stay['start_time'] - current['end_time']
            distance = self.haversine_distance(
                current['center_lat'], current['center_lon'],
                next_stay['center_lat'], next_stay['center_lon']
            )

            # Merge if within spatial_eps and max_time_gap
            if time_gap <= self.max_time_gap_s and distance <= self.spatial_eps_m:
                # Merge: extend current stay
                current['end_time'] = next_stay['end_time']
                current['duration_s'] = current['end_time'] - current['start_time']
                current['point_count'] += next_stay['point_count']
                # Recalculate center (weighted average)
                total_points = current['point_count']
                current['center_lat'] = (current['center_lat'] * (total_points - next_stay['point_count']) +
                                        next_stay['center_lat'] * next_stay['point_count']) / total_points
                current['center_lon'] = (current['center_lon'] * (total_points - next_stay['point_count']) +
                                        next_stay['center_lon'] * next_stay['point_count']) / total_points
                current['confidence'] = min(1.0, total_points / 10.0)
            else:
                merged.append(current)
                current = next_stay

        merged.append(current)
        return merged

    def save_results(self, stays):
        """Save results to database"""
        cursor = self.conn.cursor()

        # Clear existing SPATIAL stays
        cursor.execute("DELETE FROM stay_segments WHERE stay_type = 'SPATIAL'")

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
                ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'v2_dbscan_windowed',
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
        """Execute the worker with time-windowed processing"""
        try:
            self.mark_running()

            # Get time windows
            windows = self.get_time_windows()
            self.update_progress(0.05, f"Processing {len(windows)} time windows...")

            # Process each window
            all_stays = []
            for i, (start_time, end_time) in enumerate(windows):
                progress = 0.05 + (i / len(windows)) * 0.85
                self.update_progress(progress, f"Processing window {i+1}/{len(windows)}...")

                points = self.load_window_data(start_time, end_time)
                if len(points) > 0:
                    window_stays = self.process_window(points)
                    all_stays.extend(window_stays)

            # Merge cross-window stays
            self.update_progress(0.90, "Merging cross-window stays...")
            merged_stays = self.merge_cross_window_stays(all_stays)

            # Save results
            self.update_progress(0.95, f"Saving {len(merged_stays)} stays...")
            self.save_results(merged_stays)

            summary = {
                'windows_processed': len(windows),
                'stays_detected': len(merged_stays),
                'total_stay_duration_hours': sum(s['duration_s'] for s in merged_stays) / 3600,
                'avg_stay_duration_minutes': np.mean([s['duration_s'] for s in merged_stays]) / 60 if merged_stays else 0,
                'optimization': 'time_windowed_dbscan'
            }

            self.mark_completed(summary)
            print(f"Optimized stay detection completed: {len(merged_stays)} stays detected")
            return 0

        except Exception as e:
            error_msg = f"Optimized stay detection failed: {str(e)}"
            print(error_msg, file=sys.stderr)
            import traceback
            traceback.print_exc()
            self.mark_failed(error_msg)
            return 1

        finally:
            self.conn.close()

if __name__ == "__main__":
    if len(sys.argv) < 3:
        print("Usage: python stay_detection_optimized.py <db_path> <task_id> [profile_id]")
        sys.exit(1)

    db_path = sys.argv[1]
    task_id = int(sys.argv[2])
    profile_id = int(sys.argv[3]) if len(sys.argv) > 3 else None

    worker = OptimizedStayDetectionWorker(db_path, task_id, profile_id)
    sys.exit(worker.run())
