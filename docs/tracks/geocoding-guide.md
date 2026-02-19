# Geocoding Service Requirements

## Python Dependencies

```bash
pip install geopandas shapely pyproj pandas
```

## Shapefile Data Sources

### Option 1: National Geomatics Center of China (NGCC)
- Official administrative boundary data
- High accuracy
- Requires registration and approval

### Option 2: OpenStreetMap (Recommended for development)
- Free and open source
- Good accuracy for most regions
- Download from: https://download.geofabrik.de/asia/china.html

### Option 3: GADM (Global Administrative Areas)
- Free for non-commercial use
- Download from: https://gadm.org/download_country.html
- Select China and download shapefile format

## Required Administrative Levels

1. Province (省级) - Level 1
2. City (市级) - Level 2
3. County (区县级) - Level 3
4. Town (乡镇级) - Level 4
5. Village (村级/街道级) - Level 5

## Shapefile Structure

Expected files for each level:
- `china_province.shp` (省级)
- `china_city.shp` (市级)
- `china_county.shp` (区县级)
- `china_town.shp` (乡镇级)
- `china_village.shp` (村级)

Each shapefile should have:
- Geometry column (POLYGON or MULTIPOLYGON)
- Name column (Chinese name of the administrative division)
- Code column (administrative division code)

## Setup Instructions

1. Download shapefile data from one of the sources above
2. Extract to `go-backend/data/shapefiles/`
3. Verify shapefile structure matches expected format
4. Run geocoding script: `python geocode.py`

## Performance Expectations

- 100k points in ~2 minutes (with spatial index)
- Memory usage: <500MB
- Batch size: 1000 points per batch
