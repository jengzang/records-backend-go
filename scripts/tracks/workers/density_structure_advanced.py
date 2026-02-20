#!/usr/bin/env python3
"""
Skill: density_structure_advanced
Purpose: Advanced spatial density analysis using DBSCAN clustering
Algorithm: DBSCAN on all track points with cluster classification
"""

import sys
import json
import sqlite3
import numpy as np
from sklearn.cluster import DBSCAN
from scipy.spatial import ConvexHull
from datetime import datetime
from math import radians, cos, sin, asin, sqrt

class DensityStructureWorker:
    def __init__(self, db_path, task_id):
        self.db_path = db_path
        self.task_id = task_id
        self.conn = sqlite3.connect(db_path)
        self.conn.row_factory = sqlite3.Row

        # Algorithm parameters
        self.spatial_eps_m = 500  # 500 meters for density clustering
        self.min_samples = 10  # Minimum points for a density cluster

    def haversine_distance(self, lat1, lon1, lat2, lon2):
        """Calculate haversine distance in meters"""
        R = 6371000
        lat1, lon1, lat2, lon2 = map(radians, [lat1, lon1, lat2, lon2])
        dlat = lat2 - lat1
        dlon = lon2 - lon1
        a = sin(dlat/2)**2 + cos(lat1) * cos(lat2) * sin(dlon/2)**2
        c = 2 * asin(sqrt(a))
        return R * c

    def mark_running(self):
        cursor = self.conn.cursor()
        cursor.execute("""
            UPDATE analysis_tasks
            SET status = 'running', started_at = CURRENT_TIMESTAMP, progress = 0.0
            WHERE id = ?
        """, (self.task_id,))
        self.conn.commit()

    def update_progress(self, progress, message=""):
        cursor = self.conn.cursor()
        cursor.execute("""
            UPDATE analysis_tasks SET progress = ?, progress_message = ? WHERE id = ?
        """, (progress, message, self.task_id))
        self.conn.commit()

    def load_data(self):
        cursor = self.conn.cursor()
        cursor.execute("""
            SELECT id, dataTime, latitude, longitude, province, city, county
            FROM "一生足迹"
            WHERE latitude IS NOT NULL AND longitude IS NOT NULL AND outlier_flag = 0
        """)
        return cursor.fetchall()

    def classify_cluster_type(self, point_count, total_duration_s, visit_count):
        """Classify cluster by usage pattern"""
        avg_duration_hours = total_duration_s / 3600 / visit_count if visit_count > 0 else 0

        if point_count > 1000 and avg_duration_hours > 6:
            return 'HOME'
        elif point_count > 500 and avg_duration_hours > 4:
            return 'WORK'
        elif point_count > 200:
            return 'FREQUENT'
        else:
            return 'OCCASIONAL'

    def calculate_convex_hull_area(self, points):
        """Calculate convex hull area in km²"""
        if len(points) < 3:
            return 0.0

        try:
            coords = np.array([[p['latitude'], p['longitude']] for p in points])
            hull = ConvexHull(coords)
            # Approximate area (degrees² to km²)
            area_deg2 = hull.volume
            area_km2 = area_deg2 * 111 * 111  # Rough conversion
            return area_km2
        except:
            return 0.0

    def process(self, points):
        self.update_progress(0.2, "Performing DBSCAN clustering...")

        coords = np.array([[p['latitude'], p['longitude']] for p in points])

        def haversine_metric(a, b):
            return self.haversine_distance(a[0], a[1], b[0], b[1])

        eps_degrees = self.spatial_eps_m / 111000
        db = DBSCAN(eps=eps_degrees, min_samples=self.min_samples, metric=haversine_metric)
        labels = db.fit_predict(coords)

        self.update_progress(0.5, "Analyzing clusters...")

        clusters = []
        unique_labels = set(labels)
        unique_labels.discard(-1)

        for label in unique_labels:
            cluster_indices = np.where(labels == label)[0]
            cluster_points = [points[i] for i in cluster_indices]

            center_lat = np.mean([p['latitude'] for p in cluster_points])
            center_lon = np.mean([p['longitude'] for p in cluster_points])

            # Calculate density score
            point_count = len(cluster_points)
            density_score = point_count / (self.spatial_eps_m / 1000) ** 2

            # Calculate radius
            max_dist = max(self.haversine_distance(center_lat, center_lon, p['latitude'], p['longitude'])
                          for p in cluster_points)

            # Calculate convex hull area
            convex_hull_area = self.calculate_convex_hull_area(cluster_points)

            # Classify cluster type
            # Estimate visit count and duration (simplified)
            visit_count = len(set(p['dataTime'] // 86400 for p in cluster_points))
            total_duration_s = len(cluster_points) * 60  # Assume 1 min per point
            cluster_type = self.classify_cluster_type(point_count, total_duration_s, visit_count)

            confidence = min(1.0, point_count / 50.0)

            admin_info = cluster_points[0]

            clusters.append({
                'cluster_id': int(label),
                'center_lat': center_lat,
                'center_lon': center_lon,
                'point_count': point_count,
                'density_score': density_score,
                'cluster_type': cluster_type,
                'radius_m': max_dist,
                'convex_hull_area_km2': convex_hull_area,
                'province': admin_info.get('province'),
                'city': admin_info.get('city'),
                'county': admin_info.get('county'),
                'confidence': confidence
            })

        self.update_progress(0.8, f"Found {len(clusters)} density clusters")
        return clusters

    def save_results(self, clusters):
        cursor = self.conn.cursor()
        cursor.execute("DELETE FROM density_clusters")

        for cluster in clusters:
            cursor.execute("""
                INSERT INTO density_clusters (
                    cluster_id, center_lat, center_lon, point_count, density_score,
                    cluster_type, radius_m, convex_hull_area_km2,
                    province, city, county, confidence, algo_version
                ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'v1_dbscan')
            """, (
                cluster['cluster_id'], cluster['center_lat'], cluster['center_lon'],
                cluster['point_count'], cluster['density_score'], cluster['cluster_type'],
                cluster['radius_m'], cluster['convex_hull_area_km2'],
                cluster['province'], cluster['city'], cluster['county'], cluster['confidence']
            ))

        self.conn.commit()

    def mark_completed(self, summary):
        cursor = self.conn.cursor()
        cursor.execute("""
            UPDATE analysis_tasks
            SET status = 'completed', completed_at = CURRENT_TIMESTAMP,
                progress = 1.0, result_summary = ?
            WHERE id = ?
        """, (json.dumps(summary), self.task_id))
        self.conn.commit()

    def mark_failed(self, error_msg):
        cursor = self.conn.cursor()
        cursor.execute("""
            UPDATE analysis_tasks
            SET status = 'failed', completed_at = CURRENT_TIMESTAMP, error_message = ?
            WHERE id = ?
        """, (error_msg, self.task_id))
        self.conn.commit()

    def run(self):
        try:
            self.mark_running()
            points = self.load_data()
            self.update_progress(0.1, f"Loaded {len(points)} track points")

            clusters = self.process(points)
            self.save_results(clusters)

            summary = {
                'total_points': len(points),
                'clusters_found': len(clusters),
                'home_clusters': sum(1 for c in clusters if c['cluster_type'] == 'HOME'),
                'work_clusters': sum(1 for c in clusters if c['cluster_type'] == 'WORK'),
                'frequent_clusters': sum(1 for c in clusters if c['cluster_type'] == 'FREQUENT')
            }

            self.mark_completed(summary)
            print(f"Density analysis completed: {len(clusters)} clusters found")
            return 0

        except Exception as e:
            error_msg = f"Density analysis failed: {str(e)}"
            print(error_msg, file=sys.stderr)
            self.mark_failed(error_msg)
            return 1

        finally:
            self.conn.close()

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: python density_structure_advanced.py <db_path> <task_id>")
        sys.exit(1)

    db_path = sys.argv[1]
    task_id = int(sys.argv[2])

    worker = DensityStructureWorker(db_path, task_id)
    sys.exit(worker.run())
