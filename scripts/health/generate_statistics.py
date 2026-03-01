#!/usr/bin/env python3
"""
Generate health statistics from raw health records.
This script processes health_records table and generates aggregated statistics.
"""

import sqlite3
import sys
from datetime import datetime
from pathlib import Path

def generate_statistics(db_path):
    """Generate health statistics"""
    print(f"Opening database: {db_path}")
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()

    try:
        # Get date range
        print("\n=== Checking data range ===")
        cursor.execute("""
            SELECT MIN(start_date), MAX(start_date), COUNT(*)
            FROM health_records
        """)
        min_date, max_date, total_records = cursor.fetchone()
        print(f"Date range: {min_date} to {max_date}")
        print(f"Total records: {total_records:,}")

        # Get available metric types
        print("\n=== Available metric types ===")
        cursor.execute("""
            SELECT type, COUNT(*) as count
            FROM health_records
            GROUP BY type
            ORDER BY count DESC
            LIMIT 10
        """)
        metric_types = cursor.fetchall()
        for metric_type, count in metric_types:
            print(f"  {metric_type}: {count:,} records")

        # Clear old statistics
        print("\n=== Clearing old statistics ===")
        cursor.execute("DELETE FROM health_statistics")
        deleted = cursor.rowcount
        print(f"Deleted {deleted} old statistics")

        # Generate statistics for each metric type
        print("\n=== Generating new statistics ===")

        for metric_type, _ in metric_types:
            print(f"\nProcessing {metric_type}...")

            # Daily statistics
            cursor.execute("""
                INSERT OR REPLACE INTO health_statistics
                (stat_type, stat_date, metric_type, total_value, avg_value, min_value, max_value, count, created_at)
                SELECT
                    'daily' as stat_type,
                    DATE(start_date) as stat_date,
                    ? as metric_type,
                    SUM(value) as total_value,
                    AVG(value) as avg_value,
                    MIN(value) as min_value,
                    MAX(value) as max_value,
                    COUNT(*) as count,
                    CURRENT_TIMESTAMP as created_at
                FROM health_records
                WHERE type = ?
                GROUP BY DATE(start_date)
            """, (metric_type, metric_type))
            daily_count = cursor.rowcount
            print(f"  [OK] Generated {daily_count} daily statistics")

            # Weekly statistics
            cursor.execute("""
                INSERT OR REPLACE INTO health_statistics
                (stat_type, stat_date, metric_type, total_value, avg_value, min_value, max_value, count, created_at)
                SELECT
                    'weekly' as stat_type,
                    strftime('%Y-W%W', start_date) as stat_date,
                    ? as metric_type,
                    SUM(value) as total_value,
                    AVG(value) as avg_value,
                    MIN(value) as min_value,
                    MAX(value) as max_value,
                    COUNT(*) as count,
                    CURRENT_TIMESTAMP as created_at
                FROM health_records
                WHERE type = ?
                GROUP BY strftime('%Y-W%W', start_date)
            """, (metric_type, metric_type))
            weekly_count = cursor.rowcount
            print(f"  [OK] Generated {weekly_count} weekly statistics")

            # Monthly statistics
            cursor.execute("""
                INSERT OR REPLACE INTO health_statistics
                (stat_type, stat_date, metric_type, total_value, avg_value, min_value, max_value, count, created_at)
                SELECT
                    'monthly' as stat_type,
                    strftime('%Y-%m', start_date) as stat_date,
                    ? as metric_type,
                    SUM(value) as total_value,
                    AVG(value) as avg_value,
                    MIN(value) as min_value,
                    MAX(value) as max_value,
                    COUNT(*) as count,
                    CURRENT_TIMESTAMP as created_at
                FROM health_records
                WHERE type = ?
                GROUP BY strftime('%Y-%m', start_date)
            """, (metric_type, metric_type))
            monthly_count = cursor.rowcount
            print(f"  [OK] Generated {monthly_count} monthly statistics")

        # Commit changes
        conn.commit()

        # Get final summary
        print("\n=== Final Statistics Summary ===")
        cursor.execute("""
            SELECT stat_type, COUNT(*) as count
            FROM health_statistics
            GROUP BY stat_type
        """)
        for stat_type, count in cursor.fetchall():
            print(f"  {stat_type}: {count:,} records")

        cursor.execute("SELECT COUNT(*) FROM health_statistics")
        total_stats = cursor.fetchone()[0]
        print(f"\nTotal statistics generated: {total_stats:,}")

        print("\n[OK] Statistics generation completed successfully!")

    except Exception as e:
        print(f"\n[ERROR] Error: {e}", file=sys.stderr)
        conn.rollback()
        return 1
    finally:
        conn.close()

    return 0

if __name__ == "__main__":
    # Default database path
    db_path = Path(__file__).parent.parent.parent / "data" / "applehealth" / "health.db"

    # Allow custom path from command line
    if len(sys.argv) > 1:
        db_path = Path(sys.argv[1])

    if not db_path.exists():
        print(f"Error: Database not found at {db_path}", file=sys.stderr)
        sys.exit(1)

    sys.exit(generate_statistics(str(db_path)))
