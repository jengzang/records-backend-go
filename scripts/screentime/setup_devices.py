#!/usr/bin/env python3
"""
Setup devices database and rename existing databases
"""

import sqlite3
import os
import shutil
from pathlib import Path

# Paths
DATA_DIR = Path(__file__).parent.parent.parent / "data" / "screentime"
DEVICES_DB = DATA_DIR / "devices.db"
OLD_PHONE_DB = DATA_DIR / "screentime.db"
NEW_PHONE_DB = DATA_DIR / "phone_vivo_x90.db"
COMPUTER_DB = DATA_DIR / "manictime_computer.db"

def create_devices_db():
    """Create devices database with schema"""
    print(f"Creating devices database at {DEVICES_DB}")

    conn = sqlite3.connect(DEVICES_DB)
    cursor = conn.cursor()

    # Create devices table
    cursor.execute("""
    CREATE TABLE IF NOT EXISTS devices (
      id TEXT PRIMARY KEY,
      name TEXT NOT NULL,
      type TEXT NOT NULL,
      db_path TEXT NOT NULL,
      data_format TEXT NOT NULL,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
      last_sync TIMESTAMP,
      is_active BOOLEAN DEFAULT 1,
      total_records INTEGER DEFAULT 0,
      date_range_start TEXT,
      date_range_end TEXT,
      metadata TEXT
    )
    """)

    # Create indexes
    cursor.execute("CREATE INDEX IF NOT EXISTS idx_devices_type ON devices(type)")
    cursor.execute("CREATE INDEX IF NOT EXISTS idx_devices_active ON devices(is_active)")

    # Insert initial devices
    cursor.execute("""
    INSERT OR REPLACE INTO devices
    (id, name, type, db_path, data_format, is_active, total_records, date_range_start, date_range_end)
    VALUES
      ('phone_vivo_x90', 'Vivo X90', 'phone', 'phone_vivo_x90.db', 'phone_txt', 1, 22013, '20231113', '20260219'),
      ('computer_main', '主电脑', 'computer', 'manictime_computer.db', 'manictime_excel', 1, 523965, '20221124', '20260224')
    """)

    conn.commit()
    conn.close()

    print("[OK] Devices database created successfully")

def rename_phone_db():
    """Rename phone database file"""
    if OLD_PHONE_DB.exists() and not NEW_PHONE_DB.exists():
        print(f"Renaming {OLD_PHONE_DB.name} to {NEW_PHONE_DB.name}")

        # Close any open connections by copying instead of moving
        shutil.copy2(OLD_PHONE_DB, NEW_PHONE_DB)
        print(f"[OK] Phone database copied to {NEW_PHONE_DB.name}")
        print(f"  Note: Original file {OLD_PHONE_DB.name} still exists (can be deleted manually)")
    elif NEW_PHONE_DB.exists():
        print(f"[OK] Phone database already exists at {NEW_PHONE_DB.name}")
    else:
        print(f"[WARN] Warning: {OLD_PHONE_DB.name} not found")

def verify_databases():
    """Verify all databases exist and are accessible"""
    print("\nVerifying databases:")

    databases = [
        ("Devices DB", DEVICES_DB),
        ("Phone DB", NEW_PHONE_DB),
        ("Computer DB", COMPUTER_DB)
    ]

    for name, db_path in databases:
        if db_path.exists():
            size_mb = db_path.stat().st_size / (1024 * 1024)
            print(f"  [OK] {name}: {db_path.name} ({size_mb:.1f} MB)")

            # Test connection
            try:
                conn = sqlite3.connect(db_path)
                cursor = conn.cursor()

                if name == "Devices DB":
                    cursor.execute("SELECT COUNT(*) FROM devices")
                    count = cursor.fetchone()[0]
                    print(f"    - {count} devices registered")
                elif name == "Phone DB":
                    cursor.execute("SELECT COUNT(*) FROM screentime_daily")
                    count = cursor.fetchone()[0]
                    print(f"    - {count} daily records")
                elif name == "Computer DB":
                    cursor.execute("SELECT COUNT(*) FROM manictime_activities")
                    count = cursor.fetchone()[0]
                    print(f"    - {count} activity records")

                conn.close()
            except Exception as e:
                print(f"    [ERROR] Error accessing database: {e}")
        else:
            print(f"  [ERROR] {name}: Not found at {db_path}")

def main():
    print("=" * 60)
    print("Screentime Multi-Device Setup")
    print("=" * 60)
    print()

    # Create devices database
    create_devices_db()
    print()

    # Rename phone database
    rename_phone_db()
    print()

    # Verify all databases
    verify_databases()
    print()

    print("=" * 60)
    print("Setup complete!")
    print("=" * 60)

if __name__ == "__main__":
    main()
