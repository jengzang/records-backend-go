#!/usr/bin/env python3
"""
Productivity Metrics Analysis Worker for Keyboard/Mouse Data

Analyzes productivity patterns:
- Active days and streaks
- Typing intensity
- Consistency score
- Focus time estimation
"""

import sqlite3
import json
import sys
from datetime import datetime, timedelta
from typing import Dict, List, Any


class ProductivityAnalyzer:
    def __init__(self, db_path: str):
        self.db_path = db_path
        self.conn = sqlite3.connect(db_path)
        self.conn.row_factory = sqlite3.Row

    def analyze_activity_metrics(self, threshold: int = 100, start_date: str = None, end_date: str = None) -> Dict[str, Any]:
        """Calculate activity metrics (active days, streaks, etc.)"""
        where_clause = ""
        params = []

        if start_date and end_date:
            where_clause = "WHERE date >= ? AND date <= ?"
            params = [start_date, end_date]

        query = f"""
            SELECT
                date,
                keystrokes,
                left_clicks + right_clicks + middle_clicks + extra_clicks as total_clicks
            FROM daily_stats
            {where_clause}
            ORDER BY date
        """

        cursor = self.conn.execute(query, params)

        active_days = 0
        total_days = 0
        current_streak = 0
        longest_streak = 0
        temp_streak = 0
        prev_date = None

        daily_keystrokes = []

        for row in cursor:
            total_days += 1
            keystrokes = row['keystrokes']
            daily_keystrokes.append(keystrokes)

            date_obj = datetime.strptime(row['date'], '%Y%m%d')

            if keystrokes >= threshold:
                active_days += 1
                temp_streak += 1

                # Check if consecutive
                if prev_date:
                    expected_date = prev_date + timedelta(days=1)
                    if date_obj != expected_date:
                        # Streak broken
                        if temp_streak > longest_streak:
                            longest_streak = temp_streak
                        temp_streak = 1

                prev_date = date_obj
            else:
                # Inactive day, reset streak
                if temp_streak > longest_streak:
                    longest_streak = temp_streak
                temp_streak = 0
                prev_date = None

        # Check final streak
        if temp_streak > longest_streak:
            longest_streak = temp_streak

        # Current streak is the temp_streak if it extends to the last day
        current_streak = temp_streak if prev_date else 0

        # Calculate consistency (standard deviation)
        if len(daily_keystrokes) > 0:
            mean_keystrokes = sum(daily_keystrokes) / len(daily_keystrokes)
            variance = sum((x - mean_keystrokes) ** 2 for x in daily_keystrokes) / len(daily_keystrokes)
            std_dev = variance ** 0.5
            consistency_score = round(std_dev / mean_keystrokes, 4) if mean_keystrokes > 0 else 0
        else:
            mean_keystrokes = 0
            std_dev = 0
            consistency_score = 0

        results = {
            'total_days': total_days,
            'active_days': active_days,
            'inactive_days': total_days - active_days,
            'activity_rate': round(active_days / total_days, 4) if total_days > 0 else 0,
            'current_streak': current_streak,
            'longest_streak': longest_streak,
            'avg_keystrokes_per_day': round(mean_keystrokes, 2),
            'std_dev_keystrokes': round(std_dev, 2),
            'consistency_score': consistency_score,
            'threshold_used': threshold
        }

        return results

    def analyze_typing_intensity(self, start_date: str = None, end_date: str = None) -> Dict[str, Any]:
        """Calculate typing intensity metrics"""
        where_clause = ""
        params = []

        if start_date and end_date:
            where_clause = "WHERE date >= ? AND date <= ?"
            params = [start_date, end_date]

        query = f"""
            SELECT
                keystrokes,
                left_clicks + right_clicks + middle_clicks + extra_clicks as total_clicks,
                mouse_distance_m
            FROM daily_stats
            {where_clause}
            ORDER BY keystrokes DESC
        """

        cursor = self.conn.execute(query, params)

        keystrokes_list = []
        clicks_list = []
        distance_list = []

        for row in cursor:
            if row['keystrokes'] > 0:  # Only count active days
                keystrokes_list.append(row['keystrokes'])
                clicks_list.append(row['total_clicks'])
                distance_list.append(row['mouse_distance_m'])

        if len(keystrokes_list) == 0:
            return {
                'avg_keystrokes': 0,
                'avg_clicks': 0,
                'avg_distance': 0,
                'peak_keystrokes': 0,
                'peak_clicks': 0,
                'peak_distance': 0
            }

        # Calculate percentiles
        keystrokes_sorted = sorted(keystrokes_list)
        clicks_sorted = sorted(clicks_list)
        distance_sorted = sorted(distance_list)

        def percentile(data, p):
            index = int(len(data) * p / 100)
            return data[min(index, len(data) - 1)]

        results = {
            'avg_keystrokes': round(sum(keystrokes_list) / len(keystrokes_list), 2),
            'avg_clicks': round(sum(clicks_list) / len(clicks_list), 2),
            'avg_distance': round(sum(distance_list) / len(distance_list), 2),
            'peak_keystrokes': max(keystrokes_list),
            'peak_clicks': max(clicks_list),
            'peak_distance': round(max(distance_list), 2),
            'p50_keystrokes': percentile(keystrokes_sorted, 50),
            'p75_keystrokes': percentile(keystrokes_sorted, 75),
            'p95_keystrokes': percentile(keystrokes_sorted, 95),
            'active_days': len(keystrokes_list)
        }

        return results

    def analyze_peak_days(self, limit: int = 10, start_date: str = None, end_date: str = None) -> List[Dict[str, Any]]:
        """Find peak usage days"""
        where_clause = ""
        params = []

        if start_date and end_date:
            where_clause = "WHERE date >= ? AND date <= ?"
            params = [start_date, end_date]

        query = f"""
            SELECT
                date,
                keystrokes,
                left_clicks + right_clicks + middle_clicks + extra_clicks as total_clicks,
                mouse_distance_m
            FROM daily_stats
            {where_clause}
            ORDER BY keystrokes DESC
            LIMIT ?
        """

        params.append(limit)
        cursor = self.conn.execute(query, params)

        results = []
        for row in cursor:
            results.append({
                'date': row['date'],
                'keystrokes': row['keystrokes'],
                'clicks': row['total_clicks'],
                'distance': round(row['mouse_distance_m'], 2)
            })

        return results

    def close(self):
        """Close database connection"""
        self.conn.close()


def main():
    """Main entry point for CLI usage"""
    if len(sys.argv) < 3:
        print("Usage: python productivity_metrics.py <db_path> <analysis_type> [start_date] [end_date]")
        print("Analysis types: activity, intensity, peak_days")
        sys.exit(1)

    db_path = sys.argv[1]
    analysis_type = sys.argv[2]
    start_date = sys.argv[3] if len(sys.argv) > 3 else None
    end_date = sys.argv[4] if len(sys.argv) > 4 else None

    analyzer = ProductivityAnalyzer(db_path)

    try:
        if analysis_type == 'activity':
            results = analyzer.analyze_activity_metrics(100, start_date, end_date)
        elif analysis_type == 'intensity':
            results = analyzer.analyze_typing_intensity(start_date, end_date)
        elif analysis_type == 'peak_days':
            results = analyzer.analyze_peak_days(10, start_date, end_date)
        else:
            print(f"Unknown analysis type: {analysis_type}")
            sys.exit(1)

        print(json.dumps(results, indent=2))
    finally:
        analyzer.close()


if __name__ == '__main__':
    main()
