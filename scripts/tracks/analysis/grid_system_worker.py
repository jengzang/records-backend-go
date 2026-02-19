"""
Grid System Worker

Creates multi-level GeoHash-based grid cells for spatial aggregation.
Supports 5 levels from country/province (L1) to road level (L5).

Algorithm:
1. For each track point, calculate GeoHash at multiple precision levels
2. Aggregate statistics for each grid cell (point count, visit dates, modes, time range)
3. Calculate grid boundaries (bbox) from GeoHash
4. Store grid cells in grid_cells table
5. Update track points with default grid_id (L3) and grid_level

GeoHash Precision Levels:
- L1 (precision 4): ~40km x 20km (country/province level)
- L2 (precision 5): ~5km x 5km (city level)
- L3 (precision 6): ~1.2km x 0.6km (county level) - DEFAULT
- L4 (precision 7): ~150m x 150m (street level)
- L5 (precision 8): ~40m x 20m (road level) - optional, large data volume
"""

import sys
import argparse
import json
import math
from typing import List, Tuple, Dict, Any, Set
from datetime import datetime
from collections import defaultdict

# Add parent directory to path for imports
sys.path.append('/app/scripts/common')
from incremental_analyzer import IncrementalAnalyzer

try:
    import geohash2
except ImportError:
    print("ERROR: geohash2 library not installed. Install with: pip install geohash2")
    sys.exit(1)


