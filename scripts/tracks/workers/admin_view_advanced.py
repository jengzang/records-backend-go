#!/usr/bin/env python3
"""
Skill: admin_view_engine_advanced
Purpose: Advanced administrative analytics with time-series trends
Algorithm: Time-series analysis, trend detection, anomaly detection
"""

import sys
import json
import sqlite3
import numpy as np
from scipy import stats
from datetime import datetime, timedelta

class AdminViewAdvancedWorker:
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

    def load_admin_stats(self):
        cursor = self.conn.cursor()
        cursor.execute("""
            SELECT admin_level, admin_name, visit_count, total_duration_s, unique_days,
                   first_visit_ts, last_visit_ts
            FROM admin_stats
            WHERE visit_count > 0
        """)
        return cursor.fetchall()

    def load_time_series(self, admin_level, admin_name):
        """Load visit time series for an admin region"""
        cursor = self.conn.cursor()
        cursor.execute("""
            SELECT dataTime
            FROM "一生足迹"
            WHERE {} = ?
            ORDER BY dataTime
        """.format(admin_level.lower()), (admin_name,))

        timestamps = [row['dataTime'] for row in cursor.fetchall()]

        if not timestamps:
            return []

        # Group by month
        monthly_counts = {}
        for ts in timestamps:
            dt = datetime.fromtimestamp(ts)
            month_key = dt.strftime('%Y-%m')
            monthly_counts[month_key] = monthly_counts.get(month_key, 0) + 1

        return monthly_counts

    def detect_trend(self, time_series):
        """Detect trend using linear regression"""
        if len(time_series) < 3:
            return 'STABLE', 0.0, False

        # Convert to arrays
        months = sorted(time_series.keys())
        values = [time_series[m] for m in months]
        x = np.arange(len(values))

        # Linear regression
        slope, intercept, r_value, p_value, std_err = stats.linregress(x, values)

        # Determine trend type
        if abs(slope) < 0.1:
            trend_type = 'STABLE'
        elif slope > 0:
            trend_type = 'GROWTH'
        else:
            trend_type = 'DECLINE'

        # Trend score: normalized slope (-1 to 1)
        max_value = max(values) if values else 1
        trend_score = np.clip(slope / max_value, -1, 1)

        # Detect seasonality (simplified: check if variance is high)
        if len(values) >= 12:
            variance = np.var(values)
            mean = np.mean(values)
            cv = variance / mean if mean > 0 else 0
            seasonality_detected = cv > 0.5
        else:
            seasonality_detected = False

        return trend_type, trend_score, seasonality_detected

    def detect_anomalies(self, time_series):
        """Detect anomalies using z-score method"""
        if len(time_series) < 3:
            return []

        months = sorted(time_series.keys())
        values = [time_series[m] for m in months]

        # Calculate z-scores
        mean = np.mean(values)
        std = np.std(values)

        if std == 0:
            return []

        anomalies = []
        for i, (month, value) in enumerate(zip(months, values)):
            z_score = abs((value - mean) / std)
            if z_score > 2.5:  # Threshold for anomaly
                anomalies.append({
                    'month': month,
                    'value': value,
                    'z_score': float(z_score)
                })

        return anomalies

    def predict_next_month(self, time_series, trend_score):
        """Simple prediction based on trend"""
        if len(time_series) < 2:
            return 0

        months = sorted(time_series.keys())
        recent_values = [time_series[m] for m in months[-3:]]  # Last 3 months
        avg_recent = np.mean(recent_values)

        # Apply trend
        prediction = int(avg_recent * (1 + trend_score))
        return max(0, prediction)

    def process(self, admin_stats):
        self.update_progress(0.2, "Analyzing trends...")

        results = []
        total = len(admin_stats)

        for i, stat in enumerate(admin_stats):
            admin_level = stat['admin_level']
            admin_name = stat['admin_name']

            # Load time series
            time_series = self.load_time_series(admin_level, admin_name)

            if not time_series:
                continue

            # Detect trend
            trend_type, trend_score, seasonality_detected = self.detect_trend(time_series)

            # Detect anomalies
            anomalies = self.detect_anomalies(time_series)

            # Predict next month
            prediction = self.predict_next_month(time_series, trend_score)

            results.append({
                'admin_level': admin_level,
                'admin_name': admin_name,
                'trend_type': trend_type,
                'trend_score': trend_score,
                'seasonality_detected': 1 if seasonality_detected else 0,
                'anomalies_json': json.dumps(anomalies, ensure_ascii=False),
                'prediction_next_month': prediction
            })

            if (i + 1) % 10 == 0:
                progress = 0.2 + 0.6 * (i + 1) / total
                self.update_progress(progress, f"Processed {i + 1}/{total} regions")

        self.update_progress(0.8, f"Analyzed {len(results)} regions")
        return results

    def save_results(self, results):
        cursor = self.conn.cursor()
        cursor.execute("DELETE FROM admin_trends")

        for result in results:
            cursor.execute("""
                INSERT INTO admin_trends (
                    admin_level, admin_name, trend_type, trend_score,
                    seasonality_detected, anomalies_json, prediction_next_month, algo_version
                ) VALUES (?, ?, ?, ?, ?, ?, ?, 'v1')
            """, (
                result['admin_level'], result['admin_name'], result['trend_type'],
                result['trend_score'], result['seasonality_detected'],
                result['anomalies_json'], result['prediction_next_month']
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
            admin_stats = self.load_admin_stats()
            self.update_progress(0.1, f"Loaded {len(admin_stats)} admin regions")

            results = self.process(admin_stats)
            self.save_results(results)

            # Calculate summary
            trend_counts = {}
            for r in results:
                trend_type = r['trend_type']
                trend_counts[trend_type] = trend_counts.get(trend_type, 0) + 1

            seasonality_count = sum(1 for r in results if r['seasonality_detected'])

            summary = {
                'total_regions': len(results),
                'trend_distribution': trend_counts,
                'seasonality_detected_count': seasonality_count,
                'avg_trend_score': np.mean([r['trend_score'] for r in results]) if results else 0
            }

            self.mark_completed(summary)
            print(f"Admin trend analysis completed: {len(results)} regions analyzed")
            return 0

        except Exception as e:
            error_msg = f"Admin trend analysis failed: {str(e)}"
            print(error_msg, file=sys.stderr)
            self.mark_failed(error_msg)
            return 1

        finally:
            self.conn.close()

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: python admin_view_advanced.py <db_path> <task_id>")
        sys.exit(1)

    db_path = sys.argv[1]
    task_id = int(sys.argv[2])

    worker = AdminViewAdvancedWorker(db_path, task_id)
    sys.exit(worker.run())
