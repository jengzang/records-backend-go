#!/usr/bin/env python3
"""
Skill: spatial_persona
Purpose: Generate spatial behavior persona profile
Algorithm: Aggregate all spatial analysis results into persona dimensions
"""

import sys
import json
import sqlite3
import numpy as np
from datetime import datetime

class SpatialPersonaWorker:
    def __init__(self, db_path, task_id):
        self.db_path = db_path
        self.task_id = task_id
        self.conn = sqlite3.connect(db_path)
        self.conn.row_factory = sqlite3.Row

    def mark_running(self):
        cursor = self.conn.cursor()
        cursor.execute("""
            UPDATE analysis_tasks
            SET status = 'running', started_at = CURRENT_TIMESTAMP, progress = 0.0
            WHERE id = ?
        """, (self.task_id,))
        self.conn.commit()

    def update_progress(self, progress, message=""):
        cursor = self.conn.cursor()
        cursor.execute("""
            UPDATE analysis_tasks SET progress = ?, progress_message = ? WHERE id = ?
        """, (progress, message, self.task_id))
        self.conn.commit()

    def load_footprint_stats(self):
        cursor = self.conn.cursor()
        cursor.execute("SELECT * FROM footprint_stats LIMIT 1")
        return cursor.fetchone()

    def load_stay_stats(self):
        cursor = self.conn.cursor()
        cursor.execute("SELECT * FROM stay_stats")
        return cursor.fetchall()

    def load_grid_stats(self):
        cursor = self.conn.cursor()
        cursor.execute("SELECT COUNT(*) as total_grids, SUM(visit_count) as total_visits FROM grid_cells")
        return cursor.fetchone()

    def load_revisit_stats(self):
        cursor = self.conn.cursor()
        cursor.execute("SELECT * FROM revisit_patterns")
        return cursor.fetchall()

    def load_transport_modes(self):
        cursor = self.conn.cursor()
        cursor.execute("""
            SELECT transport_mode, COUNT(*) as count
            FROM segments
            WHERE transport_mode IS NOT NULL
            GROUP BY transport_mode
            ORDER BY count DESC
        """)
        return cursor.fetchall()

    def calculate_mobility_score(self, footprint_stats):
        """Calculate mobility score (0-100) based on distance and speed"""
        if not footprint_stats:
            return 0.0

        total_distance_km = footprint_stats['total_distance_m'] / 1000 if footprint_stats['total_distance_m'] else 0
        avg_speed_kmh = footprint_stats['avg_speed_kmh'] if footprint_stats['avg_speed_kmh'] else 0

        # Normalize: 10000km = 100 points, 50km/h avg = 100 points
        distance_score = min(100, (total_distance_km / 10000) * 100)
        speed_score = min(100, (avg_speed_kmh / 50) * 100)

        # Weighted average
        mobility_score = 0.7 * distance_score + 0.3 * speed_score
        return round(mobility_score, 2)

    def calculate_exploration_score(self, footprint_stats, grid_stats):
        """Calculate exploration score (0-100) based on unique locations"""
        if not footprint_stats:
            return 0.0

        unique_provinces = footprint_stats['unique_provinces'] if footprint_stats['unique_provinces'] else 0
        unique_cities = footprint_stats['unique_cities'] if footprint_stats['unique_cities'] else 0
        unique_grids = grid_stats['total_grids'] if grid_stats and grid_stats['total_grids'] else 0

        # Normalize: 30 provinces = 100, 300 cities = 100, 1000 grids = 100
        province_score = min(100, (unique_provinces / 30) * 100)
        city_score = min(100, (unique_cities / 300) * 100)
        grid_score = min(100, (unique_grids / 1000) * 100)

        # Weighted average
        exploration_score = 0.3 * province_score + 0.3 * city_score + 0.4 * grid_score
        return round(exploration_score, 2)

    def calculate_routine_score(self, revisit_patterns, grid_stats):
        """Calculate routine score (0-100) based on revisit patterns"""
        if not grid_stats or not grid_stats['total_grids']:
            return 0.0

        total_grids = grid_stats['total_grids']
        total_visits = grid_stats['total_visits'] if grid_stats['total_visits'] else 0

        # Calculate revisit ratio
        revisit_ratio = (total_visits - total_grids) / total_visits if total_visits > 0 else 0

        # Higher revisit ratio = higher routine score
        routine_score = revisit_ratio * 100
        return round(routine_score, 2)

    def calculate_diversity_score(self, transport_modes):
        """Calculate diversity score (0-100) based on transport mode variety"""
        if not transport_modes:
            return 0.0

        mode_count = len(transport_modes)
        total_segments = sum(m['count'] for m in transport_modes)

        # Calculate entropy (diversity)
        entropy = 0
        for mode in transport_modes:
            p = mode['count'] / total_segments
            if p > 0:
                entropy -= p * np.log2(p)

        # Normalize: max entropy for 5 modes = log2(5) ≈ 2.32
        max_entropy = np.log2(5)
        diversity_score = (entropy / max_entropy) * 100 if max_entropy > 0 else 0

        return round(diversity_score, 2)

    def generate_insights(self, scores, footprint_stats, transport_modes):
        """Generate textual insights based on scores"""
        insights = []

        # Mobility insights
        if scores['mobility_score'] > 80:
            insights.append("高度活跃的移动者，经常进行长距离旅行")
        elif scores['mobility_score'] < 30:
            insights.append("相对静态的生活方式，移动范围有限")

        # Exploration insights
        if scores['exploration_score'] > 80:
            insights.append("探索型人格，喜欢访问新地点")
        elif scores['exploration_score'] < 30:
            insights.append("偏好熟悉环境，较少探索新地点")

        # Routine insights
        if scores['routine_score'] > 70:
            insights.append("高度规律的生活模式，经常重访相同地点")
        elif scores['routine_score'] < 30:
            insights.append("灵活多变的生活方式，较少重复路线")

        # Diversity insights
        if scores['diversity_score'] > 70:
            insights.append("多样化的出行方式，善于使用各种交通工具")

        # Primary mode insight
        if transport_modes:
            primary_mode = transport_modes[0]['transport_mode']
            mode_names = {
                'WALK': '步行',
                'BIKE': '骑行',
                'CAR': '驾车',
                'TRAIN': '火车',
                'PLANE': '飞机'
            }
            mode_name = mode_names.get(primary_mode, primary_mode)
            insights.append(f"主要出行方式：{mode_name}")

        return insights

    def process(self):
        self.update_progress(0.2, "Loading footprint statistics...")
        footprint_stats = self.load_footprint_stats()

        self.update_progress(0.3, "Loading stay statistics...")
        stay_stats = self.load_stay_stats()

        self.update_progress(0.4, "Loading grid statistics...")
        grid_stats = self.load_grid_stats()

        self.update_progress(0.5, "Loading revisit patterns...")
        revisit_patterns = self.load_revisit_stats()

        self.update_progress(0.6, "Loading transport modes...")
        transport_modes = self.load_transport_modes()

        self.update_progress(0.7, "Calculating persona scores...")

        # Calculate scores
        mobility_score = self.calculate_mobility_score(footprint_stats)
        exploration_score = self.calculate_exploration_score(footprint_stats, grid_stats)
        routine_score = self.calculate_routine_score(revisit_patterns, grid_stats)
        diversity_score = self.calculate_diversity_score(transport_modes)

        scores = {
            'mobility_score': mobility_score,
            'exploration_score': exploration_score,
            'routine_score': routine_score,
            'diversity_score': diversity_score
        }

        self.update_progress(0.8, "Generating insights...")

        # Generate insights
        insights = self.generate_insights(scores, footprint_stats, transport_modes)

        # Prepare result
        result = {
            'persona_date': None,  # All-time persona
            'mobility_score': mobility_score,
            'exploration_score': exploration_score,
            'routine_score': routine_score,
            'diversity_score': diversity_score,
            'total_distance_km': footprint_stats['total_distance_m'] / 1000 if footprint_stats and footprint_stats['total_distance_m'] else 0,
            'unique_locations': grid_stats['total_grids'] if grid_stats else 0,
            'revisit_ratio': routine_score / 100,  # Convert back to ratio
            'primary_mode': transport_modes[0]['transport_mode'] if transport_modes else None,
            'insights_json': json.dumps(insights, ensure_ascii=False)
        }

        return result

    def save_results(self, result):
        cursor = self.conn.cursor()
        cursor.execute("DELETE FROM spatial_persona WHERE persona_date IS NULL")

        cursor.execute("""
            INSERT INTO spatial_persona (
                persona_date, mobility_score, exploration_score, routine_score, diversity_score,
                total_distance_km, unique_locations, revisit_ratio, primary_mode, insights_json, algo_version
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'v1')
        """, (
            result['persona_date'], result['mobility_score'], result['exploration_score'],
            result['routine_score'], result['diversity_score'], result['total_distance_km'],
            result['unique_locations'], result['revisit_ratio'], result['primary_mode'],
            result['insights_json']
        ))

        self.conn.commit()

    def mark_completed(self, summary):
        cursor = self.conn.cursor()
        cursor.execute("""
            UPDATE analysis_tasks
            SET status = 'completed', completed_at = CURRENT_TIMESTAMP,
                progress = 1.0, result_summary = ?
            WHERE id = ?
        """, (json.dumps(summary), self.task_id))
        self.conn.commit()

    def mark_failed(self, error_msg):
        cursor = self.conn.cursor()
        cursor.execute("""
            UPDATE analysis_tasks
            SET status = 'failed', completed_at = CURRENT_TIMESTAMP, error_message = ?
            WHERE id = ?
        """, (error_msg, self.task_id))
        self.conn.commit()

    def run(self):
        try:
            self.mark_running()
            self.update_progress(0.1, "Starting spatial persona analysis...")

            result = self.process()
            self.save_results(result)

            summary = {
                'mobility_score': result['mobility_score'],
                'exploration_score': result['exploration_score'],
                'routine_score': result['routine_score'],
                'diversity_score': result['diversity_score'],
                'insights_count': len(json.loads(result['insights_json']))
            }

            self.mark_completed(summary)
            print(f"Spatial persona analysis completed")
            return 0

        except Exception as e:
            error_msg = f"Spatial persona analysis failed: {str(e)}"
            print(error_msg, file=sys.stderr)
            self.mark_failed(error_msg)
            return 1

        finally:
            self.conn.close()

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: python spatial_persona.py <db_path> <task_id>")
        sys.exit(1)

    db_path = sys.argv[1]
    task_id = int(sys.argv[2])

    worker = SpatialPersonaWorker(db_path, task_id)
    sys.exit(worker.run())
