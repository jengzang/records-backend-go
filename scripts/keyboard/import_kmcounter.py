#!/usr/bin/env python3
"""
KMCounter Data Importer
Imports keyboard/mouse usage data from KMCounter.ini to SQLite database
"""

import configparser
import sqlite3
import re
import sys
import os
from datetime import datetime
from pathlib import Path


class KMCounterImporter:
    def __init__(self, ini_path, db_path):
        self.ini_path = ini_path
        self.db_path = db_path
        self.conn = None
        self.stats = {
            'total_days': 0,
            'total_keystrokes': 0,
            'total_clicks': 0,
            'total_distance': 0.0,
            'errors': []
        }

    def connect_db(self):
        """Connect to SQLite database and enable WAL mode"""
        self.conn = sqlite3.connect(self.db_path)
        self.conn.execute('PRAGMA journal_mode=WAL')
        self.conn.execute('PRAGMA synchronous=NORMAL')
        print(f"Connected to database: {self.db_path}")

    def read_ini_file(self):
        """Read KMCounter.ini with UTF-16 encoding"""
        print(f"Reading INI file: {self.ini_path}")

        try:
            with open(self.ini_path, 'rb') as f:
                content = f.read().decode('utf-16')

            config = configparser.ConfigParser()
            config.read_string(content)

            print(f"Successfully parsed INI file with {len(config.sections())} sections")
            return config

        except Exception as e:
            print(f"Error reading INI file: {e}")
            sys.exit(1)

    def import_date_section(self, date, data):
        """Import data for a single date section"""
        try:
            # Extract daily stats
            keystrokes = int(data.get('keystrokes', 0))
            left_clicks = int(data.get('lbcount', 0))
            right_clicks = int(data.get('rbcount', 0))
            middle_clicks = int(data.get('mbcount', 0))
            extra_clicks = int(data.get('xbcount', 0))
            wheel_scrolls = int(data.get('wheel', 0))
            h_wheel_scrolls = int(data.get('hwheel', 0))
            mouse_distance_m = float(data.get('move', 0.0))

            # Insert daily stats
            self.conn.execute('''
                INSERT OR REPLACE INTO daily_stats
                (date, keystrokes, left_clicks, right_clicks, middle_clicks,
                 extra_clicks, wheel_scrolls, h_wheel_scrolls, mouse_distance_m)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
            ''', (date, keystrokes, left_clicks, right_clicks, middle_clicks,
                  extra_clicks, wheel_scrolls, h_wheel_scrolls, mouse_distance_m))

            # Update statistics
            self.stats['total_days'] += 1
            self.stats['total_keystrokes'] += keystrokes
            self.stats['total_clicks'] += (left_clicks + right_clicks + middle_clicks + extra_clicks)
            self.stats['total_distance'] += mouse_distance_m

            # Extract scancode stats
            scancode_count = 0
            for key, value in data.items():
                if key.startswith('sc'):
                    try:
                        scancode = int(key[2:])
                        count = int(value)
                        self.conn.execute('''
                            INSERT OR REPLACE INTO scancode_stats (date, scancode, count)
                            VALUES (?, ?, ?)
                        ''', (date, scancode, count))
                        scancode_count += 1
                    except ValueError:
                        continue

            if self.stats['total_days'] % 100 == 0:
                print(f"Processed {self.stats['total_days']} days...")

        except Exception as e:
            error_msg = f"Error importing date {date}: {e}"
            self.stats['errors'].append(error_msg)
            print(error_msg)

    def import_data(self):
        """Main import process"""
        config = self.read_ini_file()

        print("Starting data import...")
        date_sections = []

        # Find all date sections (YYYYMMDD format)
        for section in config.sections():
            if re.match(r'^\d{8}$', section):
                date_sections.append(section)

        print(f"Found {len(date_sections)} date sections to import")

        # Sort dates chronologically
        date_sections.sort()

        # Import each date section
        for date in date_sections:
            self.import_date_section(date, config[date])

        # Commit all changes
        self.conn.commit()
        print("Data import completed!")

    def print_summary(self):
        """Print import summary statistics"""
        print("\n" + "="*60)
        print("IMPORT SUMMARY")
        print("="*60)
        print(f"Total days imported: {self.stats['total_days']}")
        print(f"Total keystrokes: {self.stats['total_keystrokes']:,}")
        print(f"Total clicks: {self.stats['total_clicks']:,}")
        print(f"Total mouse distance: {self.stats['total_distance']:.2f} meters")
        print(f"Total mouse distance: {self.stats['total_distance']/1000:.2f} km")

        if self.stats['errors']:
            print(f"\nErrors encountered: {len(self.stats['errors'])}")
            for error in self.stats['errors'][:5]:
                print(f"  - {error}")
            if len(self.stats['errors']) > 5:
                print(f"  ... and {len(self.stats['errors']) - 5} more")

        print("="*60)

    def close(self):
        """Close database connection"""
        if self.conn:
            self.conn.close()
            print("Database connection closed")

    def run(self):
        """Run the complete import process"""
        try:
            self.connect_db()
            self.import_data()
            self.print_summary()
        finally:
            self.close()


def main():
    """Main entry point"""
    if len(sys.argv) < 3:
        print("Usage: python import_kmcounter.py <ini_path> <db_path>")
        print("Example: python import_kmcounter.py KMCounter.ini kmcounter.db")
        sys.exit(1)

    ini_path = sys.argv[1]
    db_path = sys.argv[2]

    # Validate paths
    if not os.path.exists(ini_path):
        print(f"Error: INI file not found: {ini_path}")
        sys.exit(1)

    # Create database directory if needed
    db_dir = os.path.dirname(db_path)
    if db_dir and not os.path.exists(db_dir):
        os.makedirs(db_dir)
        print(f"Created database directory: {db_dir}")

    # Run importer
    importer = KMCounterImporter(ini_path, db_path)
    importer.run()


if __name__ == '__main__':
    main()
