#!/usr/bin/env python3
"""
Skill: trip_construction_advanced
Purpose: Advanced trip construction with ML-based purpose inference
Algorithm: Combine segments/stays into trips with purpose classification
"""

import sys
import json
import sqlite3
import numpy as np
from datetime import datetime

class TripConstructionAdvancedWorker:
    def __init__(self, db_path, task_id):
        self.db_path = db_path
        self.task_id = task_id
        self.conn = sqlite3.connect(db_path)
        self.conn.row_factory = sqlite3.Row

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

    def load_trips(self):
        cursor = self.conn.cursor()
        cursor.execute("""
            SELECT * FROM trips ORDER BY start_ts
        """)
        return cursor.fetchall()

    def extract_features(self, trip):
        """Extract features for ML classification"""
        # Time features
        start_dt = datetime.fromtimestamp(trip['start_ts'])
        hour = start_dt.hour
        day_of_week = start_dt.weekday()  # 0=Monday, 6=Sunday
        is_weekend = 1 if day_of_week >= 5 else 0

        # Distance and duration features
        distance_km = trip['total_distance_m'] / 1000 if trip['total_distance_m'] else 0
        duration_hours = trip['duration_s'] / 3600 if trip['duration_s'] else 0

        # Location features
        is_same_city = 1 if trip['origin_city'] == trip['dest_city'] else 0

        return {
            'hour': hour,
            'day_of_week': day_of_week,
            'is_weekend': is_weekend,
            'distance_km': distance_km,
            'duration_hours': duration_hours,
            'is_same_city': is_same_city,
            'primary_mode': trip['primary_mode']
        }

    def infer_purpose(self, trip, features):
        """Rule-based purpose inference (simplified ML)"""
        hour = features['hour']
        is_weekend = features['is_weekend']
        distance_km = features['distance_km']
        duration_hours = features['duration_hours']
        is_same_city = features['is_same_city']

        # Rule-based classification
        confidence = 0.5  # Base confidence

        # COMMUTE: weekday morning/evening, short distance, same city
        if not is_weekend and is_same_city and distance_km < 20:
            if (7 <= hour <= 9) or (17 <= hour <= 19):
                return 'COMMUTE', 0.8

        # WORK: weekday daytime, medium duration
        if not is_weekend and (9 <= hour <= 17) and duration_hours > 2:
            return 'WORK', 0.7

        # LEISURE: weekend or evening, various distances
        if is_weekend or (hour >= 19 or hour <= 7):
            if distance_km < 50:
                return 'LEISURE', 0.7

        # SHOPPING: short trips, daytime
        if is_same_city and distance_km < 10 and duration_hours < 2:
            if 10 <= hour <= 20:
                return 'SHOPPING', 0.6

        # TRAVEL: long distance, any time
        if distance_km > 100 or not is_same_city:
            return 'TRAVEL', 0.8

        # Default: OTHER
        return 'OTHER', 0.4

    def process(self, trips):
        self.update_progress(0.2, "Extracting features...")

        results = []
        for i, trip in enumerate(trips):
            features = self.extract_features(trip)
            purpose, confidence = self.infer_purpose(trip, features)

            results.append({
                'trip_id': trip['id'],
                'purpose_ml': purpose,
                'confidence_ml': confidence,
                'features_json': json.dumps(features)
            })

            if (i + 1) % 100 == 0:
                progress = 0.2 + 0.6 * (i + 1) / len(trips)
                self.update_progress(progress, f"Processed {i + 1}/{len(trips)} trips")

        self.update_progress(0.8, f"Classified {len(results)} trips")
        return results

    def save_results(self, results):
        cursor = self.conn.cursor()

        for result in results:
            cursor.execute("""
                UPDATE trips
                SET purpose_ml = ?,
                    confidence_ml = ?,
                    features_json = ?
                WHERE id = ?
            """, (
                result['purpose_ml'],
                result['confidence_ml'],
                result['features_json'],
                result['trip_id']
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
            trips = self.load_trips()
            self.update_progress(0.1, f"Loaded {len(trips)} trips")

            results = self.process(trips)
            self.save_results(results)

            # Calculate summary statistics
            purpose_counts = {}
            for r in results:
                purpose = r['purpose_ml']
                purpose_counts[purpose] = purpose_counts.get(purpose, 0) + 1

            summary = {
                'total_trips': len(results),
                'purpose_distribution': purpose_counts,
                'avg_confidence': np.mean([r['confidence_ml'] for r in results])
            }

            self.mark_completed(summary)
            print(f"Trip purpose classification completed: {len(results)} trips")
            return 0

        except Exception as e:
            error_msg = f"Trip classification failed: {str(e)}"
            print(error_msg, file=sys.stderr)
            self.mark_failed(error_msg)
            return 1

        finally:
            self.conn.close()

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: python trip_construction_advanced.py <db_path> <task_id>")
        sys.exit(1)

    db_path = sys.argv[1]
    task_id = int(sys.argv[2])

    worker = TripConstructionAdvancedWorker(db_path, task_id)
    sys.exit(worker.run())
