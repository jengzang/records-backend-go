-- Device registration table
CREATE TABLE IF NOT EXISTS devices (
  id TEXT PRIMARY KEY,              -- phone_vivo_x90, computer_main, etc.
  name TEXT NOT NULL,               -- "Vivo X90", "主电脑"
  type TEXT NOT NULL,               -- phone, computer
  db_path TEXT NOT NULL,            -- Relative path to database file
  data_format TEXT NOT NULL,        -- phone_txt, manictime_excel
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  last_sync TIMESTAMP,
  is_active BOOLEAN DEFAULT 1,
  total_records INTEGER DEFAULT 0,
  date_range_start TEXT,            -- YYYYMMDD
  date_range_end TEXT,              -- YYYYMMDD
  metadata TEXT                     -- JSON blob for additional info
);

-- Insert initial devices
INSERT INTO devices (id, name, type, db_path, data_format, is_active, total_records, date_range_start, date_range_end)
VALUES
  ('phone_vivo_x90', 'Vivo X90', 'phone', 'phone_vivo_x90.db', 'phone_txt', 1, 22013, '20231113', '20260219'),
  ('computer_main', '主电脑', 'computer', 'manictime_computer.db', 'manictime_excel', 1, 523965, '20221124', '20260224');

-- Create index on type for filtering
CREATE INDEX IF NOT EXISTS idx_devices_type ON devices(type);
CREATE INDEX IF NOT EXISTS idx_devices_active ON devices(is_active);
