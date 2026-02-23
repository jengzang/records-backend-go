import os
import re
import sqlite3
import tkinter as tk
from tkinter import filedialog
from typing import Iterable, Optional, Tuple, List, Dict
from datetime import datetime
import argparse
import sys

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

def import_data_to_sqlite(
    db_path: str,
    table_name: str,
    file_path: Optional[str] = None,
    sheet_name: Optional[str] = None,
    if_exists: str = "append",
    deduplicate: bool = True,
    use_gui: bool = True,
) -> Dict[str, any]:

    # --- 1. 定義指定的列名 ---
    target_cols = [
        "dataTime", "longitude", "latitude", "heading",
        "accuracy", "speed", "distance", "altitude"
    ]

    # --- 2. 選擇文件 (GUI or CLI) ---
    if file_path is None and use_gui:
        root = tk.Tk()
        root.withdraw()
        file_path = filedialog.askopenfilename(
            title="選擇數據文件",
            filetypes=[
                ("Data files", "*.xlsx *.xlsm *.xls *.csv"),
                ("Excel files", "*.xlsx *.xlsm *.xls"),
                ("CSV files", "*.csv"),
                ("All files", "*.*")
            ]
        )
        root.destroy()

    if not file_path:
        raise RuntimeError("未選擇文件，已取消。")

    if not os.path.exists(file_path):
        raise RuntimeError(f"文件不存在: {file_path}")

    # --- 3. 讀取數據 (支持Excel和CSV) ---
    file_ext = os.path.splitext(file_path)[1].lower()

    if file_ext in ['.xlsx', '.xlsm', '.xls']:
        # Excel格式
        if sheet_name is None and use_gui:
            xls = pd.ExcelFile(file_path)
            print("檢測到的 sheets：")
            for s in xls.sheet_names: print(f"  - {s}")
            sheet_name = input("請輸入要導入的 sheet 名：").strip()
        elif sheet_name is None:
            sheet_name = 0  # 默認第一個sheet

        df = pd.read_excel(file_path, sheet_name=sheet_name, engine="openpyxl")
    elif file_ext == '.csv':
        # CSV格式
        df = pd.read_csv(file_path)
        sheet_name = "csv"
    else:
        raise RuntimeError(f"不支持的文件格式: {file_ext}")

    # --- 4. 過濾數據 ---
    # A. 只保留 stepType 為 0 的數據
    if "stepType" in df.columns:
        # 強制轉換類型以防萬一（處理 0.0 或 "0" 的情況）
        df = df[pd.to_numeric(df["stepType"], errors='coerce') == 0].copy()
    else:
        print("[Warning] 未在文件中找到 'stepType' 列，跳過過濾步驟。")

    # B. 只提取你指定的那些列（如果文件裡有這些列的話）
    existing_target_cols = [c for c in target_cols if c in df.columns]
    df = df[existing_target_cols].copy()

    total_records = len(df)


    # --- 5. 計算新列：time_visually 和 time ---
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

    # --- 6. 數據清理與準備寫入 ---
    # 清理列名（防止非法字符）
    df.columns = [_sanitize_identifier(c) for c in df.columns]

    # 處理 NaN
    df2 = df.where(pd.notnull(df), None)

    # --- 7. 寫入 SQLite ---
    os.makedirs(os.path.dirname(db_path) or ".", exist_ok=True)
    conn = sqlite3.connect(db_path)
    conn.execute("PRAGMA journal_mode=WAL")

    new_records = 0
    duplicate_records = 0

    try:
        cur = conn.cursor()

        # 檢查表是否存在
        cur.execute(f"SELECT name FROM sqlite_master WHERE type='table' AND name='{table_name}'")
        table_exists = cur.fetchone() is not None

        if if_exists == "replace":
            # 全量替換模式：刪除舊表
            cur.execute(f'DROP TABLE IF EXISTS "{table_name}"')
            table_exists = False

        if not table_exists:
            # 建表時顯式加入 id 主鍵
            col_defs = ['"id" INTEGER PRIMARY KEY AUTOINCREMENT']
            for col in df2.columns:
                sql_type = _detect_sqlite_type(df2[col])
                col_defs.append(f'"{col}" {sql_type}')

            create_sql = f'CREATE TABLE IF NOT EXISTS "{table_name}" ({", ".join(col_defs)})'
            cur.execute(create_sql)

        # --- 8. 去重邏輯 (僅在append模式下) ---
        if if_exists == "append" and deduplicate and "dataTime" in df2.columns:
            # 獲取已存在的dataTime值
            cur.execute(f'SELECT dataTime FROM "{table_name}"')
            existing_times = set(row[0] for row in cur.fetchall() if row[0] is not None)

            # 過濾掉已存在的記錄
            df_before = len(df2)
            df2 = df2[~df2["dataTime"].isin(existing_times)]
            duplicate_records = df_before - len(df2)

        # --- 9. 插入數據 ---
        if len(df2) > 0:
            placeholders = ", ".join(["?"] * len(df2.columns))
            col_list = ", ".join([f'"{c}"' for c in df2.columns])
            insert_sql = f'INSERT INTO "{table_name}" ({col_list}) VALUES ({placeholders})'

            cur.executemany(insert_sql, df2.values.tolist())
            new_records = len(df2)

        conn.commit()

    finally:
        conn.close()

    return {
        "file_path": file_path,
        "sheet_name": sheet_name,
        "columns": list(df2.columns),
        "total_records": total_records,
        "new_records": new_records,
        "duplicate_records": duplicate_records,
        "mode": if_exists,
    }


def main():
    parser = argparse.ArgumentParser(description="導入GPS軌跡數據到SQLite")
    parser.add_argument("--file", type=str, help="數據文件路徑 (Excel或CSV)")
    parser.add_argument("--db", type=str, default="data/tracks.db", help="數據庫路徑")
    parser.add_argument("--table", type=str, default="一生足迹", help="表名")
    parser.add_argument("--sheet", type=str, help="Excel sheet名稱")
    parser.add_argument("--mode", type=str, choices=["append", "replace"], default="append", help="導入模式")
    parser.add_argument("--deduplicate", type=str, choices=["true", "false"], default="true", help="是否去重")
    parser.add_argument("--gui", action="store_true", help="使用GUI模式選擇文件")

    args = parser.parse_args()

    # 解析布爾值
    deduplicate = args.deduplicate.lower() == "true"
    use_gui = args.gui or args.file is None

    try:
        result = import_data_to_sqlite(
            db_path=args.db,
            table_name=args.table,
            file_path=args.file,
            sheet_name=args.sheet,
            if_exists=args.mode,
            deduplicate=deduplicate,
            use_gui=use_gui,
        )

        print("-" * 50)
        print(f"導入完成！")
        print(f"文件路徑: {result['file_path']}")
        print(f"Sheet名稱: {result['sheet_name']}")
        print(f"導入模式: {result['mode']}")
        print(f"總記錄數: {result['total_records']}")
        print(f"新增記錄: {result['new_records']}")
        print(f"重複記錄: {result['duplicate_records']}")
        print(f"寫入列: {result['columns']}")
        print("-" * 50)

        # 返回JSON格式供Go調用
        import json
        print("\n=== JSON OUTPUT ===")
        print(json.dumps(result, ensure_ascii=False))

    except Exception as e:
        print(f"錯誤: {str(e)}", file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main()