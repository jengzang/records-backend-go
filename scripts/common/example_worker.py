"""
Example Analysis Worker

This is a template/example showing how to create an analysis worker
using the TaskExecutor and IncrementalAnalyzer base classes.

To create a new analysis skill:
1. Copy this file and rename it (e.g., outlier_detection_worker.py)
2. Implement the process_batch() method with your analysis logic
3. Optionally implement clear_previous_results() for full recompute support
4. Create a Dockerfile for the worker
5. Register the skill in the Go backend service
"""

import argparse
import sys
import os

# Add parent directory to path to import common modules
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from common.incremental_analyzer import IncrementalAnalyzer, FullRecomputeAnalyzer


class ExampleAnalyzer(FullRecomputeAnalyzer):
    """
    Example analyzer implementation

    This demonstrates the minimal implementation required for an analysis worker.
    """

    def process_batch(self, points):
        """
        Process a batch of trajectory points

        Args:
            points: List of point tuples from database

        Returns:
            Number of points that failed processing
        """
        failed_count = 0

        for point in points:
            try:
                # Extract point data
                point_id = point['id']
                datatime = point['dataTime']
                longitude = point['longitude']
                latitude = point['latitude']
                speed = point['speed']
                province = point['province']
                city = point['city']

                # TODO: Implement your analysis logic here
                # Example: Detect outliers, classify transport mode, etc.

                # Example: Update point with analysis results
                self.conn.execute("""
                    UPDATE "一生足迹"
                    SET segment_id = ?,
                        mode = ?,
                        mode_confidence = ?,
                        updated_at = CURRENT_TIMESTAMP
                    WHERE id = ?
                """, (1, 'UNKNOWN', 0.5, point_id))

            except Exception as e:
                self.logger.error(f"Failed to process point {point_id}: {e}")
                failed_count += 1

        # Commit batch updates
        self.conn.commit()

        return failed_count

    def clear_previous_results(self):
        """
        Clear previous analysis results for full recompute

        This method is called before processing when task_type is FULL_RECOMPUTE.
        """
        self.logger.info("Clearing previous analysis results...")

        # Example: Clear segment_id and mode fields
        self.conn.execute("""
            UPDATE "一生足迹"
            SET segment_id = NULL,
                mode = NULL,
                mode_confidence = NULL,
                mode_reason_codes = NULL
        """)
        self.conn.commit()

        self.logger.info("Previous results cleared")


def main():
    """
    Main entry point for the worker

    Parses command-line arguments and runs the analyzer.
    """
    parser = argparse.ArgumentParser(description='Example Analysis Worker')
    parser.add_argument('--task-id', type=int, required=True,
                        help='Analysis task ID')
    parser.add_argument('--db-path', type=str,
                        default='/data/tracks.db',
                        help='Path to SQLite database')
    parser.add_argument('--batch-size', type=int, default=1000,
                        help='Batch size for processing')

    args = parser.parse_args()

    # Create and run analyzer
    analyzer = ExampleAnalyzer(
        db_path=args.db_path,
        task_id=args.task_id,
        batch_size=args.batch_size
    )

    try:
        analyzer.run()
        sys.exit(0)
    except Exception as e:
        print(f"Worker failed: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == '__main__':
    main()
