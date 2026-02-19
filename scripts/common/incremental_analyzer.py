"""
Incremental Analyzer Base Class

Extends TaskExecutor to provide incremental analysis functionality:
- Fetches only unanalyzed points from database
- Processes data in batches for memory efficiency
- Automatically updates progress and ETA
- Supports both incremental and full recompute modes
"""

import time
from typing import List, Tuple, Optional
from .task_executor import TaskExecutor


class IncrementalAnalyzer(TaskExecutor):
    """Base class for incremental analysis tasks"""

    def __init__(self, db_path: str, task_id: int, batch_size: int = 1000):
        """
        Initialize incremental analyzer

        Args:
            db_path: Path to SQLite database
            task_id: ID of the analysis task
            batch_size: Number of points to process per batch
        """
        super().__init__(db_path, task_id)
        self.batch_size = batch_size
        self.task_info = None

    def get_unanalyzed_points(self, limit: Optional[int] = None) -> List[Tuple]:
        """
        Get points that haven't been analyzed yet

        Args:
            limit: Maximum number of points to fetch (default: batch_size)

        Returns:
            List of tuples containing point data
        """
        if not self.conn:
            raise RuntimeError("Database connection not established")

        if limit is None:
            limit = self.batch_size

        # Query depends on task type
        if self.task_info['task_type'] == 'INCREMENTAL':
            # Only get points without segment_id (unanalyzed)
            query = """
                SELECT id, dataTime, longitude, latitude, heading,
                       accuracy, speed, distance, altitude,
                       province, city, county, town, village,
                       time_visually, time
                FROM "一生足迹"
                WHERE segment_id IS NULL
                ORDER BY dataTime
                LIMIT ?
            """
        else:
            # Full recompute: get all points
            query = """
                SELECT id, dataTime, longitude, latitude, heading,
                       accuracy, speed, distance, altitude,
                       province, city, county, town, village,
                       time_visually, time
                FROM "一生足迹"
                ORDER BY dataTime
                LIMIT ?
            """

        cursor = self.conn.execute(query, (limit,))
        return cursor.fetchall()

    def process_batch(self, points: List[Tuple]) -> int:
        """
        Process a batch of points - must be implemented by subclasses

        Args:
            points: List of point tuples to process

        Returns:
            Number of points that failed processing

        Raises:
            NotImplementedError: If not overridden by subclass
        """
        raise NotImplementedError("Subclasses must implement process_batch() method")

    def run(self):
        """
        Main execution loop for incremental analysis

        This method:
        1. Connects to database
        2. Marks task as running
        3. Processes points in batches
        4. Updates progress after each batch
        5. Marks task as completed or failed
        """
        try:
            # Connect to database
            self.connect()

            # Get task information
            self.task_info = self.get_task_info()
            self.logger.info(
                f"Starting {self.task_info['skill_name']} analysis "
                f"(mode: {self.task_info['task_type']}, "
                f"total points: {self.task_info['total_points']})"
            )

            # Mark task as running
            self.mark_running()

            # Process points in batches
            processed = 0
            failed = 0
            total = self.task_info['total_points']

            while processed < total:
                # Fetch next batch
                points = self.get_unanalyzed_points()
                if not points:
                    self.logger.info("No more points to process")
                    break

                # Process batch
                batch_start = time.time()
                batch_failed = self.process_batch(points)
                batch_duration = time.time() - batch_start

                # Update counters
                processed += len(points)
                failed += batch_failed

                # Calculate progress and ETA
                progress_percent = int((processed / total) * 100)
                eta_seconds = self.calculate_eta(processed, total)

                # Update progress in database
                self.update_progress(
                    processed=processed,
                    failed=failed,
                    progress_percent=progress_percent,
                    eta_seconds=eta_seconds
                )

                # Log batch completion
                self.logger.info(
                    f"Batch completed: {len(points)} points in {batch_duration:.2f}s "
                    f"({len(points)/batch_duration:.1f} points/sec), "
                    f"{batch_failed} failed"
                )

            # Calculate final statistics
            success_rate = (processed - failed) / processed if processed > 0 else 0
            total_duration = time.time() - self.start_time if self.start_time else 0

            result_summary = {
                'processed': processed,
                'failed': failed,
                'success_rate': round(success_rate, 4),
                'duration_seconds': int(total_duration),
                'avg_speed': round(processed / total_duration, 2) if total_duration > 0 else 0
            }

            # Mark task as completed
            self.mark_completed(result_summary)

        except Exception as e:
            self.logger.exception(f"Task execution failed: {e}")
            self.mark_failed(str(e))
            raise

        finally:
            # Always disconnect
            self.disconnect()


class FullRecomputeAnalyzer(IncrementalAnalyzer):
    """
    Analyzer for full recompute mode

    This class extends IncrementalAnalyzer but clears previous analysis results
    before processing.
    """

    def clear_previous_results(self):
        """
        Clear previous analysis results - must be implemented by subclasses

        This method should delete or reset analysis-specific data in the database.

        Raises:
            NotImplementedError: If not overridden by subclass
        """
        raise NotImplementedError("Subclasses must implement clear_previous_results() method")

    def run(self):
        """
        Main execution loop for full recompute

        This method:
        1. Clears previous results
        2. Runs the standard incremental analysis loop
        """
        try:
            self.connect()
            self.task_info = self.get_task_info()

            if self.task_info['task_type'] == 'FULL_RECOMPUTE':
                self.logger.info("Clearing previous analysis results...")
                self.clear_previous_results()
                self.logger.info("Previous results cleared")

            # Disconnect and reconnect to ensure clean state
            self.disconnect()

            # Run standard incremental analysis
            super().run()

        except Exception as e:
            self.logger.exception(f"Full recompute failed: {e}")
            if self.conn:
                self.mark_failed(str(e))
                self.disconnect()
            raise
