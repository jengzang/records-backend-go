-- Flight data schema for TrainPlane module
-- Stores flight tracking data from Variflight/FlightAware

CREATE TABLE IF NOT EXISTS flights (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  flight_number TEXT NOT NULL,           -- e.g., CA1332, KN5290
  aircraft_number TEXT,                  -- Aircraft registration
  airline TEXT,                          -- Airline name
  departure_airport TEXT,                -- IATA code
  arrival_airport TEXT,                  -- IATA code
  departure_time TIMESTAMP,              -- Scheduled departure
  arrival_time TIMESTAMP,                -- Scheduled arrival
  actual_departure TIMESTAMP,            -- Actual departure
  actual_arrival TIMESTAMP,              -- Actual arrival
  flight_date TEXT NOT NULL,             -- YYYYMMDD
  data_source TEXT,                      -- variflight, flightaware
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(flight_number, flight_date)
);

CREATE INDEX idx_flights_number ON flights(flight_number);
CREATE INDEX idx_flights_date ON flights(flight_date);
CREATE INDEX idx_flights_airline ON flights(airline);

-- Flight tracking points (GPS coordinates during flight)
CREATE TABLE IF NOT EXISTS flight_points (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  flight_id INTEGER NOT NULL,
  update_time INTEGER NOT NULL,         -- Unix timestamp (seconds)
  utc_time TEXT,                        -- UTC time string
  longitude REAL NOT NULL,
  latitude REAL NOT NULL,
  altitude REAL,                        -- meters
  speed REAL,                           -- km/h
  heading REAL,                         -- degrees (0-360)
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (flight_id) REFERENCES flights(id) ON DELETE CASCADE
);

CREATE INDEX idx_flight_points_flight ON flight_points(flight_id);
CREATE INDEX idx_flight_points_time ON flight_points(update_time);
CREATE INDEX idx_flight_points_location ON flight_points(longitude, latitude);

-- Flight statistics (pre-computed)
CREATE TABLE IF NOT EXISTS flight_statistics (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  flight_id INTEGER NOT NULL UNIQUE,
  total_distance REAL,                  -- km
  max_altitude REAL,                    -- meters
  max_speed REAL,                       -- km/h
  avg_speed REAL,                       -- km/h
  duration_minutes INTEGER,             -- actual flight duration
  point_count INTEGER,                  -- number of tracking points
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (flight_id) REFERENCES flights(id) ON DELETE CASCADE
);

CREATE INDEX idx_flight_stats_flight ON flight_statistics(flight_id);

-- Flight-trajectory matching (links flights to GPS tracks)
CREATE TABLE IF NOT EXISTS flight_track_matches (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  flight_id INTEGER NOT NULL,
  track_point_id INTEGER NOT NULL,      -- Reference to 一生足迹.id
  match_confidence REAL,                -- 0.0-1.0
  time_diff_seconds INTEGER,            -- Time difference between flight point and track point
  distance_meters REAL,                 -- Spatial distance
  match_type TEXT,                      -- exact, interpolated, estimated
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (flight_id) REFERENCES flights(id) ON DELETE CASCADE
);

CREATE INDEX idx_flight_matches_flight ON flight_track_matches(flight_id);
CREATE INDEX idx_flight_matches_track ON flight_track_matches(track_point_id);
CREATE INDEX idx_flight_matches_confidence ON flight_track_matches(match_confidence DESC);
