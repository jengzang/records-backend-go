"""
Stay Annotation Worker

Generates semantic annotations for stay segments:
- Infers stay purpose (HOME, WORK, TRANSIT, VISIT, etc.)
- Suggests location labels based on context
- Calculates stay importance scores
- Provides annotation confidence levels

Algorithm:
1. For each stay segment, analyze:
   - Time of day (morning/afternoon/evening/night)
   - Day of week (weekday/weekend)
   - Duration (short/medium/long)
   - Frequency (how often visited)
   - Administrative level (province/city/county/town)

2. Infer stay purpose:
   - HOME: Night stays (22:00-06:00), high frequency, long duration
   - WORK: Weekday daytime (09:00-18:00), high frequency, medium-long duration
   - TRANSIT: Short duration (< 1 hour), transportation hubs
   - VISIT: Weekend or evening, medium duration, low-medium frequency
   - MEAL: Meal times (11:00-13:00, 17:00-19:00), short-medium duration
   - SHOPPING: Weekend afternoon, medium duration
   - UNKNOWN: Cannot determine

3. Generate location labels:
   - Use administrative division names
   - Add context (e.g., "Home in Beijing", "Work in Haidian District")
   - Suggest POI categories (residential, office, restaurant, etc.)

4. Calculate importance score (0-100):
   - Frequency weight: 40%
   - Duration weight: 30%
   - Recency weight: 20%
   - Administrative level weight: 10%

5. Update stay_segments table with:
   - annotation_label: Suggested label
   - annotation_confidence: Confidence (0.0-1.0)
   - importance_score: Importance (0-100)
   - metadata: JSON with detailed analysis
"""

import sys
import argparse
import json
from typing import List, Tuple, Dict, Any
from datetime import datetime
from collections import defaultdict

# Add parent directory to path for imports
sys.path.append('/app/scripts/common')
from task_executor import TaskExecutor


