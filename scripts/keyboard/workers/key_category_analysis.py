#!/usr/bin/env python3
"""
Key Category Analysis Worker for Keyboard Data

Analyzes key usage by category:
- Category distribution (letters, numbers, function, modifiers, special)
- Modifier key usage patterns
- Top keys by category
"""

import sqlite3
import json
import sys
from typing import Dict, List, Any


class KeyCategoryAnalyzer:
    def __init__(self, db_path: str):
        self.db_path = db_path
        self.conn = sqlite3.connect(db_path)
        self.conn.row_factory = sqlite3.Row

    def analyze_category_distribution(self, start_date: str = None, end_date: str = None) -> List[Dict[str, Any]]:
        """Analyze key usage by category"""
        where_clause = ""
        params = []

        if start_date and end_date:
            where_clause = "AND s.date >= ? AND s.date <= ?"
            params = [start_date, end_date]

        query = f"""
            SELECT
                COALESCE(m.key_category, 'unknown') as category,
                SUM(s.count) as total_count,
                COUNT(DISTINCT s.scancode) as unique_keys
            FROM scancode_stats s
            LEFT JOIN scancode_mapping m ON s.scancode = m.scancode
            WHERE 1=1 {where_clause}
            GROUP BY m.key_category
            ORDER BY total_count DESC
        """

        cursor = self.conn.execute(query, params)

        # Calculate total for percentage
        total_keystrokes = 0
        results_raw = []
        for row in cursor:
            count = row['total_count']
            total_keystrokes += count
            results_raw.append({
                'category': row['category'],
                'count': count,
                'unique_keys': row['unique_keys']
            })

        # Add percentages
        results = []
        for item in results_raw:
            results.append({
                'category': item['category'],
                'count': item['count'],
                'percentage': round((item['count'] / total_keystrokes * 100), 2) if total_keystrokes > 0 else 0,
                'unique_keys': item['unique_keys']
            })

        return results

    def analyze_top_keys_by_category(self, category: str, limit: int = 5,
                                    start_date: str = None, end_date: str = None) -> List[Dict[str, Any]]:
        """Get top N keys for a specific category"""
        where_clause = "WHERE m.key_category = ?"
        params = [category]

        if start_date and end_date:
            where_clause += " AND s.date >= ? AND s.date <= ?"
            params.extend([start_date, end_date])

        query = f"""
            SELECT
                s.scancode,
                m.key_name,
                m.key_category,
                SUM(s.count) as total_count
            FROM scancode_stats s
            LEFT JOIN scancode_mapping m ON s.scancode = m.scancode
            {where_clause}
            GROUP BY s.scancode, m.key_name, m.key_category
            ORDER BY total_count DESC
            LIMIT ?
        """

        params.append(limit)
        cursor = self.conn.execute(query, params)

        results = []
        for row in cursor:
            results.append({
                'scancode': row['scancode'],
                'key_name': row['key_name'],
                'category': row['key_category'],
                'count': row['total_count']
            })

        return results

    def analyze_all_top_keys(self, limit: int = 5, start_date: str = None, end_date: str = None) -> Dict[str, List[Dict[str, Any]]]:
        """Get top keys for all categories"""
        categories = ['letter', 'number', 'function', 'modifier', 'special']
        results = {}

        for category in categories:
            results[category] = self.analyze_top_keys_by_category(category, limit, start_date, end_date)

        return results

    def analyze_modifier_usage(self, start_date: str = None, end_date: str = None) -> Dict[str, Any]:
        """Analyze modifier key usage patterns"""
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
            WHERE m.key_category = 'modifier' {where_clause}
            GROUP BY m.key_name
            ORDER BY total_count DESC
        """

        cursor = self.conn.execute(query, params)

        results = {
            'ctrl': 0,
            'shift': 0,
            'alt': 0,
            'win': 0
        }

        for row in cursor:
            key_name = row['key_name'].lower()
            count = row['total_count']

            if 'ctrl' in key_name:
                results['ctrl'] += count
            elif 'shift' in key_name:
                results['shift'] += count
            elif 'alt' in key_name:
                results['alt'] += count
            elif 'win' in key_name:
                results['win'] += count

        return results

    def close(self):
        """Close database connection"""
        self.conn.close()


def main():
    """Main entry point for CLI usage"""
    if len(sys.argv) < 3:
        print("Usage: python key_category_analysis.py <db_path> <analysis_type> [start_date] [end_date]")
        print("Analysis types: distribution, top_keys, modifiers")
        sys.exit(1)

    db_path = sys.argv[1]
    analysis_type = sys.argv[2]
    start_date = sys.argv[3] if len(sys.argv) > 3 else None
    end_date = sys.argv[4] if len(sys.argv) > 4 else None

    analyzer = KeyCategoryAnalyzer(db_path)

    try:
        if analysis_type == 'distribution':
            results = analyzer.analyze_category_distribution(start_date, end_date)
        elif analysis_type == 'top_keys':
            results = analyzer.analyze_all_top_keys(5, start_date, end_date)
        elif analysis_type == 'modifiers':
            results = analyzer.analyze_modifier_usage(start_date, end_date)
        else:
            print(f"Unknown analysis type: {analysis_type}")
            sys.exit(1)

        print(json.dumps(results, indent=2))
    finally:
        analyzer.close()


if __name__ == '__main__':
    main()
