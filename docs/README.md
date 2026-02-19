# Go Backend Documentation

## Overview

This is the backend service for the personal data analysis platform. It provides RESTful APIs for GPS trajectories, keyboard/mouse usage, screen time, and health data analysis.

**Current Status:** Planning phase. Python processing scripts exist, but Go backend is not yet implemented.

## Architecture

### Technology Stack

- **Language:** Go 1.21+
- **Web Framework:** Gin
- **Database:** SQLite with WAL mode
- **Authentication:** JWT (planned)
- **Deployment:** record.yzup.top

### Server Constraints

- **CPU:** 2 cores
- **RAM:** 2GB total, ≤1GB for Go backend
- **Request Rate:** ≤3 req/s
- **Storage:** ~500MB current, ~2GB projected (5 years)

### Design Principles

1. **Batch Processing:** Heavy computations run in background jobs
2. **Incremental Updates:** Avoid full recalculation
3. **Aggressive Caching:** Cache computed results
4. **Pagination:** Large result sets must be paginated
5. **Admin-Only Writes:** Read-only access for public, write access requires authentication

## Module Structure

### 1. Tracks Module (`/docs/tracks/`)

GPS trajectory analysis with 30 processing skills.

**Status:** Python scripts exist, Go API not implemented
**Priority:** P0 (Core module)
**Documentation:** See `tracks/README.md`

### 2. Keyboard Module (`/docs/keyboard/`)

Keyboard and mouse usage tracking and analysis.

**Status:** Python scripts and database exist, Go API not implemented
**Priority:** P1
**Documentation:** See `keyboard/README.md`

### 3. Screen Time Module (`/docs/screentime/`)

Phone app usage analysis.

**Status:** Data collection only, no processing yet
**Priority:** P2
**Documentation:** See `screentime/README.md`

### 4. Apple Health Module (`/docs/applehealth/`)

Health data and workout route analysis.

**Status:** Data collection only, no processing yet
**Priority:** P2
**Documentation:** See `applehealth/README.md`

### 5. Train/Plane Module (`/docs/trainplane/`)

Flight and train route visualization.

**Status:** Data collection only, no processing yet
**Priority:** P3
**Documentation:** See `trainplane/README.md`

## API Design

See `api-overview.md` for detailed API design principles.

### Base URL

```
https://record.yzup.top/api/v1
```

### Authentication

```
Authorization: Bearer <JWT_TOKEN>
```

### Response Format

```json
{
  "code": 0,
  "message": "success",
  "data": { ... },
  "timestamp": 1234567890
}
```

### Error Codes

- `0` - Success
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `429` - Too Many Requests
- `500` - Internal Server Error

## Database Design

See `database-design.md` for detailed database architecture.

### Database Files

- `data/tracks/tracks.db` - GPS trajectory data
- `data/keyboard/kmcounter.db` - Keyboard/mouse data
- `data/screentime/screentime.db` - Phone usage data (planned)
- `data/applehealth/health.db` - Health data (planned)
- `data/trainplane/routes.db` - Route data (planned)

### Common Patterns

- **WAL Mode:** All databases use `PRAGMA journal_mode=WAL`
- **Timestamps:** Unix timestamps (seconds) + formatted strings
- **Versioning:** `created_at`, `updated_at`, `algo_version` fields
- **Indexes:** On frequently queried fields

## Development

### Prerequisites

- Go 1.21+
- SQLite 3.35+
- Git

### Setup

```bash
cd go-backend
go mod download
go run main.go
```

### Testing

```bash
go test ./...
```

### Building

```bash
go build -o records-backend main.go
```

## Deployment

### Production

```bash
./records-backend --config production.yaml
```

### Environment Variables

- `DB_PATH` - Database directory path
- `JWT_SECRET` - JWT signing secret
- `PORT` - Server port (default: 8080)
- `LOG_LEVEL` - Logging level (debug/info/warn/error)

## Documentation Index

- `api-overview.md` - API design principles and conventions
- `database-design.md` - Database architecture and schemas
- `tracks/` - Tracks module documentation
- `keyboard/` - Keyboard module documentation
- `screentime/` - Screen time module documentation
- `applehealth/` - Apple Health module documentation
- `trainplane/` - Train/Plane module documentation
