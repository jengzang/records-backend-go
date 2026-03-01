-- Railway routes database schema

-- Railway lines table (铁路线路)
CREATE TABLE IF NOT EXISTS railway_lines (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  line_name TEXT NOT NULL,              -- 线路名称 (如: 京广线)
  line_code TEXT,                       -- 线路代码
  line_type TEXT NOT NULL,              -- 线路类型 (高速/普速/城际)
  total_distance REAL,                  -- 总里程 (km)
  start_station TEXT,                   -- 起点站
  end_station TEXT,                     -- 终点站
  opened_date TEXT,                     -- 开通日期
  max_speed INTEGER,                    -- 最高速度 (km/h)
  source TEXT DEFAULT 'openrailmap',    -- 数据来源
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Railway segments table (线路段)
CREATE TABLE IF NOT EXISTS railway_segments (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  line_id INTEGER NOT NULL,
  segment_name TEXT,                    -- 段名称 (如: 广州南-清远)
  start_station TEXT NOT NULL,
  end_station TEXT NOT NULL,
  distance REAL,                        -- 段距离 (km)
  sequence INTEGER NOT NULL,            -- 顺序号
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (line_id) REFERENCES railway_lines(id) ON DELETE CASCADE
);

-- Railway points table (线路坐标点)
CREATE TABLE IF NOT EXISTS railway_points (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  segment_id INTEGER NOT NULL,
  longitude REAL NOT NULL,
  latitude REAL NOT NULL,
  altitude REAL,
  sequence INTEGER NOT NULL,            -- 点序号
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (segment_id) REFERENCES railway_segments(id) ON DELETE CASCADE
);

-- Railway trips table (乘车记录)
CREATE TABLE IF NOT EXISTS railway_trips (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  train_number TEXT NOT NULL,           -- 车次 (如: G1234)
  line_id INTEGER,                      -- 关联线路
  departure_station TEXT NOT NULL,      -- 出发站
  arrival_station TEXT NOT NULL,        -- 到达站
  departure_time TIMESTAMP NOT NULL,    -- 出发时间
  arrival_time TIMESTAMP NOT NULL,      -- 到达时间
  duration_minutes INTEGER,             -- 行程时长 (分钟)
  distance REAL,                        -- 行程距离 (km)
  seat_type TEXT,                       -- 座位类型
  ticket_price REAL,                    -- 票价
  notes TEXT,                           -- 备注
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (line_id) REFERENCES railway_lines(id) ON DELETE SET NULL
);

-- Railway statistics table (统计数据)
CREATE TABLE IF NOT EXISTS railway_statistics (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  total_lines INTEGER DEFAULT 0,
  total_trips INTEGER DEFAULT 0,
  total_distance REAL DEFAULT 0,
  total_duration INTEGER DEFAULT 0,
  unique_trains INTEGER DEFAULT 0,
  date_range_start TEXT,
  date_range_end TEXT,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_railway_lines_name ON railway_lines(line_name);
CREATE INDEX IF NOT EXISTS idx_railway_lines_type ON railway_lines(line_type);
CREATE INDEX IF NOT EXISTS idx_railway_segments_line ON railway_segments(line_id);
CREATE INDEX IF NOT EXISTS idx_railway_points_segment ON railway_points(segment_id);
CREATE INDEX IF NOT EXISTS idx_railway_trips_train ON railway_trips(train_number);
CREATE INDEX IF NOT EXISTS idx_railway_trips_date ON railway_trips(departure_time);
CREATE INDEX IF NOT EXISTS idx_railway_trips_line ON railway_trips(line_id);

-- Enable WAL mode for better concurrency
PRAGMA journal_mode=WAL;
PRAGMA foreign_keys=ON;
