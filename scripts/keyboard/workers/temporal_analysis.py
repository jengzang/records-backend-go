#!/usr/bin/env python3
"""
Temporal Analysis Worker for Keyboard/Mouse Data

Analyzes temporal patterns in keyboard/mouse usage:
- Hour-of-day distribution (if hourly data available)
- Day-of-week patterns
- Month-of-year seasonality
- Weekday vs weekend comparison
"""

import sqlite3
import json
import sys
from datetime import datetime
from typing import Dict, List, Any


class TemporalAnalyzer:
    def __init__(self, db_path: str):
        self.db_path = db_path
        self.conn = sqlite3.connect(db_path)
        self.conn.row_factory = sqlite3.Row

    def analyze_hourly_distribution(self, start_date: str = None, end_date: str = None) -> List[Dict[str, Any]]:
        """Analyze hour-of-day distribution (0-23) - requires hourly_stats table"""
        where_clause = ""
        params = []

        if start_date and end_date:
            where_clause = "WHERE date >= ? AND date <= ?"
            params = [start_date, end_date]

        query = f"""
            SELECT
                hour,
                SUM(keystrokes) as total_keystrokes,
                SUM(mouse_clicks) as total_clicks,
                SUM(mouse_distance_m) as total_distance,
                AVG(keystrokes) as avg_keystrokes,
                AVG(mouse_clicks) as avg_clicks,
                COUNT(*) as day_count
            FROM hourly_stats
            {where_clause}
            GROUP BY hour
            ORDER BY hour
        """

        cursor = self.conn.execute(query, params)
        results = []
        for row in cursor:
            results.append({
                'hour': row['hour'],
                'total_keystrokes': row['total_keystrokes'],
                'total_clicks': row['total_clicks'],
                'total_distance': row['total_distance'],
                'avg_keystrokes': round(row['avg_keystrokes'], 2),
                'avg_clicks': round(row['avg_clicks'], 2),
                'day_count': row['day_count']
            })

        return results

    def analyze_day_of_week(self, start_date: str = None, end_date: str = None) -> List[Dict[str, Any]]:
        """Analyze day-of-week patterns (0=Monday, 6=Sunday)"""
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
        """

        cursor = self.conn.execute(query, params)

        # Group by day of week
        day_stats = {i: {'keystrokes': 0, 'clicks': 0, 'distance': 0, 'count': 0} for i in range(7)}

        for row in cursor:
            date_str = row['date']
            date_obj = datetime.strptime(date_str, '%Y%m%d')
            day_of_week = date_obj.weekday()  # 0=Monday, 6=Sunday

            day_stats[day_of_week]['keystrokes'] += row['keystrokes']
            day_stats[day_of_week]['clicks'] += row['total_clicks']
            day_stats[day_of_week]['distance'] += row['mouse_distance_m']
            day_stats[day_of_week]['count'] += 1

        # Calculate averages and format results
        day_names = ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday']
        results = []

        for day in range(7):
            stats = day_stats[day]
            count = stats['count']
            if count > 0:
                results.append({
                    'day': day,
                    'day_name': day_names[day],
                    'is_weekend': day >= 5,
                    'total_keystrokes': stats['keystrokes'],
                    'total_clicks': stats['clicks'],
                    'total_distance': round(stats['distance'], 2),
                    'avg_keystrokes': round(stats['keystrokes'] / count, 2),
                    'avg_clicks': round(stats['clicks'] / count, 2),
                    'avg_distance': round(stats['distance'] / count, 2),
                    'day_count': count
                })

        return results

    def analyze_monthly_patterns(self, start_date: str = None, end_date: str = None) -> List[Dict[str, Any]]:
        """Analyze month-of-year seasonality"""
        where_clause = ""
        params = []

        if start_date and end_date:
            where_clause = "WHERE date >= ? AND date <= ?"
            params = [start_date, end_date]

        query = f"""
            SELECT
                SUBSTR(date, 1, 4) as year,
                SUBSTR(date, 5, 2) as month,
                SUM(keystrokes) as total_keystrokes,
                SUM(left_clicks + right_clicks + middle_clicks + extra_clicks) as total_clicks,
                SUM(mouse_distance_m) as total_distance,
                AVG(keystrokes) as avg_keystrokes,
                AVG(left_clicks + right_clicks + middle_clicks + extra_clicks) as avg_clicks,
                COUNT(*) as day_count
            FROM daily_stats
            {where_clause}
            GROUP BY year, month
            ORDER BY year, month
        """

        cursor = self.conn.execute(query, params)
        results = []

        month_names = ['', 'January', 'February', 'March', 'April', 'May', 'June',
                      'July', 'August', 'September', 'October', 'November', 'December']

        for row in cursor:
            month_num = int(row['month'])
            results.append({
                'year': int(row['year']),
                'month': month_num,
                'month_name': month_names[month_num],
                'total_keystrokes': row['total_keystrokes'],
                'total_clicks': row['total_clicks'],
                'total_distance': round(row['total_distance'], 2),
                'avg_keystrokes': round(row['avg_keystrokes'], 2),
                'avg_clicks': round(row['avg_clicks'], 2),
                'day_count': row['day_count']
            })

        return results

    def analyze_weekday_vs_weekend(self, start_date: str = None, end_date: str = None) -> Dict[str, Any]:
        """Compare weekday vs weekend usage"""
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
        """

        cursor = self.conn.execute(query, params)

        weekday_stats = {'keystrokes': 0, 'clicks': 0, 'distance': 0, 'count': 0}
        weekend_stats = {'keystrokes': 0, 'clicks': 0, 'distance': 0, 'count': 0}

        for row in cursor:
            date_str = row['date']
            date_obj = datetime.strptime(date_str, '%Y%m%d')
            day_of_week = date_obj.weekday()

            if day_of_week >= 5:  # Saturday or Sunday
                weekend_stats['keystrokes'] += row['keystrokes']
                weekend_stats['clicks'] += row['total_clicks']
                weekend_stats['distance'] += row['mouse_distance_m']
                weekend_stats['count'] += 1
            else:  # Monday to Friday
                weekday_stats['keystrokes'] += row['keystrokes']
                weekday_stats['clicks'] += row['total_clicks']
                weekday_stats['distance'] += row['mouse_distance_m']
                weekday_stats['count'] += 1

        result = {
            'weekday': {
                'total_keystrokes': weekday_stats['keystrokes'],
                'total_clicks': weekday_stats['clicks'],
                'total_distance': round(weekday_stats['distance'], 2),
                'avg_keystrokes': round(weekday_stats['keystrokes'] / weekday_stats['count'], 2) if weekday_stats['count'] > 0 else 0,
                'avg_clicks': round(weekday_stats['clicks'] / weekday_stats['count'], 2) if weekday_stats['count'] > 0 else 0,
                'day_count': weekday_stats['count']
            },
            'weekend': {
                'total_keystrokes': weekend_stats['keystrokes'],
                'total_clicks': weekend_stats['clicks'],
                'total_distance': round(weekend_stats['distance'], 2),
                'avg_keystrokes': round(weekend_stats['keystrokes'] / weekend_stats['count'], 2) if weekend_stats['count'] > 0 else 0,
                'avg_clicks': round(weekend_stats['clicks'] / weekend_stats['count'], 2) if weekend_stats['count'] > 0 else 0,
                'day_count': weekend_stats['count']
            }
        }

        # Calculate ratio
        if weekend_stats['keystrokes'] > 0:
            result['weekday_to_weekend_ratio'] = round(weekday_stats['keystrokes'] / weekend_stats['keystrokes'], 2)
        else:
            result['weekday_to_weekend_ratio'] = 0

        return result

    def close(self):
        """Close database connection"""
        self.conn.close()


def main():
    """Main entry point for CLI usage"""
    if len(sys.argv) < 3:
        print("Usage: python temporal_analysis.py <db_path> <analysis_type> [start_date] [end_date]")
        print("Analysis types: hourly, daily, monthly, weekday_vs_weekend")
        sys.exit(1)

    db_path = sys.argv[1]
    analysis_type = sys.argv[2]
    start_date = sys.argv[3] if len(sys.argv) > 3 else None
    end_date = sys.argv[4] if len(sys.argv) > 4 else None

    analyzer = TemporalAnalyzer(db_path)

    try:
        if analysis_type == 'hourly':
            results = analyzer.analyze_hourly_distribution(start_date, end_date)
        elif analysis_type == 'daily':
            results = analyzer.analyze_day_of_week(start_date, end_date)
        elif analysis_type == 'monthly':
            results = analyzer.analyze_monthly_patterns(start_date, end_date)
        elif analysis_type == 'weekday_vs_weekend':
            results = analyzer.analyze_weekday_vs_weekend(start_date, end_date)
        else:
            print(f"Unknown analysis type: {analysis_type}")
            sys.exit(1)

        print(json.dumps(results, indent=2))
    finally:
        analyzer.close()


if __name__ == '__main__':
    main()