class StayAnnotationWorker(TaskExecutor):
    """Worker for stay segment annotation"""

    # Time of day definitions (hour ranges)
    TIME_OF_DAY = {
        'NIGHT': (22, 6),      # 22:00-06:00
        'MORNING': (6, 12),    # 06:00-12:00
        'AFTERNOON': (12, 18), # 12:00-18:00
        'EVENING': (18, 22)    # 18:00-22:00
    }

    # Meal time definitions
    MEAL_TIMES = {
        'BREAKFAST': (7, 9),
        'LUNCH': (11, 13),
        'DINNER': (17, 19)
    }

    # Duration thresholds (seconds)
    DURATION_SHORT = 3600      # 1 hour
    DURATION_MEDIUM = 7200     # 2 hours
    DURATION_LONG = 14400      # 4 hours

    def __init__(self, db_path: str, task_id: int):
        """
        Initialize stay annotation worker

        Args:
            db_path: Path to SQLite database
            task_id: ID of the analysis task
        """
        super().__init__(db_path, task_id)
        self.stay_frequency = defaultdict(int)  # Count visits per location

    def get_time_of_day(self, timestamp: int) -> str:
        """
        Get time of day category

        Args:
            timestamp: Unix timestamp

        Returns:
            Time of day category (NIGHT/MORNING/AFTERNOON/EVENING)
        """
        hour = datetime.fromtimestamp(timestamp).hour

        for category, (start, end) in self.TIME_OF_DAY.items():
            if start < end:
                if start <= hour < end:
                    return category
            else:  # Wraps around midnight (NIGHT)
                if hour >= start or hour < end:
                    return category

        return 'UNKNOWN'

    def is_meal_time(self, timestamp: int) -> str:
        """
        Check if timestamp is during meal time

        Args:
            timestamp: Unix timestamp

        Returns:
            Meal type (BREAKFAST/LUNCH/DINNER) or None
        """
        hour = datetime.fromtimestamp(timestamp).hour

        for meal, (start, end) in self.MEAL_TIMES.items():
            if start <= hour < end:
                return meal

        return None

    def is_weekday(self, timestamp: int) -> bool:
        """Check if timestamp is on a weekday"""
        weekday = datetime.fromtimestamp(timestamp).weekday()
        return weekday < 5  # Monday=0, Friday=4

    def infer_stay_purpose(self, stay: Dict[str, Any]) -> Tuple[str, float]:
        """
        Infer stay purpose based on context

        Args:
            stay: Stay segment data

        Returns:
            Tuple of (purpose, confidence)
        """
        start_time = stay['start_time']
        duration_s = stay['duration_s']
        frequency = self.stay_frequency.get(stay['location_key'], 1)

        time_of_day = self.get_time_of_day(start_time)
        is_weekday = self.is_weekday(start_time)
        meal_time = self.is_meal_time(start_time)

        # HOME: Night stays, high frequency, long duration
        if time_of_day == 'NIGHT' and duration_s > self.DURATION_LONG and frequency > 10:
            return ('HOME', 0.9)

        # WORK: Weekday daytime, high frequency, medium-long duration
        if is_weekday and time_of_day in ['MORNING', 'AFTERNOON'] and \
           duration_s > self.DURATION_MEDIUM and frequency > 5:
            return ('WORK', 0.8)

        # MEAL: Meal times, short-medium duration
        if meal_time and duration_s < self.DURATION_MEDIUM:
            return (f'MEAL_{meal_time}', 0.7)

        # TRANSIT: Short duration
        if duration_s < self.DURATION_SHORT:
            return ('TRANSIT', 0.6)

        # VISIT: Weekend or evening, medium duration
        if (not is_weekday or time_of_day == 'EVENING') and \
           self.DURATION_SHORT < duration_s < self.DURATION_LONG:
            return ('VISIT', 0.5)

        # UNKNOWN: Cannot determine
        return ('UNKNOWN', 0.3)

    def generate_label(self, stay: Dict[str, Any], purpose: str) -> str:
        """
        Generate human-readable label for stay

        Args:
            stay: Stay segment data
            purpose: Inferred purpose

        Returns:
            Label string
        """
        # Get location name
        location = stay.get('county') or stay.get('city') or stay.get('province') or 'Unknown'

        # Generate label based on purpose
        if purpose == 'HOME':
            return f"Home in {location}"
        elif purpose == 'WORK':
            return f"Work in {location}"
        elif purpose.startswith('MEAL_'):
            meal_type = purpose.split('_')[1].capitalize()
            return f"{meal_type} in {location}"
        elif purpose == 'TRANSIT':
            return f"Transit in {location}"
        elif purpose == 'VISIT':
            return f"Visit to {location}"
        else:
            return f"Stay in {location}"

    def calculate_importance(self, stay: Dict[str, Any], frequency: int) -> int:
        """
        Calculate importance score (0-100)

        Args:
            stay: Stay segment data
            frequency: Visit frequency

        Returns:
            Importance score
        """
        score = 0

        # Frequency weight (40%)
        freq_score = min(frequency / 20.0, 1.0) * 40
        score += freq_score

        # Duration weight (30%)
        duration_s = stay['duration_s']
        duration_score = min(duration_s / 86400.0, 1.0) * 30  # Normalize to 1 day
        score += duration_score

        # Recency weight (20%)
        # More recent stays are more important
        days_ago = (datetime.now().timestamp() - stay['end_time']) / 86400
        recency_score = max(0, 1 - days_ago / 365) * 20  # Normalize to 1 year
        score += recency_score

        # Administrative level weight (10%)
        # Higher level (province) = more important
        admin_level = stay.get('admin_level', 'COUNTY')
        admin_scores = {'PROVINCE': 10, 'CITY': 7, 'COUNTY': 5, 'TOWN': 3}
        score += admin_scores.get(admin_level, 5)

        return int(min(100, max(0, score)))

    def process_stays(self):
        """Process all stay segments and generate annotations"""
        # First pass: Count frequency for each location
        self.logger.info("Counting stay frequencies...")
        cursor = self.conn.execute(
            '''
            SELECT id, center_lat, center_lon, province, city, county, town
            FROM stay_segments
            WHERE confidence > 0.5
            '''
        )

        for row in cursor.fetchall():
            stay_id, lat, lon, province, city, county, town = row
            location_key = f"{county or city or province}_{lat:.4f}_{lon:.4f}"
            self.stay_frequency[location_key] += 1

        # Second pass: Annotate each stay
        self.logger.info("Annotating stay segments...")
        cursor = self.conn.execute(
            '''
            SELECT id, start_time, end_time, duration_s,
                   center_lat, center_lon,
                   province, city, county, town,
                   stay_type, confidence
            FROM stay_segments
            WHERE confidence > 0.5
            ORDER BY start_time
            '''
        )

        processed = 0
        failed = 0

        for row in cursor.fetchall():
            try:
                stay = {
                    'id': row[0],
                    'start_time': row[1],
                    'end_time': row[2],
                    'duration_s': row[3],
                    'center_lat': row[4],
                    'center_lon': row[5],
                    'province': row[6],
                    'city': row[7],
                    'county': row[8],
                    'town': row[9],
                    'stay_type': row[10],
                    'confidence': row[11]
                }

                # Generate location key
                location_key = f"{stay['county'] or stay['city'] or stay['province']}_{stay['center_lat']:.4f}_{stay['center_lon']:.4f}"
                frequency = self.stay_frequency[location_key]

                # Infer purpose
                purpose, purpose_confidence = self.infer_stay_purpose(stay)

                # Generate label
                label = self.generate_label(stay, purpose)

                # Calculate importance
                importance = self.calculate_importance(stay, frequency)

                # Prepare metadata
                metadata = {
                    'purpose': purpose,
                    'purpose_confidence': purpose_confidence,
                    'frequency': frequency,
                    'time_of_day': self.get_time_of_day(stay['start_time']),
                    'is_weekday': self.is_weekday(stay['start_time']),
                    'meal_time': self.is_meal_time(stay['start_time'])
                }

                # Update stay segment
                self.conn.execute(
                    '''
                    UPDATE stay_segments
                    SET annotation_label = ?,
                        annotation_confidence = ?,
                        importance_score = ?,
                        metadata = json_set(COALESCE(metadata, '{}'), '$.annotation', ?)
                    WHERE id = ?
                    ''',
                    (label, purpose_confidence, importance, json.dumps(metadata), stay['id'])
                )

                processed += 1

                if processed % 100 == 0:
                    self.conn.commit()
                    self.logger.info(f"Processed {processed} stays...")

            except Exception as e:
                self.logger.error(f"Failed to annotate stay {row[0]}: {e}")
                failed += 1

        # Final commit
        self.conn.commit()

        return processed, failed

    def run(self):
        """Main execution"""
        try:
            # Connect to database
            self.connect()

            # Get task information
            task_info = self.get_task_info()
            self.logger.info(f"Starting stay annotation (task {self.task_id})")

            # Mark task as running
            self.mark_running()

            # Process stays
            processed, failed = self.process_stays()

            # Calculate statistics
            success_rate = (processed - failed) / processed if processed > 0 else 0

            result_summary = {
                'processed': processed,
                'failed': failed,
                'success_rate': round(success_rate, 4)
            }

            # Mark task as completed
            self.mark_completed(result_summary)

            self.logger.info(f"Stay annotation completed: {processed} stays processed, {failed} failed")

        except Exception as e:
            self.logger.exception(f"Task execution failed: {e}")
            self.mark_failed(str(e))
            raise

        finally:
            # Always disconnect
            self.disconnect()


def main():
    """Main entry point"""
    parser = argparse.ArgumentParser(description='Stay Annotation Worker')
    parser.add_argument('--task-id', type=int, required=True,
                       help='Analysis task ID')
    parser.add_argument('--db-path', type=str,
                       default='/data/tracks/tracks.db',
                       help='Path to SQLite database')

    args = parser.parse_args()

    # Create and run worker
    worker = StayAnnotationWorker(
        db_path=args.db_path,
        task_id=args.task_id
    )

    worker.run()


if __name__ == '__main__':
    main()
