# Quick Start Guide: Geocoding Service

## Overview

The geocoding service converts GPS coordinates to administrative divisions (province/city/county/town) using Docker containers that run on-demand.

## Architecture

- **Go Backend:** Always running, manages tasks and triggers Python containers
- **Python Worker:** Runs in Docker only when geocoding is needed
- **SQLite Database:** Stores track points and task status

## Setup

### 1. Build Docker Image

```bash
cd go-backend
docker build -f Dockerfile.geocoding -t records-geocoding:latest .
```

### 2. Apply Database Migration

```bash
cd scripts/tracks
python run_migration.py
```

This creates the `geocoding_tasks` table.

### 3. Start Go Backend

```bash
# Development
go run cmd/server/main.go

# Production (Docker)
docker-compose up -d go-backend
```

## Usage

### Create Geocoding Task (API)

```bash
# Create task
curl -X POST http://localhost:8080/api/v1/admin/geocoding/tasks

# Response:
{
  "code": 200,
  "data": {
    "id": 1,
    "status": "pending",
    "total_points": 12345,
    "processed_points": 0,
    "failed_points": 0,
    "created_by": "admin",
    "created_at": "2026-02-19T10:00:00Z"
  }
}
```

### Check Task Status

```bash
curl http://localhost:8080/api/v1/admin/geocoding/tasks/1

# Response:
{
  "code": 200,
  "data": {
    "id": 1,
    "status": "running",
    "total_points": 12345,
    "processed_points": 8024,
    "failed_points": 23,
    "eta_seconds": 272,
    "start_time": "2026-02-19T10:00:05Z",
    "updated_at": "2026-02-19T10:02:15Z"
  }
}
```

### List All Tasks

```bash
curl "http://localhost:8080/api/v1/admin/geocoding/tasks?status=completed&limit=10"
```

### Cancel Task

```bash
curl -X DELETE http://localhost:8080/api/v1/admin/geocoding/tasks/1
```

## Manual Testing

### Test Python Worker Directly

```bash
# Windows paths
docker run --rm \
  -v C:/Users/joengzaang/CodeProject/records/go-backend/data/tracks:/data \
  -v C:/Users/joengzaang/CodeProject/records/go-backend/data/geo:/geo \
  records-geocoding:latest \
  python /app/geocode_worker.py --task-id 1 --batch-size 1000

# Linux paths
docker run --rm \
  -v /path/to/data/tracks:/data \
  -v /path/to/data/geo:/geo \
  records-geocoding:latest \
  python /app/geocode_worker.py --task-id 1 --batch-size 1000
```

## Task Status Flow

```
pending → running → completed
                 → failed
```

- **pending:** Task created, waiting to start
- **running:** Python container is processing points
- **completed:** All points processed successfully
- **failed:** Error occurred during processing

## Performance

- **Speed:** 1,000-1,500 points/sec (with GeoHash cache)
- **Memory:** ~1.1GB during processing (Go + Python + SQLite)
- **Cache:** Persistent across runs, 80%+ hit rate expected
- **Batch Size:** 1000 points (configurable)

## Monitoring

### Check Container Status

```bash
# List running containers
docker ps

# View container logs
docker logs <container_id>
```

### Check Database

```bash
sqlite3 data/tracks/tracks.db

# Count ungeocoded points
SELECT COUNT(*) FROM "一生足迹" WHERE province IS NULL;

# Check task status
SELECT * FROM geocoding_tasks ORDER BY created_at DESC LIMIT 5;

# Check geocoding results
SELECT province, city, county, town, COUNT(*) as count
FROM "一生足迹"
WHERE province IS NOT NULL
GROUP BY province, city, county, town
ORDER BY count DESC
LIMIT 10;
```

## Troubleshooting

### Container Won't Start

```bash
# Check if image exists
docker images | grep records-geocoding

# Rebuild image
docker build -f Dockerfile.geocoding -t records-geocoding:latest .
```

### Task Stuck in "pending"

- Check Go backend logs for Docker errors
- Verify Docker is running
- Check volume paths are correct

### Low Success Rate

- Verify shapefile exists: `data/geo/2024全国乡镇边界/2024全国乡镇边界.shp`
- Check shapefile integrity with `scripts/common/check_shapefile.py`
- Review failed points with `scripts/geocoding/verify_geocoding.py`

### Out of Memory

- Reduce batch size: `--batch-size 500`
- Check available memory: `docker stats`
- Ensure swap is configured on server

## Configuration

### Batch Size

Adjust in `internal/service/geocoding_service.go`:

```go
cmd := exec.Command("docker", "run", "--rm",
    // ... volumes ...
    "python", "/app/geocode_worker.py",
    "--task-id", strconv.Itoa(taskID),
    "--batch-size", "500")  // Change here
```

### Cache Precision

Adjust in `scripts/geocoding/geocode_worker.py`:

```python
cache = GeoHashCache(precision=6)  # 6 = ~1.2km x 0.6km
```

Higher precision = more cache entries, lower hit rate
Lower precision = fewer cache entries, higher hit rate

## Next Steps

1. **Frontend Development:** Create UI for task management
2. **Deployment:** Deploy to production server
3. **Monitoring:** Set up alerts for failed tasks
4. **Optimization:** Tune batch size and cache precision

## Support

- **Issues:** Check `IMPLEMENTATION_SUMMARY.md` for known issues
- **Logs:** Go backend logs + Docker container logs
- **Database:** Query `geocoding_tasks` table for task history
