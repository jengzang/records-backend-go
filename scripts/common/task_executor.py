"""
Task Executor Base Class

Provides common functionality for all analysis task executors:
- Database connection management
- Progress tracking and updates
- Task status management (running, completed, failed)
- Error handling and logging
"""

import sqlite3
import json
import time
import logging
from typing import Optional, Dict, Any

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)


class TaskExecutor:
    """Base class for all analysis task executors"""

    def __init__(self, db_path: str, task_id: int):
        """
        Initialize task executor

        Args:
            db_path: Path to SQLite database
            task_id: ID of the analysis task
        """
        self.db_path = db_path
        self.task_id = task_id
        self.conn = None
        self.logger = logging.getLogger(self.__class__.__name__)
        self.start_time = None

    def connect(self):
        """Establish database connection"""
        self.conn = sqlite3.connect(self.db_path)
        self.conn.row_factory = sqlite3.Row  # Enable column access by name
        self.logger.info(f"Connected to database: {self.db_path}")

    def disconnect(self):
        """Close database connection"""
        if self.conn:
            self.conn.close()
            self.logger.info("Database connection closed")

    def update_progress(
        self,
        processed: int,
        failed: int = 0,
        progress_percent: Optional[int] = None,
        eta_seconds: Optional[int] = None
    ):
        """
        Update task progress in database

        Args:
            processed: Number of points processed
            failed: Number of points failed
            progress_percent: Progress percentage (0-100)
            eta_seconds: Estimated time to completion in seconds
        """
        if not self.conn:
            raise RuntimeError("Database connection not established")

        self.conn.execute("""
            UPDATE analysis_tasks
            SET processed_points = ?,
                failed_points = ?,
                progress_percent = ?,
                eta_seconds = ?,
                updated_at = CURRENT_TIMESTAMP
            WHERE id = ?
        """, (processed, failed, progress_percent, eta_seconds, self.task_id))
        self.conn.commit()

        self.logger.info(
            f"Progress updated: {processed} processed, {failed} failed, "
            f"{progress_percent}% complete, ETA: {eta_seconds}s"
        )

    def mark_running(self):
        """Mark task as running"""
        if not self.conn:
            raise RuntimeError("Database connection not established")

        self.start_time = time.time()
        start_timestamp = int(self.start_time)

        self.conn.execute("""
            UPDATE analysis_tasks
            SET status = 'running',
                start_time = ?,
                updated_at = CURRENT_TIMESTAMP
            WHERE id = ?
        """, (start_timestamp, self.task_id))
        self.conn.commit()

        self.logger.info(f"Task {self.task_id} marked as running")

    def mark_completed(self, result_summary: Dict[str, Any]):
        """
        Mark task as completed with result summary

        Args:
            result_summary: Dictionary containing summary statistics
        """
        if not self.conn:
            raise RuntimeError("Database connection not established")

        end_time = int(time.time())
        result_json = json.dumps(result_summary)

        self.conn.execute("""
            UPDATE analysis_tasks
            SET status = 'completed',
                end_time = ?,
                result_summary = ?,
                progress_percent = 100,
                updated_at = CURRENT_TIMESTAMP
            WHERE id = ?
        """, (end_time, result_json, self.task_id))
        self.conn.commit()

        duration = end_time - int(self.start_time) if self.start_time else 0
        self.logger.info(
            f"Task {self.task_id} completed in {duration}s. "
            f"Summary: {result_summary}"
        )

    def mark_failed(self, error_message: str):
        """
        Mark task as failed with error message

        Args:
            error_message: Error description
        """
        if not self.conn:
            raise RuntimeError("Database connection not established")

        end_time = int(time.time())

        self.conn.execute("""
            UPDATE analysis_tasks
            SET status = 'failed',
                end_time = ?,
                error_message = ?,
                updated_at = CURRENT_TIMESTAMP
            WHERE id = ?
        """, (end_time, error_message, self.task_id))
        self.conn.commit()

        self.logger.error(f"Task {self.task_id} failed: {error_message}")

    def get_task_info(self) -> Dict[str, Any]:
        """
        Get task information from database

        Returns:
            Dictionary containing task details
        """
        if not self.conn:
            raise RuntimeError("Database connection not established")

        cursor = self.conn.execute("""
            SELECT skill_name, task_type, params_json, total_points
            FROM analysis_tasks
            WHERE id = ?
        """, (self.task_id,))

        row = cursor.fetchone()
        if not row:
            raise ValueError(f"Task {self.task_id} not found")

        return {
            'skill_name': row['skill_name'],
            'task_type': row['task_type'],
            'params': json.loads(row['params_json']) if row['params_json'] else {},
            'total_points': row['total_points']
        }

    def calculate_eta(self, processed: int, total: int) -> int:
        """
        Calculate estimated time to completion

        Args:
            processed: Number of points processed so far
            total: Total number of points to process

        Returns:
            Estimated seconds to completion
        """
        if not self.start_time or processed == 0:
            return 0

        elapsed = time.time() - self.start_time
        rate = processed / elapsed  # points per second
        remaining = total - processed

        if rate > 0:
            return int(remaining / rate)
        return 0

    def run(self):
        """
        Main execution method - must be implemented by subclasses

        Raises:
            NotImplementedError: If not overridden by subclass
        """
        raise NotImplementedError("Subclasses must implement run() method")