class GridSystemWorker(IncrementalAnalyzer):
    """Worker for grid system creation and aggregation"""

    # GeoHash precision levels
    GRID_LEVELS = {
        1: 4,  # Country/Province (~40km x 20km)
        2: 5,  # City (~5km x 5km)
        3: 6,  # County (~1.2km x 0.6km) - DEFAULT
        4: 7,  # Street (~150m x 150m)
        5: 8   # Road (~40m x 20m) - optional
    }

    DEFAULT_LEVEL = 3  # County level

    def __init__(self, db_path: str, task_id: int, batch_size: int = 1000,
                 max_level: int = 4):
        """
        Initialize grid system worker

        Args:
            db_path: Path to SQLite database
            task_id: ID of the analysis task
            batch_size: Number of points to process per batch
            max_level: Maximum grid level to generate (1-5, default 4)
        """
        super().__init__(db_path, task_id, batch_size)
        self.max_level = min(max_level, 5)  # Cap at level 5
        self.grid_stats = defaultdict(lambda: {
            'point_count': 0,
            'visit_dates': set(),
            'modes': set(),
            'min_time': None,
            'max_time': None
        })

    def encode_geohash(self, lat: float, lon: float, precision: int) -> str:
        """
        Encode latitude/longitude to GeoHash

        Args:
            lat: Latitude
            lon: Longitude
            precision: GeoHash precision (4-8)

        Returns:
            GeoHash string
        """
        return geohash2.encode(lat, lon, precision)

    def decode_geohash(self, geohash_str: str) -> Dict[str, float]:
        """
        Decode GeoHash to center point and bounding box

        Args:
            geohash_str: GeoHash string

        Returns:
            Dict with center_lat, center_lon, min_lat, max_lat, min_lon, max_lon
        """
        # Get center point and error margins
        center_lat, center_lon, lat_err, lon_err = geohash2.decode_exactly(geohash_str)

        return {
            'center_lat': center_lat,
            'center_lon': center_lon,
            'min_lat': center_lat - lat_err,
            'max_lat': center_lat + lat_err,
            'min_lon': center_lon - lon_err,
            'max_lon': center_lon + lon_err
        }

    def process_batch(self, points: List[Tuple]) -> int:
        """
        Process a batch of points and aggregate into grid cells

        Args:
            points: List of point tuples

        Returns:
            Number of points that failed processing
        """
        failed = 0

        for point in points:
            try:
                # Extract point data
                point_id = point[0]
                dataTime = point[1]
                lon = point[2]
                lat = point[3]
                mode = None  # Will be fetched from database if available

                # Skip invalid coordinates
                if lat is None or lon is None or lat == 0 or lon == 0:
                    failed += 1
                    continue

                # Skip outlier points
                outlier_flag = self.get_outlier_flag(point_id)
                if outlier_flag:
                    continue

                # Get mode if available
                mode = self.get_mode(point_id)

                # Get date from timestamp
                visit_date = datetime.fromtimestamp(dataTime).strftime('%Y-%m-%d')

                # Calculate GeoHash for each level
                default_grid_id = None
                for level in range(1, self.max_level + 1):
                    precision = self.GRID_LEVELS[level]
                    geohash_str = self.encode_geohash(lat, lon, precision)
                    grid_id = f"L{level}_{geohash_str}"

                    # Store default level (L3)
                    if level == self.DEFAULT_LEVEL:
                        default_grid_id = grid_id

                    # Aggregate statistics
                    stats = self.grid_stats[grid_id]
                    stats['point_count'] += 1
                    stats['visit_dates'].add(visit_date)
                    if mode:
                        stats['modes'].add(mode)

                    # Update time range
                    if stats['min_time'] is None or dataTime < stats['min_time']:
                        stats['min_time'] = dataTime
                    if stats['max_time'] is None or dataTime > stats['max_time']:
                        stats['max_time'] = dataTime

                # Update track point with default grid_id
                if default_grid_id:
                    self.update_point_grid(point_id, default_grid_id, self.DEFAULT_LEVEL)

            except Exception as e:
                self.logger.error(f"Failed to process point {point[0]}: {e}")
                failed += 1

        # Flush grid cells to database every batch
        self.flush_grid_cells()

        return failed

    def get_outlier_flag(self, point_id: int) -> bool:
        """Get outlier flag for a point"""
        cursor = self.conn.execute(
            'SELECT outlier_flag FROM "一生足迹" WHERE id = ?',
            (point_id,)
        )
        row = cursor.fetchone()
        return row[0] if row and row[0] is not None else False

    def get_mode(self, point_id: int) -> str:
        """Get transport mode for a point"""
        cursor = self.conn.execute(
            'SELECT mode FROM "一生足迹" WHERE id = ?',
            (point_id,)
        )
        row = cursor.fetchone()
        return row[0] if row else None

    def update_point_grid(self, point_id: int, grid_id: str, grid_level: int):
        """Update track point with grid_id and grid_level"""
        self.conn.execute(
            '''
            UPDATE "一生足迹"
            SET grid_id = ?, grid_level = ?
            WHERE id = ?
            ''',
            (grid_id, grid_level, point_id)
        )

    def flush_grid_cells(self):
        """Flush aggregated grid cells to database"""
        if not self.grid_stats:
            return

        for grid_id, stats in self.grid_stats.items():
            # Extract level and geohash from grid_id
            level_str, geohash_str = grid_id.split('_', 1)
            level = int(level_str[1:])  # Remove 'L' prefix

            # Decode geohash to get boundaries
            bbox = self.decode_geohash(geohash_str)

            # Convert sets to JSON arrays
            modes_json = json.dumps(list(stats['modes']))
            visit_dates_json = json.dumps(sorted(list(stats['visit_dates'])))

            # Calculate visit count (unique dates)
            visit_count = len(stats['visit_dates'])

            # UPSERT grid cell
            self.conn.execute(
                '''
                INSERT INTO grid_cells (
                    grid_id, level, geohash, center_lat, center_lon,
                    min_lat, max_lat, min_lon, max_lon,
                    point_count, visit_count, modes, visit_dates,
                    first_visit, last_visit, created_at, updated_at
                )
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
                ON CONFLICT(grid_id) DO UPDATE SET
                    point_count = point_count + excluded.point_count,
                    visit_count = excluded.visit_count,
                    modes = excluded.modes,
                    visit_dates = excluded.visit_dates,
                    first_visit = MIN(first_visit, excluded.first_visit),
                    last_visit = MAX(last_visit, excluded.last_visit),
                    updated_at = CURRENT_TIMESTAMP
                ''',
                (
                    grid_id, level, geohash_str,
                    bbox['center_lat'], bbox['center_lon'],
                    bbox['min_lat'], bbox['max_lat'],
                    bbox['min_lon'], bbox['max_lon'],
                    stats['point_count'], visit_count,
                    modes_json, visit_dates_json,
                    stats['min_time'], stats['max_time']
                )
            )

        # Commit and clear buffer
        self.conn.commit()
        self.grid_stats.clear()

    def clear_previous_results(self):
        """Clear previous grid system results"""
        self.logger.info("Clearing previous grid system results...")

        # Clear grid_id and grid_level from track points
        self.conn.execute(
            '''
            UPDATE "一生足迹"
            SET grid_id = NULL, grid_level = NULL
            '''
        )

        # Delete all grid cells
        self.conn.execute('DELETE FROM grid_cells')

        self.conn.commit()
        self.logger.info("Previous results cleared")


def main():
    """Main entry point"""
    parser = argparse.ArgumentParser(description='Grid System Worker')
    parser.add_argument('--task-id', type=int, required=True,
                       help='Analysis task ID')
    parser.add_argument('--db-path', type=str,
                       default='/data/tracks/tracks.db',
                       help='Path to SQLite database')
    parser.add_argument('--batch-size', type=int, default=1000,
                       help='Batch size for processing')
    parser.add_argument('--max-level', type=int, default=4,
                       help='Maximum grid level (1-5, default 4)')

    args = parser.parse_args()

    # Create and run worker
    worker = GridSystemWorker(
        db_path=args.db_path,
        task_id=args.task_id,
        batch_size=args.batch_size,
        max_level=args.max_level
    )

    worker.run()


if __name__ == '__main__':
    main()
