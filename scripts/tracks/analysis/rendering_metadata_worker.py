"""
Rendering Metadata Worker

Generates visualization metadata for frontend map rendering:
- Speed-based color coding
- LOD (Level of Detail) control for performance
- Line width based on transport mode
- Opacity based on data quality

Algorithm:
1. For each track point, calculate rendering properties:
   - render_color: Speed-based color (hex format)
   - render_width: Line width based on mode (1-5 pixels)
   - render_opacity: Opacity based on accuracy and outlier status (0.0-1.0)
   - lod_level: Level of detail for map zoom (1-5)

2. Speed-based color scheme:
   - STAY (0-1 km/h): #808080 (gray)
   - WALK (1-10 km/h): #00FF00 (green)
   - CAR (10-80 km/h): #FFA500 (orange)
   - TRAIN (80-200 km/h): #FF0000 (red)
   - FLIGHT (200+ km/h): #0000FF (blue)

3. LOD levels (for map zoom optimization):
   - L1: Major highways/flights only (speed > 80 km/h)
   - L2: All motorized transport (speed > 10 km/h)
   - L3: Include walking (speed > 1 km/h) - DEFAULT
   - L4: Include stays (all points)
   - L5: Full detail with synthetic points

4. Line width by mode:
   - WALK: 1px
   - CAR: 2px
   - TRAIN: 3px
   - FLIGHT: 4px
   - STAY: 1px (point marker)
"""

import sys
import argparse
from typing import List, Tuple, Dict, Any

# Add parent directory to path for imports
sys.path.append('/app/scripts/common')
from incremental_analyzer import IncrementalAnalyzer


