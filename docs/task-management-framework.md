# Task Management Framework

## Overview

The Task Management Framework provides a unified system for managing and executing trajectory analysis tasks. It supports:

- **Incremental Analysis**: Only process new, unanalyzed data
- **Full Recompute**: Reprocess all historical data
- **Progress Tracking**: Real-time progress updates with ETA
- **Task Dependencies**: Automatic execution of dependent tasks
- **Docker Integration**: Isolated Python workers for each analysis skill

## Architecture

```
User Request → Go Backend API → Create Task Record → Start Docker Container
                                        ↓
                                  Python Worker
                                        ↓
                                  Process Data in Batches
                                        ↓
                                  Update Progress → Database
                                        ↓
                                  Mark Completed/Failed
```

## Components

### 1. Go Backend (Task Orchestration)

**Repository**: `internal/repository/analysis_task_repository.go`
- CRUD operations for analysis_tasks table
- Progress tracking methods
- Status management (pending, running, completed, failed)

**Service**: `internal/service/analysis_task_service.go`
- Task creation and validation
- Docker container management
- Task chain execution (with dependencies)

**Handler**: `internal/handler/analysis_task_handler.go`
- REST API endpoints
- Request validation
- Response formatting

**API Endpoints**:
```
POST   /api/admin/analysis/tasks          # Create a new task
GET    /api/admin/analysis/tasks          # List all tasks
GET    /api/admin/analysis/tasks/:id      # Get task status
DELETE /api/admin/analysis/tasks/:id      # Cancel a task
POST   /api/admin/analysis/trigger-chain  # Trigger analysis chain
```

### 2. Python Workers (Task Execution)

**Base Classes**:
- `scripts/common/task_executor.py` - Core task execution logic
- `scripts/common/incremental_analyzer.py` - Incremental analysis support

**Worker Template**:
- `scripts/common/example_worker.py` - Template for creating new workers

## Usage

### Creating a New Analysis Task

**Request**:
```bash
curl -X POST http://localhost:8080/api/admin/analysis/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "skill_name": "outlier_detection",
    "task_type": "INCREMENTAL",
    "params": {
      "threshold": 0.95
    }
  }'
```

**Response**:
```json
{
  "code": 200,
  "data": {
    "id": 123,
    "skill_name": "outlier_detection",
    "task_type": "INCREMENTAL",
    "status": "pending",
    "total_points": 12345,
    "processed_points": 0,
    "progress_percent": 0,
    "created_at": "2026-02-19T10:00:00Z"
  }
}
```

### Checking Task Status

**Request**:
```bash
curl http://localhost:8080/api/admin/analysis/tasks/123
```

**Response**:
```json
{
  "code": 200,
  "data": {
    "id": 123,
    "skill_name": "outlier_detection",
    "status": "running",
    "progress_percent": 65,
    "processed_points": 8024,
    "total_points": 12345,
    "failed_points": 23,
    "eta_seconds": 180,
    "start_time": 1708336800,
    "updated_at": "2026-02-19T10:05:30Z"
  }
}
```

### Triggering Analysis Chain

**Request**:
```bash
curl -X POST http://localhost:8080/api/admin/analysis/trigger-chain \
  -H "Content-Type: application/json" \
  -d '{
    "task_type": "INCREMENTAL"
  }'
```

**Response**:
```json
{
  "code": 200,
  "data": {
    "message": "Analysis chain triggered successfully",
    "task_ids": [123, 124, 125, 126, 127, 128, 129, 130]
  }
}
```

## Creating a New Analysis Worker

### Step 1: Implement Worker Class

Create a new file `scripts/tracks/analysis/my_skill_worker.py`:

```python
import argparse
import sys
import os

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__)))))

from common.incremental_analyzer import FullRecomputeAnalyzer


class MySkillAnalyzer(FullRecomputeAnalyzer):
    """My custom analysis skill"""

    def process_batch(self, points):
        """Process a batch of points"""
        failed_count = 0

        for point in points:
            try:
                # Your analysis logic here
                point_id = point['id']
                # ... analyze point ...

                # Update database
                self.conn.execute("""
                    UPDATE "一生足迹"
                    SET my_field = ?
                    WHERE id = ?
                """, (result, point_id))

            except Exception as e:
                self.logger.error(f"Failed to process point {point_id}: {e}")
                failed_count += 1

        self.conn.commit()
        return failed_count

    def clear_previous_results(self):
        """Clear previous results for full recompute"""
        self.conn.execute("""
            UPDATE "一生足迹"
            SET my_field = NULL
        """)
        self.conn.commit()


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--task-id', type=int, required=True)
    parser.add_argument('--db-path', type=str, default='/data/tracks.db')
    parser.add_argument('--batch-size', type=int, default=1000)
    args = parser.parse_args()

    analyzer = MySkillAnalyzer(args.db_path, args.task_id, args.batch_size)
    analyzer.run()


if __name__ == '__main__':
    main()
```

