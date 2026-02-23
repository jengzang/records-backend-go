#!/usr/bin/env python3
"""
Typing Behavior Analysis Worker for Keyboard Data

Analyzes typing behavior patterns:
- Backspace ratio (correction frequency)
- Enter frequency (line breaks)
- Space frequency (word count estimate)
- Common key combinations
"""

import sqlite3
import json
import sys
from typing import Dict, List, Any


class TypingBehaviorAnalyzer:
    def __init__(self, db_path: str):
        self.db_path = db_path
        self.conn = sqlite3.connect(db_path)
        self.conn.row_factory = sqlite3.Row

    def analyze_typing_metrics(self, start_date: str = None, end_date: str = None) -> Dict[str, Any]:
        """Calculate comprehensive typing behavior metrics"""
        where_clause = ""
        params = []

        if start_date and end_date:
            where_clause = "WHERE date >= ? AND date <= ?"
            params = [start_date, end_date]

        # Get total keystrokes
        query_total = f"""
            SELECT SUM(keystrokes) as total_keystrokes
            FROM daily_stats
            {where_clause}
        """
        cursor = self.conn.execute(query_total, params)
        row = cursor.fetchone()
        total_keystrokes = row['total_keystrokes'] if row['total_keystrokes'] else 0

        # Get specific key counts
        where_clause_scancode = ""
        if start_date and end_date:
            where_clause_scancode = "AND s.date >= ? AND s.date <= ?"

        query_keys = f"""
            SELECT
                m.key_name,
                SUM(s.count) as total_count
            FROM scancode_stats s
            LEFT JOIN scancode_mapping m ON s.scancode = m.scancode
            WHERE m.key_name IN ('Backspace', 'Enter', 'Space', 'Delete')
            {where_clause_scancode}
            GROUP BY m.key_name
        """

        cursor = self.conn.execute(query_keys, params)

        key_counts = {
            'backspace': 0,
            'enter': 0,
            'space': 0,
            'delete': 0
        }

        for row in cursor:
            key_name = row['key_name'].lower()
            if key_name in key_counts:
                key_counts[key_name] = row['total_count']

        # Calculate metrics
        backspace_ratio = round(key_counts['backspace'] / total_keystrokes, 4) if total_keystrokes > 0 else 0
        delete_ratio = round(key_counts['delete'] / total_keystrokes, 4) if total_keystrokes > 0 else 0
        correction_ratio = round((key_counts['backspace'] + key_counts['delete']) / total_keystrokes, 4) if total_keystrokes > 0 else 0

        # Estimate words (space count + 1)
        estimated_words = key_counts['space'] + 1 if key_counts['space'] > 0 else 0

        # Estimate lines (enter count + 1)
        estimated_lines = key_counts['enter'] + 1 if key_counts['enter'] > 0 else 0

        results = {
            'total_keystrokes': total_keystrokes,
            'backspace_count': key_counts['backspace'],
            'enter_count': key_counts['enter'],
            'space_count': key_counts['space'],
            'delete_count': key_counts['delete'],
            'backspace_ratio': backspace_ratio,
            'delete_ratio': delete_ratio,
            'correction_ratio': correction_ratio,
            'estimated_words': estimated_words,
            'estimated_lines': estimated_lines,
            'avg_word_length': round((total_keystrokes - key_counts['space']) / estimated_words, 2) if estimated_words > 0 else 0
        }

        return results

    def analyze_special_key_usage(self, start_date: str = None, end_date: str = None) -> List[Dict[str, Any]]:
        """Analyze usage of special keys"""
        where_clause = ""
        params = []

        if start_date and end_date:
            where_clause = "AND s.date >= ? AND s.date <= ?"
            params = [start_date, end_date]

        query = f"""
            SELECT
                m.key_name,
                SUM(s.count) as total_count
            FROM scancode_stats s
            LEFT JOIN scancode_mapping m ON s.scancode = m.scancode
            WHERE m.key_category = 'special'
            {where_clause}
            GROUP BY m.key_name
            ORDER BY total_count DESC
            LIMIT 20
        """

        cursor = self.conn.execute(query, params)

        results = []
        for row in cursor:
            results.append({
                'key_name': row['key_name'],
                'count': row['total_count']
            })

        return results

    def analyze_letter_frequency(self, start_date: str = None, end_date: str = None) -> List[Dict[str, Any]]:
        """Analyze letter frequency distribution"""
        where_clause = ""
        params = []

        if start_date and end_date:
            where_clause = "AND s.date >= ? AND s.date <= ?"
            params = [start_date, end_date]

        query = f"""
            SELECT
                m.key_name,
                SUM(s.count) as total_count
            FROM scancode_stats s
            LEFT JOIN scancode_mapping m ON s.scancode = m.scancode
            WHERE m.key_category = 'letter'
            {where_clause}
            GROUP BY m.key_name
            ORDER BY total_count DESC
        """

        cursor = self.conn.execute(query, params)

        # Calculate total for percentage
        results_raw = []
        total_letters = 0
        for row in cursor:
            count = row['total_count']
            total_letters += count
            results_raw.append({
                'letter': row['key_name'],
                'count': count
            })

        # Add percentages
        results = []
        for item in results_raw:
            results.append({
                'letter': item['letter'],
                'count': item['count'],
                'percentage': round((item['count'] / total_letters * 100), 2) if total_letters > 0 else 0
            })

        return results

    def close(self):
        """Close database connection"""
        self.conn.close()


def main():
    """Main entry point for CLI usage"""
    if len(sys.argv) < 3:
        print("Usage: python typing_behavior.py <db_path> <analysis_type> [start_date] [end_date]")
        print("Analysis types: metrics, special_keys, letter_frequency")
        sys.exit(1)

    db_path = sys.argv[1]
    analysis_type = sys.argv[2]
    start_date = sys.argv[3] if len(sys.argv) > 3 else None
    end_date = sys.argv[4] if len(sys.argv) > 4 else None

    analyzer = TypingBehaviorAnalyzer(db_path)

    try:
        if analysis_type == 'metrics':
            results = analyzer.analyze_typing_metrics(start_date, end_date)
        elif analysis_type == 'special_keys':
            results = analyzer.analyze_special_key_usage(start_date, end_date)
        elif analysis_type == 'letter_frequency':
            results = analyzer.analyze_letter_frequency(start_date, end_date)
        else:
            print(f"Unknown analysis type: {analysis_type}")
            sys.exit(1)

        print(json.dumps(results, indent=2))
    finally:
        analyzer.close()


if __name__ == '__main__':
    main()
