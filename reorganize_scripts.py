#!/usr/bin/env python3
"""
Script to reorganize Python scripts according to the new structure.
"""
import os
import shutil
from pathlib import Path

# Base directory
BASE_DIR = Path(__file__).parent
SCRIPTS_DIR = BASE_DIR / "scripts"

# Define the reorganization plan
MOVES = [
    # Move check scripts to common
    ("check_dbf.py", "scripts/common/check_dbf.py"),
    ("check_shapefile.py", "scripts/common/check_shapefile.py"),
    ("check_schema.py", "scripts/common/check_schema.py"),

    # Move geocoding scripts
    ("scripts/tracks/geocode.py", "scripts/geocoding/geocode.py"),
    ("scripts/tracks/check_geocoded.py", "scripts/geocoding/check_geocoded.py"),
    ("scripts/tracks/verify_geocoding.py", "scripts/geocoding/verify_geocoding.py"),
    ("scripts/tracks/inspect_shapefile.py", "scripts/geocoding/inspect_shapefile.py"),

    # Move tracks import scripts
    ("scripts/tracks/write2sql.py", "scripts/tracks/import/write2sql.py"),

    # Move keyboard scripts
    ("scripts/keyboard/dataprocessing.py", "scripts/keyboard/import/ini_parser.py"),
    ("scripts/keyboard/ini_to_sqlite.py", "scripts/keyboard/import/ini_to_sqlite.py"),
    ("scripts/keyboard/verify_database.py", "scripts/keyboard/import/verify_database.py"),
    ("scripts/keyboard/photos.py", "scripts/keyboard/analysis/charts.py"),
    ("scripts/keyboard/frequency2.py", "scripts/keyboard/analysis/frequency.py"),
]

# Files to delete (duplicates)
DELETE_FILES = [
    "scripts/tracks/process_tracks/stop2.py",
    "scripts/tracks/process_tracks/stop_old.py",
    "scripts/tracks/process_tracks/test.py",
    "scripts/keyboard/frequecy.py",  # typo version
]

def create_directories():
    """Create the new directory structure."""
    dirs = [
        "scripts/common",
        "scripts/geocoding",
        "scripts/tracks/import",
        "scripts/tracks/analysis",
        "scripts/tracks/migrations",
        "scripts/keyboard/import",
        "scripts/keyboard/analysis",
    ]

    for dir_path in dirs:
        full_path = BASE_DIR / dir_path
        full_path.mkdir(parents=True, exist_ok=True)
        print(f"[OK] Created directory: {dir_path}")

def move_files():
    """Move files according to the plan."""
    for src, dst in MOVES:
        src_path = BASE_DIR / src
        dst_path = BASE_DIR / dst

        if src_path.exists():
            # Ensure destination directory exists
            dst_path.parent.mkdir(parents=True, exist_ok=True)

            # Move the file
            shutil.move(str(src_path), str(dst_path))
            print(f"[OK] Moved: {src} -> {dst}")
        else:
            print(f"[WARN] Source not found: {src}")

def delete_duplicates():
    """Delete duplicate files."""
    for file_path in DELETE_FILES:
        full_path = BASE_DIR / file_path
        if full_path.exists():
            full_path.unlink()
            print(f"[OK] Deleted: {file_path}")
        else:
            print(f"[WARN] Not found (already deleted?): {file_path}")

def main():
    print("=" * 60)
    print("Python Scripts Reorganization")
    print("=" * 60)

    print("\n[1/3] Creating new directory structure...")
    create_directories()

    print("\n[2/3] Moving files...")
    move_files()

    print("\n[3/3] Deleting duplicate files...")
    delete_duplicates()

    print("\n" + "=" * 60)
    print("Reorganization complete!")
    print("=" * 60)

    print("\nNext steps:")
    print("1. Merge stop*.py files into stay_detection.py")
    print("2. Merge tracks*.py files into statistics.py")
    print("3. Update import paths in all scripts")
    print("4. Test all scripts")

if __name__ == "__main__":
    main()
