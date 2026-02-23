-- Keyboard/Mouse Analysis Module - Initial Schema
-- Created: 2026-02-23
-- Description: Database schema for keyboard and mouse usage tracking

-- Enable WAL mode for better concurrency
PRAGMA journal_mode=WAL;

-- Daily aggregated statistics
CREATE TABLE IF NOT EXISTS daily_stats (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  date TEXT NOT NULL UNIQUE,  -- YYYYMMDD format
  keystrokes INTEGER DEFAULT 0,
  left_clicks INTEGER DEFAULT 0,
  right_clicks INTEGER DEFAULT 0,
  middle_clicks INTEGER DEFAULT 0,
  extra_clicks INTEGER DEFAULT 0,
  wheel_scrolls INTEGER DEFAULT 0,
  h_wheel_scrolls INTEGER DEFAULT 0,
  mouse_distance_m REAL DEFAULT 0.0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_daily_stats_date ON daily_stats(date);

-- Per-scancode daily statistics
CREATE TABLE IF NOT EXISTS scancode_stats (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  date TEXT NOT NULL,
  scancode INTEGER NOT NULL,
  count INTEGER DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(date, scancode)
);

CREATE INDEX IF NOT EXISTS idx_scancode_date ON scancode_stats(date);
CREATE INDEX IF NOT EXISTS idx_scancode_code ON scancode_stats(scancode);

-- Scancode to key name mapping
CREATE TABLE IF NOT EXISTS scancode_mapping (
  scancode INTEGER PRIMARY KEY,
  key_name TEXT NOT NULL,
  key_category TEXT,  -- letter/number/function/modifier/special
  description TEXT
);

-- Hourly aggregated statistics (for detailed time analysis)
CREATE TABLE IF NOT EXISTS hourly_stats (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  date TEXT NOT NULL,
  hour INTEGER NOT NULL,  -- 0-23
  keystrokes INTEGER DEFAULT 0,
  mouse_clicks INTEGER DEFAULT 0,
  mouse_distance_m REAL DEFAULT 0.0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(date, hour)
);

CREATE INDEX IF NOT EXISTS idx_hourly_date ON hourly_stats(date);

-- Insert common scancode mappings
-- Based on standard Windows scancode values
INSERT OR IGNORE INTO scancode_mapping (scancode, key_name, key_category, description) VALUES
-- Letters
(30, 'A', 'letter', 'Letter A'),
(48, 'B', 'letter', 'Letter B'),
(46, 'C', 'letter', 'Letter C'),
(32, 'D', 'letter', 'Letter D'),
(18, 'E', 'letter', 'Letter E'),
(33, 'F', 'letter', 'Letter F'),
(34, 'G', 'letter', 'Letter G'),
(35, 'H', 'letter', 'Letter H'),
(23, 'I', 'letter', 'Letter I'),
(36, 'J', 'letter', 'Letter J'),
(37, 'K', 'letter', 'Letter K'),
(38, 'L', 'letter', 'Letter L'),
(50, 'M', 'letter', 'Letter M'),
(49, 'N', 'letter', 'Letter N'),
(24, 'O', 'letter', 'Letter O'),
(25, 'P', 'letter', 'Letter P'),
(16, 'Q', 'letter', 'Letter Q'),
(19, 'R', 'letter', 'Letter R'),
(31, 'S', 'letter', 'Letter S'),
(20, 'T', 'letter', 'Letter T'),
(22, 'U', 'letter', 'Letter U'),
(47, 'V', 'letter', 'Letter V'),
(17, 'W', 'letter', 'Letter W'),
(45, 'X', 'letter', 'Letter X'),
(21, 'Y', 'letter', 'Letter Y'),
(44, 'Z', 'letter', 'Letter Z'),
-- Numbers
(2, '1', 'number', 'Number 1'),
(3, '2', 'number', 'Number 2'),
(4, '3', 'number', 'Number 3'),
(5, '4', 'number', 'Number 4'),
(6, '5', 'number', 'Number 5'),
(7, '6', 'number', 'Number 6'),
(8, '7', 'number', 'Number 7'),
(9, '8', 'number', 'Number 8'),
(10, '9', 'number', 'Number 9'),
(11, '0', 'number', 'Number 0'),
-- Function keys
(59, 'F1', 'function', 'Function key F1'),
(60, 'F2', 'function', 'Function key F2'),
(61, 'F3', 'function', 'Function key F3'),
(62, 'F4', 'function', 'Function key F4'),
(63, 'F5', 'function', 'Function key F5'),
(64, 'F6', 'function', 'Function key F6'),
(65, 'F7', 'function', 'Function key F7'),
(66, 'F8', 'function', 'Function key F8'),
(67, 'F9', 'function', 'Function key F9'),
(68, 'F10', 'function', 'Function key F10'),
(87, 'F11', 'function', 'Function key F11'),
(88, 'F12', 'function', 'Function key F12'),
-- Modifiers
(42, 'Left Shift', 'modifier', 'Left Shift key'),
(54, 'Right Shift', 'modifier', 'Right Shift key'),
(29, 'Left Ctrl', 'modifier', 'Left Control key'),
(157, 'Right Ctrl', 'modifier', 'Right Control key'),
(56, 'Left Alt', 'modifier', 'Left Alt key'),
(184, 'Right Alt', 'modifier', 'Right Alt key'),
(91, 'Left Win', 'modifier', 'Left Windows key'),
(92, 'Right Win', 'modifier', 'Right Windows key'),
-- Special keys
(1, 'Esc', 'special', 'Escape key'),
(14, 'Backspace', 'special', 'Backspace key'),
(15, 'Tab', 'special', 'Tab key'),
(28, 'Enter', 'special', 'Enter key'),
(57, 'Space', 'special', 'Space bar'),
(58, 'Caps Lock', 'special', 'Caps Lock key'),
(69, 'Num Lock', 'special', 'Num Lock key'),
(70, 'Scroll Lock', 'special', 'Scroll Lock key'),
(83, 'Delete', 'special', 'Delete key'),
(82, 'Insert', 'special', 'Insert key'),
(71, 'Home', 'special', 'Home key'),
(79, 'End', 'special', 'End key'),
(73, 'Page Up', 'special', 'Page Up key'),
(81, 'Page Down', 'special', 'Page Down key'),
-- Arrow keys
(72, 'Up Arrow', 'special', 'Up arrow key'),
(80, 'Down Arrow', 'special', 'Down arrow key'),
(75, 'Left Arrow', 'special', 'Left arrow key'),
(77, 'Right Arrow', 'special', 'Right arrow key'),
-- Punctuation
(12, '-', 'special', 'Minus/Underscore'),
(13, '=', 'special', 'Equals/Plus'),
(26, '[', 'special', 'Left bracket'),
(27, ']', 'special', 'Right bracket'),
(39, ';', 'special', 'Semicolon/Colon'),
(40, '''', 'special', 'Apostrophe/Quote'),
(41, '`', 'special', 'Grave accent/Tilde'),
(43, '\', 'special', 'Backslash/Pipe'),
(51, ',', 'special', 'Comma/Less than'),
(52, '.', 'special', 'Period/Greater than'),
(53, '/', 'special', 'Slash/Question mark');
