-- Migration: Add administrative division columns and metadata fields
-- Date: 2026-02-19
-- Description: Adds province, city, county, town, village columns plus metadata fields to support geocoding

-- Add administrative division columns
ALTER TABLE "一生足迹" ADD COLUMN province TEXT;
ALTER TABLE "一生足迹" ADD COLUMN city TEXT;
ALTER TABLE "一生足迹" ADD COLUMN county TEXT;
ALTER TABLE "一生足迹" ADD COLUMN town TEXT;
ALTER TABLE "一生足迹" ADD COLUMN village TEXT;

-- Add metadata columns (SQLite doesn't support non-constant defaults in ALTER TABLE)
ALTER TABLE "一生足迹" ADD COLUMN created_at TEXT;
ALTER TABLE "一生足迹" ADD COLUMN updated_at TEXT;
ALTER TABLE "一生足迹" ADD COLUMN algo_version TEXT;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_datatime ON "一生足迹"(dataTime);
CREATE INDEX IF NOT EXISTS idx_time ON "一生足迹"(time);
CREATE INDEX IF NOT EXISTS idx_admin_province ON "一生足迹"(province);
CREATE INDEX IF NOT EXISTS idx_admin_city ON "一生足迹"(city);
CREATE INDEX IF NOT EXISTS idx_admin_county ON "一生足迹"(county);
CREATE INDEX IF NOT EXISTS idx_admin_full ON "一生足迹"(province, city, county);
CREATE INDEX IF NOT EXISTS idx_coordinates ON "一生足迹"(longitude, latitude);
