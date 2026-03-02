import sqlite3

# Keyboard data
conn = sqlite3.connect('data/keyboard/kmcounter.db')
cursor = conn.cursor()
cursor.execute('SELECT * FROM keyboard_data ORDER BY date DESC LIMIT 3')
print('Keyboard data (date, keystrokes):')
for row in cursor.fetchall():
    print(row)
conn.close()

print('\n')

# Screentime data
conn = sqlite3.connect('data/screentime/screentime.db')
cursor = conn.cursor()
cursor.execute('SELECT date, app_name, package_id, duration_ms, launch_count FROM screentime_daily ORDER BY date DESC LIMIT 5')
print('Screentime data (date, app_name, package_id, duration_ms, launch_count):')
for row in cursor.fetchall():
    print(row)
conn.close()

print('\n')

# Health data
conn = sqlite3.connect('data/applehealth/health.db')
cursor = conn.cursor()
cursor.execute("SELECT type, value, unit, start_date FROM health_records WHERE type LIKE '%HeartRate%' ORDER BY start_date DESC LIMIT 5")
print('Health data (type, value, unit, start_date):')
for row in cursor.fetchall():
    print(row)
conn.close()
