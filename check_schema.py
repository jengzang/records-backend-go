import sqlite3

db_path = r'C:\Users\joengzaang\CodeProject\records\go-backend\data\tracks\tracks.db'
conn = sqlite3.connect(db_path)
cursor = conn.cursor()

# Get all tables
cursor.execute("SELECT name FROM sqlite_master WHERE type='table'")
tables = cursor.fetchall()

print("Tables in database:")
for table in tables:
    print(f"\n{table[0]}:")
    cursor.execute(f"SELECT sql FROM sqlite_master WHERE type='table' AND name='{table[0]}'")
    schema = cursor.fetchone()
    if schema and schema[0]:
        print(schema[0])

    # Get sample row count
    cursor.execute(f"SELECT COUNT(*) FROM \"{table[0]}\"")
    count = cursor.fetchone()[0]
    print(f"Row count: {count}")

conn.close()
