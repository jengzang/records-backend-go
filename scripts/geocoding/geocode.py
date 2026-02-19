#!/usr/bin/env python3
"""
Geocoding service for GPS track points.
Performs reverse geocoding using shapefile-based point-in-polygon queries.

Usage:
    python geocode.py [--batch-size 1000] [--limit 0]

Options:
    --batch-size: Number of points to process per batch (default: 1000)
    --limit: Maximum number of points to geocode (0 = all, default: 0)
"""

import sqlite3
import sys
import time
import argparse
from pathlib import Path
from typing import List, Tuple, Optional

try:
    import geopandas as gpd
    from shapely.geometry import Point
    import pandas as pd
except ImportError:
    print("Error: Required packages not installed.")
    print("Please install: pip install geopandas shapely pyproj pandas")
    sys.exit(1)


class GeocodingService:
    """Geocoding service using shapefile-based reverse geocoding."""

    def __init__(self, shapefile_dir: Path, db_path: Path):
        """
        Initialize geocoding service.

        Args:
            shapefile_dir: Directory containing shapefile data
            db_path: Path to SQLite database
        """
        self.shapefile_dir = shapefile_dir
        self.db_path = db_path
        self.gdf = None  # Single GeoDataFrame with all admin levels

        # Load shapefile
        self._load_shapefiles()

    def _load_shapefiles(self):
        """Load administrative boundary shapefile (single file with all levels)."""
        print("Loading shapefile...")

        # Load the single shapefile containing all administrative levels
        shapefile_path = self.shapefile_dir / "2024全国乡镇边界" / "2024全国乡镇边界.shp"

        if not shapefile_path.exists():
            print(f"Error: Shapefile not found: {shapefile_path}")
            print("Please ensure the shapefile is in the correct location.")
            sys.exit(1)

        try:
            self.gdf = gpd.read_file(shapefile_path)
            # Create spatial index for faster queries
            self.gdf.sindex
            print(f"  Loaded shapefile: {len(self.gdf)} features (乡镇级)")
            print(f"  Columns: {list(self.gdf.columns)}")
        except Exception as e:
            print(f"Error: Failed to load shapefile: {e}")
            sys.exit(1)

    def geocode_point(self, longitude: float, latitude: float) -> dict:
        """
        Reverse geocode a single point.

        Args:
            longitude: Longitude coordinate
            latitude: Latitude coordinate

        Returns:
            Dictionary with administrative divisions
        """
        point = Point(longitude, latitude)
        result = {
            'province': None,
            'city': None,
            'county': None,
            'town': None,
            'village': None,  # Not available in this dataset
        }

        try:
            # Use spatial index for faster queries
            possible_matches_index = list(self.gdf.sindex.intersection(point.bounds))
            possible_matches = self.gdf.iloc[possible_matches_index]

            # Check which polygon contains the point
            matches = possible_matches[possible_matches.contains(point)]

            if not matches.empty:
                # Get the first match (should be unique at town level)
                match = matches.iloc[0]

                # Extract all administrative levels from the single feature
                # Column names have encoding issues, so we access by column position
                columns = list(self.gdf.columns)
                result['province'] = match[columns[1]]  # 省级
                result['city'] = match[columns[2]]      # 市级
                result['county'] = match[columns[4]]    # 区县级
                result['town'] = match[columns[6]]      # 乡镇级

        except Exception as e:
            print(f"    Warning: Error geocoding point ({longitude}, {latitude}): {e}")

        return result

    def get_ungeocoded_points(self, limit: int = 0) -> List[Tuple]:
        """
        Get track points without administrative divisions.

        Args:
            limit: Maximum number of points to retrieve (0 = all)

        Returns:
            List of tuples (id, longitude, latitude)
        """
        conn = sqlite3.connect(str(self.db_path))
        cursor = conn.cursor()

        query = '''
            SELECT id, longitude, latitude
            FROM "一生足迹"
            WHERE province IS NULL OR province = ''
            ORDER BY dataTime ASC
        '''

        if limit > 0:
            query += f' LIMIT {limit}'

        cursor.execute(query)
        points = cursor.fetchall()
        conn.close()

        return points

    def update_admin_divisions(self, updates: List[Tuple]):
        """
        Update administrative divisions in database.

        Args:
            updates: List of tuples (id, province, city, county, town, village)
        """
        conn = sqlite3.connect(str(self.db_path))
        cursor = conn.cursor()

        # Enable WAL mode
        cursor.execute("PRAGMA journal_mode=WAL")

        # Prepare update statement
        update_sql = '''
            UPDATE "一生足迹"
            SET province = ?, city = ?, county = ?, town = ?, village = ?,
                updated_at = datetime('now'), algo_version = '1.0'
            WHERE id = ?
        '''

        # Execute batch update
        cursor.executemany(update_sql, [
            (u[1], u[2], u[3], u[4], u[5], u[0]) for u in updates
        ])

        conn.commit()
        conn.close()

    def run(self, batch_size: int = 1000, limit: int = 0):
        """
        Run geocoding process.

        Args:
            batch_size: Number of points to process per batch
            limit: Maximum number of points to geocode (0 = all)
        """
        print(f"\nStarting geocoding process...")
        print(f"Database: {self.db_path}")
        print(f"Batch size: {batch_size}")
        print(f"Limit: {limit if limit > 0 else 'all'}")

        # Get ungeocoded points
        print("\nFetching ungeocoded points...")
        points = self.get_ungeocoded_points(limit)
        total_points = len(points)

        if total_points == 0:
            print("No ungeocoded points found.")
            return

        print(f"Found {total_points} ungeocoded points")

        # Process in batches
        start_time = time.time()
        processed = 0
        failed = 0

        for i in range(0, total_points, batch_size):
            batch = points[i:i + batch_size]
            batch_updates = []

            print(f"\nProcessing batch {i // batch_size + 1}/{(total_points + batch_size - 1) // batch_size}...")

            for point_id, longitude, latitude in batch:
                try:
                    # Geocode point
                    result = self.geocode_point(longitude, latitude)

                    # Add to batch updates
                    batch_updates.append((
                        point_id,
                        result['province'],
                        result['city'],
                        result['county'],
                        result['town'],
                        result['village'],
                    ))

                    processed += 1

                    # Progress indicator
                    if processed % 100 == 0:
                        elapsed = time.time() - start_time
                        rate = processed / elapsed
                        remaining = (total_points - processed) / rate if rate > 0 else 0
                        print(f"  Progress: {processed}/{total_points} ({processed * 100 / total_points:.1f}%) "
                              f"- {rate:.1f} points/sec - ETA: {remaining:.0f}s")

                except Exception as e:
                    print(f"  Error geocoding point {point_id}: {e}")
                    failed += 1

            # Update database
            if batch_updates:
                try:
                    self.update_admin_divisions(batch_updates)
                    print(f"  Updated {len(batch_updates)} points in database")
                except Exception as e:
                    print(f"  Error updating database: {e}")

        # Summary
        elapsed = time.time() - start_time
        print(f"\nGeocoding completed!")
        print(f"  Total processed: {processed}")
        print(f"  Failed: {failed}")
        print(f"  Time elapsed: {elapsed:.1f}s")
        print(f"  Average rate: {processed / elapsed:.1f} points/sec")


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(description='Geocode GPS track points')
    parser.add_argument('--batch-size', type=int, default=1000,
                        help='Number of points to process per batch (default: 1000)')
    parser.add_argument('--limit', type=int, default=0,
                        help='Maximum number of points to geocode (0 = all, default: 0)')

    args = parser.parse_args()

    # Paths
    script_dir = Path(__file__).parent
    shapefile_dir = script_dir.parent.parent / "data" / "geo"
    db_path = script_dir.parent.parent / "data" / "tracks" / "tracks.db"

    # Verify paths
    if not shapefile_dir.exists():
        print(f"Error: Shapefile directory not found: {shapefile_dir}")
        print("Please create the directory and add the shapefile:")
        print("  Expected: {shapefile_dir}/2024全国乡镇边界/2024全国乡镇边界.shp")
        sys.exit(1)

    if not db_path.exists():
        print(f"Error: Database not found: {db_path}")
        sys.exit(1)

    # Run geocoding
    service = GeocodingService(shapefile_dir, db_path)
    service.run(batch_size=args.batch_size, limit=args.limit)


if __name__ == "__main__":
    main()
