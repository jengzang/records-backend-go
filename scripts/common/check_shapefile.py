#!/usr/bin/env python3
"""
Check shapefile structure and fields.
"""

import sys
from pathlib import Path

try:
    import geopandas as gpd
except ImportError:
    print("Error: geopandas not installed")
    print("Please install: pip install geopandas")
    sys.exit(1)

# Path to shapefile
shp_path = Path(__file__).parent.parent / "data" / "geo" / "2024全国乡镇边界" / "2024全国乡镇边界.shp"

if not shp_path.exists():
    print(f"Error: Shapefile not found at {shp_path}")
    sys.exit(1)

print(f"Reading shapefile: {shp_path}")
gdf = gpd.read_file(shp_path)

print(f"\nTotal features: {len(gdf)}")
print(f"\nColumns: {list(gdf.columns)}")
print(f"\nData types:")
print(gdf.dtypes)

print(f"\nFirst 5 rows:")
print(gdf.head())

print(f"\nSample data:")
for i in range(min(5, len(gdf))):
    print(f"\nRow {i}:")
    for col in gdf.columns:
        if col != 'geometry':
            print(f"  {col}: {gdf.iloc[i][col]}")
