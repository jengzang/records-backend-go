-- Migration 022: Add stay detection threshold profiles
-- Purpose: Add preset profiles for SPATIAL and ADMIN_AREA stay detection
-- Date: 2026-02-23

-- Insert SPATIAL stay detection profiles
INSERT OR IGNORE INTO threshold_profiles (name, params_json, description) VALUES
('spatial_strict', '{
  "min_duration_s": 7200,
  "spatial_eps_m": 100,
  "min_samples": 10,
  "max_time_gap_s": 900
}', '严格空间停留: 2小时+, 100米, 10点+, 15分钟间隔'),

('spatial_relaxed', '{
  "min_duration_s": 1800,
  "spatial_eps_m": 200,
  "min_samples": 5,
  "max_time_gap_s": 1800
}', '宽松空间停留: 30分钟+, 200米, 5点+, 30分钟间隔'),

('spatial_default', '{
  "min_duration_s": 7200,
  "spatial_eps_m": 100,
  "min_samples": 10,
  "max_time_gap_s": 900
}', '默认空间停留: 2小时+, 100米, 10点+, 15分钟间隔 (与strict相同)');

-- Insert ADMIN_AREA stay detection profiles
INSERT OR IGNORE INTO threshold_profiles (name, params_json, description) VALUES
('admin_city', '{
  "admin_level": "city",
  "min_duration_s": 14400,
  "merge_gap_s": 1800,
  "min_points": 3
}', '城市级行政区停留: 4小时+, 30分钟合并间隔'),

('admin_county', '{
  "admin_level": "county",
  "min_duration_s": 7200,
  "merge_gap_s": 1800,
  "min_points": 3
}', '区县级行政区停留: 2小时+, 30分钟合并间隔'),

('admin_town', '{
  "admin_level": "town",
  "min_duration_s": 3600,
  "merge_gap_s": 900,
  "min_points": 3
}', '乡镇级行政区停留: 1小时+, 15分钟合并间隔'),

('admin_province', '{
  "admin_level": "province",
  "min_duration_s": 28800,
  "merge_gap_s": 3600,
  "min_points": 5
}', '省级行政区停留: 8小时+, 1小时合并间隔');
