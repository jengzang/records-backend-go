#!/usr/bin/env python3
"""
Geocoding worker for Docker container execution.
Integrates with Go backend task management system.

Usage:
    python geocode_worker.py --task-id <task_id> [--batch-size 1000]
"""

import sqlite3
import sys
import time
import argparse
from pathlib import Path
from typing import List, Tuple, Optional
import pickle

try:
    import geopandas as gpd
    from shapely.geometry import Point
    import pandas as pd
    import geohash2 as geohash
except ImportError:
    print("Error: Required packages not installed.")
    print("Please install: pip install geopandas shapely pyproj pandas geohash2")
    sys.exit(1)


class GeoHashCache:
    """GeoHash-based caching for geocoding results."""

    def __init__(self, precision=6):
        """
        Initialize cache.

        Args:
            precision: GeoHash precision (6 = ~1.2km x 0.6km)
        """
        self.precision = precision
        self.cache = {}

    def get(self, longitude: float, latitude: float) -> Optional[dict]:
        """Get cached result for a point."""
        gh = geohash.encode(latitude, longitude, precision=self.precision)
        return self.cache.get(gh)

    def set(self, longitude: float, latitude: float, result: dict):
        """Cache result for a point."""
        gh = geohash.encode(latitude, longitude, precision=self.precision)
        self.cache[gh] = result

    def save(self, filepath: Path):
        """Save cache to file."""
        with open(filepath, 'wb') as f:
            pickle.dump(self.cache, f)

    def load(self, filepath: Path):
        """Load cache from file."""
        if filepath.exists():
            with open(filepath, 'rb') as f:
                self.cache = pickle.load(f)
            print(f"Loaded cache with {len(self.cache)} entries")


class TaskManager:
    """Manages task status updates in the database."""

    def __init__(self, db_path: Path, task_id: int):
        self.db_path = db_path
        self.task_id = task_id

    def update_status(self, status: str, error_message: str = None):
        """Update task status."""
        conn = sqlite3.connect(str(self.db_path))
        cursor = conn.cursor()

        if status == 'running':
            cursor.execute("""
                UPDATE geocoding_tasks
                SET status = ?, start_time = datetime('now')
                WHERE id = ?
            """, (status, self.task_id))
        elif status == 'completed':
            cursor.execute("""
                UPDATE geocoding_tasks
                SET status = ?, end_time = datetime('now')
                WHERE id = ?
            """, (status, self.task_id))
        elif status == 'failed':
            cursor.execute("""
                UPDATE geocoding_tasks
                SET status = ?, end_time = datetime('now'), error_message = ?
                WHERE id = ?
            """, (status, error_message, self.task_id))

        conn.commit()
        conn.close()

    def update_progress(self, processed: int, failed: int, eta_seconds: int = None):
        """Update task progress."""
        conn = sqlite3.connect(str(self.db_path))
        cursor = conn.cursor()

        cursor.execute("""
            UPDATE geocoding_tasks
            SET processed_points = ?, failed_points = ?, eta_seconds = ?
            WHERE id = ?
        """, (processed, failed, eta_seconds, self.task_id))

        conn.commit()
        conn.close()


