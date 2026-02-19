#!/usr/bin/env python3
"""
Quick script to inspect the shapefile structure.
"""
import geopandas as gpd
from pathlib import Path

shapefile_path = Path(__file__).parent.parent.parent / "data" / "geo" / "2024全国乡镇边界" / "2024全国乡镇边界.shp"

print(f"Loading shapefile: {shapefile_path}")
gdf = gpd.read_file(shapefile_path)

print(f"\nTotal features: {len(gdf)}")
print(f"\nColumns: {list(gdf.columns)}")
print(f"\nColumn dtypes:")
print(gdf.dtypes)

print(f"\nFirst 3 records:")
for idx in range(min(3, len(gdf))):
    print(f"\nRecord {idx}:")
    for col in gdf.columns:
        if col != 'geometry':
            print(f"  {col}: {gdf.iloc[idx][col]}")
