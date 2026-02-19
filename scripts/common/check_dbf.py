#!/usr/bin/env python3
"""
Check shapefile DBF structure (attribute table).
"""

import struct
from pathlib import Path

# Path to DBF file
dbf_path = Path(__file__).parent / "data" / "geo" / "2024全国乡镇边界" / "2024全国乡镇边界.dbf"

if not dbf_path.exists():
    print(f"Error: DBF file not found at {dbf_path}")
    exit(1)

print(f"Reading DBF file: {dbf_path}")

with open(dbf_path, 'rb') as f:
    # Read header
    version = struct.unpack('B', f.read(1))[0]
    year, month, day = struct.unpack('BBB', f.read(3))
    num_records = struct.unpack('<I', f.read(4))[0]
    header_length = struct.unpack('<H', f.read(2))[0]
    record_length = struct.unpack('<H', f.read(2))[0]

    print(f"\nDBF Info:")
    print(f"  Version: {version}")
    print(f"  Last update: {year + 1900}-{month:02d}-{day:02d}")
    print(f"  Number of records: {num_records}")
    print(f"  Header length: {header_length}")
    print(f"  Record length: {record_length}")

    # Skip reserved bytes
    f.read(20)

    # Read field descriptors
    fields = []
    while True:
        field_info = f.read(32)
        if field_info[0] == 0x0D:  # End of field descriptors
            break

        field_name = field_info[:11].decode('gbk', errors='ignore').rstrip('\x00')
        field_type = chr(field_info[11])
        field_length = field_info[16]
        field_decimal = field_info[17]

        fields.append({
            'name': field_name,
            'type': field_type,
            'length': field_length,
            'decimal': field_decimal
        })

    print(f"\nFields ({len(fields)}):")
    for i, field in enumerate(fields, 1):
        print(f"  {i}. {field['name']:20s} Type: {field['type']} Length: {field['length']:3d} Decimal: {field['decimal']}")

    # Read first few records
    print(f"\nFirst 3 records:")
    for record_num in range(min(3, num_records)):
        f.read(1)  # Skip deletion flag
        record_data = f.read(record_length - 1)

        print(f"\nRecord {record_num + 1}:")
        offset = 0
        for field in fields:
            value = record_data[offset:offset + field['length']].decode('gbk', errors='ignore').strip()
            print(f"  {field['name']:20s}: {value}")
            offset += field['length']