class GeocodingWorker:
    """Geocoding worker with task management integration."""

    def __init__(self, shapefile_dir: Path, db_path: Path, task_id: int, cache: GeoHashCache):
        self.shapefile_dir = shapefile_dir
        self.db_path = db_path
        self.task_id = task_id
        self.cache = cache
        self.task_manager = TaskManager(db_path, task_id)
        self.gdf = None

        # Load shapefile
        self._load_shapefiles()

    def _load_shapefiles(self):
        """Load administrative boundary shapefile."""
        print("Loading shapefile...")

        shapefile_path = self.shapefile_dir / "2024全国乡镇边界" / "2024全国乡镇边界.shp"

        if not shapefile_path.exists():
            raise FileNotFoundError(f"Shapefile not found: {shapefile_path}")

        self.gdf = gpd.read_file(shapefile_path)
        self.gdf.sindex  # Create spatial index
        print(f"  Loaded shapefile: {len(self.gdf)} features")

    def geocode_point(self, longitude: float, latitude: float) -> dict:
        """Reverse geocode a single point with caching."""
        # Check cache first
        cached = self.cache.get(longitude, latitude)
        if cached:
            return cached

        point = Point(longitude, latitude)
        result = {
            'province': None,
            'city': None,
            'county': None,
            'town': None,
            'village': None,
        }

        try:
            possible_matches_index = list(self.gdf.sindex.intersection(point.bounds))
            possible_matches = self.gdf.iloc[possible_matches_index]
            matches = possible_matches[possible_matches.contains(point)]

            if not matches.empty:
                match = matches.iloc[0]
                columns = list(self.gdf.columns)
                result['province'] = match[columns[1]]
                result['city'] = match[columns[2]]
                result['county'] = match[columns[4]]
                result['town'] = match[columns[6]]

                # Cache the result
                self.cache.set(longitude, latitude, result)

        except Exception as e:
            print(f"    Warning: Error geocoding point ({longitude}, {latitude}): {e}")

        return result

    def get_ungeocoded_points(self) -> List[Tuple]:
        """Get track points without administrative divisions."""
        conn = sqlite3.connect(str(self.db_path))
        cursor = conn.cursor()

        query = '''
            SELECT id, longitude, latitude
            FROM "一生足迹"
            WHERE province IS NULL OR province = ''
            ORDER BY dataTime ASC
        '''

        cursor.execute(query)
        points = cursor.fetchall()
        conn.close()

        return points

    def update_admin_divisions(self, updates: List[Tuple]):
        """Update administrative divisions in database."""
        conn = sqlite3.connect(str(self.db_path))
        cursor = conn.cursor()

        cursor.execute("PRAGMA journal_mode=WAL")

        update_sql = '''
            UPDATE "一生足迹"
            SET province = ?, city = ?, county = ?, town = ?, village = ?,
                updated_at = datetime('now'), algo_version = '1.0'
            WHERE id = ?
        '''

        cursor.executemany(update_sql, [
            (u[1], u[2], u[3], u[4], u[5], u[0]) for u in updates
        ])

        conn.commit()
        conn.close()

    def run(self, batch_size: int = 1000):
        """Run geocoding process with task management."""
        print(f"\nStarting geocoding worker for task {self.task_id}")
        print(f"Database: {self.db_path}")
        print(f"Batch size: {batch_size}")

        # Mark task as running
        self.task_manager.update_status('running')

        try:
            # Get ungeocoded points
            print("\nFetching ungeocoded points...")
            points = self.get_ungeocoded_points()
            total_points = len(points)

            if total_points == 0:
                print("No ungeocoded points found.")
                self.task_manager.update_status('completed')
                return

            print(f"Found {total_points} ungeocoded points")

            # Process in batches
            start_time = time.time()
            processed = 0
            failed = 0
            batch = []

            for i, (point_id, longitude, latitude) in enumerate(points, 1):
                # Geocode point
                result = self.geocode_point(longitude, latitude)

                # Track failures
                if result['province'] is None:
                    failed += 1

                # Add to batch
                batch.append((
                    point_id,
                    result['province'],
                    result['city'],
                    result['county'],
                    result['town'],
                    result['village']
                ))

                # Update database when batch is full
                if len(batch) >= batch_size:
                    self.update_admin_divisions(batch)
                    processed += len(batch)
                    batch = []

                    # Calculate ETA
                    elapsed = time.time() - start_time
                    rate = processed / elapsed
                    remaining = total_points - processed
                    eta = int(remaining / rate) if rate > 0 else 0

                    # Update progress
                    self.task_manager.update_progress(processed, failed, eta)

                    print(f"  Progress: {processed}/{total_points} ({processed/total_points*100:.1f}%) "
                          f"| Speed: {rate:.1f} pts/sec | ETA: {eta}s | Failed: {failed}")

            # Update remaining points
            if batch:
                self.update_admin_divisions(batch)
                processed += len(batch)
                self.task_manager.update_progress(processed, failed, 0)

            # Mark task as completed
            elapsed = time.time() - start_time
            print(f"\nGeocoding completed!")
            print(f"  Total: {total_points}")
            print(f"  Processed: {processed}")
            print(f"  Failed: {failed}")
            print(f"  Success rate: {(processed-failed)/processed*100:.2f}%")
            print(f"  Time: {elapsed:.1f}s")
            print(f"  Speed: {processed/elapsed:.1f} pts/sec")

            self.task_manager.update_status('completed')

        except Exception as e:
            print(f"\nError: {e}")
            self.task_manager.update_status('failed', str(e))
            raise


def main():
    parser = argparse.ArgumentParser(description='Geocoding worker')
    parser.add_argument('--task-id', type=int, required=True, help='Task ID')
    parser.add_argument('--batch-size', type=int, default=1000, help='Batch size')
    args = parser.parse_args()

    # Paths (Docker container paths)
    db_path = Path('/data/tracks.db')
    shapefile_dir = Path('/geo')
    cache_path = Path('/data/geocoding_cache.pkl')

    # Load cache
    cache = GeoHashCache(precision=6)
    if cache_path.exists():
        cache.load(cache_path)

    # Run worker
    worker = GeocodingWorker(shapefile_dir, db_path, args.task_id, cache)
    worker.run(args.batch_size)

    # Save cache
    cache.save(cache_path)
    print(f"\nCache saved with {len(cache.cache)} entries")


if __name__ == "__main__":
    main()
