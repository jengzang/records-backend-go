-- Device registration table for multi-source screentime data
-- This table manages all screentime data sources (phone, computer, etc.)

CREATE TABLE IF NOT EXISTS devices (
  id TEXT PRIMARY KEY,              -- vivo_x90, computer_main, etc.
  name TEXT NOT NULL,               -- "Vivo X90", "主电脑"
  type TEXT NOT NULL,               -- phone, computer
  db_path TEXT NOT NULL,            -- 数据库文件路径
  data_format TEXT NOT NULL,        -- phone_txt, manictime_excel
  is_active BOOLEAN DEFAULT 1,      -- 是否启用
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  last_sync TIMESTAMP,              -- 最后同步时间
  total_records INTEGER DEFAULT 0,  -- 总记录数
  date_range_start TEXT,            -- 数据起始日期
  date_range_end TEXT,              -- 数据结束日期
  metadata TEXT                     -- JSON格式的额外元数据
);

CREATE INDEX idx_devices_type ON devices(type);
CREATE INDEX idx_devices_active ON devices(is_active);

-- Insert default device (existing phone data)
INSERT OR IGNORE INTO devices (id, name, type, db_path, data_format, is_active)
VALUES ('vivo_x90', 'Vivo X90', 'phone', './data/screentime/phone_vivo_x90.db', 'phone_txt', 1);