class RenderingMetadataWorker(IncrementalAnalyzer):
    """Worker for generating rendering metadata"""

    # Speed-based color scheme (hex colors)
    SPEED_COLORS = {
        'STAY': '#808080',    # Gray
        'WALK': '#00FF00',    # Green
        'CAR': '#FFA500',     # Orange
        'TRAIN': '#FF0000',   # Red
        'FLIGHT': '#0000FF'   # Blue
    }

    # Mode-based line width (pixels)
    MODE_WIDTHS = {
        'STAY': 1,
        'WALK': 1,
        'CAR': 2,
        'TRAIN': 3,
        'FLIGHT': 4,
        'UNKNOWN': 1
    }

    def __init__(self, db_path: str, task_id: int, batch_size: int = 1000):
        """
        Initialize rendering metadata worker

        Args:
            db_path: Path to SQLite database
            task_id: ID of the analysis task
            batch_size: Number of points to process per batch
        """
        super().__init__(db_path, task_id, batch_size)

    def get_color_by_speed(self, speed: float, mode: str = None) -> str:
        """
        Get color based on speed and mode

        Args:
            speed: Speed in km/h
            mode: Transport mode (optional)

        Returns:
            Hex color string
        """
        if mode and mode in self.SPEED_COLORS:
            return self.SPEED_COLORS[mode]

        # Fallback to speed-based coloring
        if speed < 1:
            return self.SPEED_COLORS['STAY']
        elif speed < 10:
            return self.SPEED_COLORS['WALK']
        elif speed < 80:
            return self.SPEED_COLORS['CAR']
        elif speed < 200:
            return self.SPEED_COLORS['TRAIN']
        else:
            return self.SPEED_COLORS['FLIGHT']

    def get_lod_level(self, speed: float, mode: str = None, is_synthetic: bool = False) -> int:
        """
        Get LOD level for map zoom optimization

        Args:
            speed: Speed in km/h
            mode: Transport mode (optional)
            is_synthetic: Whether point is synthetic

        Returns:
            LOD level (1-5)
        """
        # L5: Full detail (synthetic points)
        if is_synthetic:
            return 5

        # L1: Major highways/flights only
        if speed > 80:
            return 1

        # L2: All motorized transport
        if speed > 10:
            return 2

        # L3: Include walking (DEFAULT)
        if speed > 1:
            return 3

        # L4: Include stays
        return 4

    def get_opacity(self, accuracy: float, outlier_flag: bool, mode_confidence: float = None) -> float:
        """
        Get opacity based on data quality

        Args:
            accuracy: GPS accuracy in meters
            outlier_flag: Whether point is marked as outlier
            mode_confidence: Mode classification confidence (0-1)

        Returns:
            Opacity value (0.0-1.0)
        """
        # Start with full opacity
        opacity = 1.0

        # Reduce opacity for outliers
        if outlier_flag:
            opacity *= 0.3

        # Reduce opacity for low accuracy
        if accuracy and accuracy > 100:
            opacity *= 0.5
        elif accuracy and accuracy > 50:
            opacity *= 0.7

        # Reduce opacity for low confidence
        if mode_confidence and mode_confidence < 0.5:
            opacity *= 0.6

        return max(0.1, min(1.0, opacity))  # Clamp to [0.1, 1.0]

    def process_batch(self, points: List[Tuple]) -> int:
        """
        Process a batch of points and generate rendering metadata

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
                speed = point[6] if len(point) > 6 else 0  # speed field
                accuracy = point[5] if len(point) > 5 else None  # accuracy field

                # Get additional data from database
                mode, mode_confidence, outlier_flag, is_synthetic = self.get_point_metadata(point_id)

                # Calculate rendering properties
                render_color = self.get_color_by_speed(speed, mode)
                render_width = self.MODE_WIDTHS.get(mode, 1) if mode else 1
                render_opacity = self.get_opacity(accuracy, outlier_flag, mode_confidence)
                lod_level = self.get_lod_level(speed, mode, is_synthetic)

                # Update track point
                self.update_rendering_metadata(
                    point_id,
                    render_color,
                    render_width,
                    render_opacity,
                    lod_level
                )

            except Exception as e:
                self.logger.error(f"Failed to process point {point[0]}: {e}")
                failed += 1

        # Commit after each batch
        self.conn.commit()

        return failed

    def get_point_metadata(self, point_id: int) -> Tuple[str, float, bool, bool]:
        """
        Get additional metadata for a point

        Args:
            point_id: Point ID

        Returns:
            Tuple of (mode, mode_confidence, outlier_flag, is_synthetic)
        """
        cursor = self.conn.execute(
            '''
            SELECT mode, mode_confidence, outlier_flag, is_synthetic
            FROM "一生足迹"
            WHERE id = ?
            ''',
            (point_id,)
        )
        row = cursor.fetchone()

        if row:
            return (
                row[0],  # mode
                row[1] if row[1] is not None else 0.5,  # mode_confidence
                bool(row[2]) if row[2] is not None else False,  # outlier_flag
                bool(row[3]) if row[3] is not None else False   # is_synthetic
            )
        else:
            return (None, 0.5, False, False)

    def update_rendering_metadata(self, point_id: int, color: str, width: int,
                                   opacity: float, lod_level: int):
        """
        Update track point with rendering metadata

        Args:
            point_id: Point ID
            color: Hex color string
            width: Line width in pixels
            opacity: Opacity value (0.0-1.0)
            lod_level: LOD level (1-5)
        """
        self.conn.execute(
            '''
            UPDATE "一生足迹"
            SET render_color = ?,
                render_width = ?,
                render_opacity = ?,
                lod_level = ?
            WHERE id = ?
            ''',
            (color, width, opacity, lod_level, point_id)
        )

    def clear_previous_results(self):
        """Clear previous rendering metadata"""
        self.logger.info("Clearing previous rendering metadata...")

        self.conn.execute(
            '''
            UPDATE "一生足迹"
            SET render_color = NULL,
                render_width = NULL,
                render_opacity = NULL,
                lod_level = NULL
            '''
        )

        self.conn.commit()
        self.logger.info("Previous results cleared")


def main():
    """Main entry point"""
    parser = argparse.ArgumentParser(description='Rendering Metadata Worker')
    parser.add_argument('--task-id', type=int, required=True,
                       help='Analysis task ID')
    parser.add_argument('--db-path', type=str,
                       default='/data/tracks/tracks.db',
                       help='Path to SQLite database')
    parser.add_argument('--batch-size', type=int, default=1000,
                       help='Batch size for processing')

    args = parser.parse_args()

    # Create and run worker
    worker = RenderingMetadataWorker(
        db_path=args.db_path,
        task_id=args.task_id,
        batch_size=args.batch_size
    )

    worker.run()


if __name__ == '__main__':
    main()
