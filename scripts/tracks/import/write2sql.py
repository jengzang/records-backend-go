import os
import re
import sqlite3
import tkinter as tk
from tkinter import filedialog
from typing import Iterable, Optional, Tuple, List
from datetime import datetime

import pandas as pd

def _sanitize_identifier(name: str) -> str:
    if name is None:
        return "unnamed"
    s = str(name).strip()
    s = re.sub(r"\s+", "_", s)
    s = re.sub(r"[^\w\u4e00-\u9fff]+", "_", s)
    s = s.strip("_")
    return s or "unnamed"

def _detect_sqlite_type(series: pd.Series) -> str:
    if pd.api.types.is_bool_dtype(series):
        return "INTEGER"
    if pd.api.types.is_integer_dtype(series):
        return "INTEGER"
    if pd.api.types.is_float_dtype(series):
        return "REAL"
    return "TEXT"

def import_excel_sheet_columns_to_sqlite_via_tk(
    db_path: str,
    table_name: str,
    sheet_name: Optional[str] = None,
    if_exists: str = "replace",
) -> Tuple[str, str, List[str]]:
    
    # --- 1. 定義指定的列名 ---
    target_cols = [
        "dataTime", "longitude", "latitude", "heading", 
        "accuracy", "speed", "distance", "altitude"
    ]

    # --- 2. Tk 選擇文件 ---
    root = tk.Tk()
    root.withdraw()
    excel_path = filedialog.askopenfilename(
        title="選擇 Excel 文件",
        filetypes=[("Excel files", "*.xlsx *.xlsm *.xls"), ("All files", "*.*")]
    )
    root.destroy()

    if not excel_path:
        raise RuntimeError("未選擇 Excel 文件，已取消。")

    if sheet_name is None:
        xls = pd.ExcelFile(excel_path)
        print("檢測到的 sheets：")
        for s in xls.sheet_names: print(f"  - {s}")
        sheet_name = input("請輸入要導入的 sheet 名：").strip()

    # --- 3. 讀取與過濾數據 ---
    # 讀取時確保包含 stepType 用於過濾，以及 target_cols 用於提取
    df = pd.read_excel(excel_path, sheet_name=sheet_name, engine="openpyxl")

    # A. 只保留 stepType 為 0 的數據
    if "stepType" in df.columns:
        # 強制轉換類型以防萬一（處理 0.0 或 "0" 的情況）
        df = df[pd.to_numeric(df["stepType"], errors='coerce') == 0].copy()
    else:
        print("[Warning] 未在 Excel 中找到 'stepType' 列，跳過過濾步驟。")

    # B. 只提取你指定的那些列（如果 Excel 裡有這些列的話）
    existing_target_cols = [c for c in target_cols if c in df.columns]
    df = df[existing_target_cols].copy()

    # --- 4. 計算新列：time_visually 和 time ---
    if "dataTime" in df.columns:
        def convert_time(ts):
            try:
                # 假設 dataTime 是秒級時間戳 (10位數)
                dt = datetime.fromtimestamp(float(ts))
                # 格式: 2025/01/22 21:42:18.000
                v = dt.strftime("%Y/%m/%d %H:%M:%S.000")
                # 格式: 20250122214218
                t = dt.strftime("%Y%m%d%H%M%S")
                return v, t
            except:
                return None, None

        # 套用轉換
        time_results = df["dataTime"].apply(convert_time)
        df["time_visually"] = time_results.apply(lambda x: x[0])
        df["time"] = time_results.apply(lambda x: x[1])

    # --- 5. 數據清理與準備寫入 ---
    # 清理列名（防止非法字符）
    df.columns = [_sanitize_identifier(c) for c in df.columns]
    
    # 處理 NaN
    df2 = df.where(pd.notnull(df), None)

    # --- 6. 寫入 SQLite ---
    os.makedirs(os.path.dirname(db_path) or ".", exist_ok=True)
    conn = sqlite3.connect(db_path)
    conn.execute("PRAGMA journal_mode=WAL")
    
    try:
        cur = conn.cursor()
        if if_exists == "replace":
            cur.execute(f'DROP TABLE IF EXISTS "{table_name}"')

        # 【改動 1】建表時顯式加入 id 主鍵
        col_defs = ['"id" INTEGER PRIMARY KEY AUTOINCREMENT'] # 這裡是新增的主鍵
        for col in df2.columns:
            sql_type = _detect_sqlite_type(df2[col])
            col_defs.append(f'"{col}" {sql_type}')
        
        create_sql = f'CREATE TABLE IF NOT EXISTS "{table_name}" ({", ".join(col_defs)})'
        cur.execute(create_sql)

        # 【改動 2】插入時不提供 id，SQLite 會自動從 1 開始順序創建
        placeholders = ", ".join(["?"] * len(df2.columns))
        col_list = ", ".join([f'"{c}"' for c in df2.columns])
        insert_sql = f'INSERT INTO "{table_name}" ({col_list}) VALUES ({placeholders})'
        
        cur.executemany(insert_sql, df2.values.tolist())
        conn.commit()
        
    finally:
        conn.close()

    return excel_path, sheet_name, list(df2.columns)

# --- 執行 ---
db_file = "data/tracks.db"
target_table = "一生足迹"

excel_path, used_sheet, inserted_cols = import_excel_sheet_columns_to_sqlite_via_tk(
    db_path=db_file,
    table_name=target_table,
    sheet_name="processing",
    if_exists="replace",
)

print("-" * 30)
print(f"Excel 路徑: {excel_path}")
print(f"使用的 Sheet: {used_sheet}")
print(f"成功寫入列: {inserted_cols}")