### Step 2: Create Dockerfile

Create `Dockerfile.my_skill`:

```dockerfile
FROM python:3.11-slim

# Install dependencies
RUN pip install --no-cache-dir numpy pandas

# Copy worker script
COPY scripts/tracks/analysis/my_skill_worker.py /app/worker.py
COPY scripts/common/ /app/common/

WORKDIR /app
CMD ["python", "worker.py"]
```

### Step 3: Build Docker Image

```bash
cd go-backend
docker build -f Dockerfile.my_skill -t records-analysis-my_skill:latest .
```

### Step 4: Register Skill in Go Backend

Add to `internal/service/analysis_task_service.go`:

```go
func isValidSkillName(skillName string) bool {
    validSkills := map[string]bool{
        // ... existing skills ...
        "my_skill": true,  // Add your skill here
    }
    return validSkills[skillName]
}
```

### Step 5: Test

```bash
curl -X POST http://localhost:8080/api/admin/analysis/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "skill_name": "my_skill",
    "task_type": "INCREMENTAL"
  }'
```

## Task Dependencies

The framework supports automatic task chain execution with dependencies:

```
outlier_detection
    ↓
transport_mode
    ↓
stay_detection
    ↓
trip_construction
    ↓
grid_system
    ↓
footprint_stats + stay_stats
    ↓
rendering_metadata
```

Tasks are executed sequentially, and if any task fails, subsequent tasks are not started.

## Database Schema

### analysis_tasks Table

```sql
CREATE TABLE analysis_tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    skill_name TEXT NOT NULL,
    task_type TEXT NOT NULL,  -- INCREMENTAL, FULL_RECOMPUTE
    status TEXT NOT NULL,      -- pending, running, completed, failed
    progress_percent INTEGER DEFAULT 0,
    eta_seconds INTEGER,
    params_json TEXT,
    threshold_profile_id INTEGER,
    total_points INTEGER,
    processed_points INTEGER DEFAULT 0,
    failed_points INTEGER DEFAULT 0,
    start_time INTEGER,
    end_time INTEGER,
    result_summary TEXT,
    error_message TEXT,
    depends_on_task_ids TEXT,
    blocks_task_ids TEXT,
    created_by TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Performance Considerations

- **Batch Size**: Default 1000 points per batch. Adjust based on memory constraints.
- **Progress Updates**: Updated after each batch to minimize database writes.
- **ETA Calculation**: Based on average processing speed of completed batches.
- **Memory Usage**: Python workers typically use 200-500MB depending on analysis complexity.

## Error Handling

- **Worker Failures**: Automatically marked as failed in database
- **Database Errors**: Logged and task marked as failed
- **Docker Errors**: Logged in Go backend, task marked as failed
- **Partial Failures**: Individual point failures tracked in `failed_points` field

## Monitoring

### Check Task Status

```bash
# List all tasks
curl http://localhost:8080/api/admin/analysis/tasks

# List running tasks
curl http://localhost:8080/api/admin/analysis/tasks?status=running

# List failed tasks
curl http://localhost:8080/api/admin/analysis/tasks?status=failed
```

### Docker Logs

```bash
# View worker logs
docker logs <container_id>

# Follow worker logs in real-time
docker logs -f <container_id>
```

## Next Steps

1. Implement Phase 4.1: Foundation Layer
   - outlier_detection worker
   - trajectory_completion worker

2. Implement Phase 4.2: Behavior Layer
   - transport_mode worker
   - stay_detection worker
   - trip_construction worker

3. Implement Phase 4.3: Spatial Analysis Layer
   - grid_system worker
   - footprint_stats worker
   - stay_stats worker

## References

- [Database Schema Design](../docs/database-schema-design.md)
- [30 Skills Documentation](../../.claude/skills/tracks/README.md)
- [Implementation Status](../docs/tracks/implementation-status.md)
