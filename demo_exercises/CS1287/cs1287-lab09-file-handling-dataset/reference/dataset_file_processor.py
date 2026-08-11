"""
Medical Lab Test CSV & JSON Data Processor
Reference Solution
"""

import csv
import json


def read_lab_results_csv(file_path: str) -> list:
    """
    Reads CSV file with header: patient_id,test_name,test_value,units,is_abnormal
    Parses rows into a list of dicts:
    [{"patient_id": str, "test_name": str, "test_value": float, "units": str, "is_abnormal": bool}, ...]
    Handles FileNotFoundError gracefully by returning [].
    """
    records = []
    try:
        with open(file_path, mode='r', newline='', encoding='utf-8') as f:
            reader = csv.DictReader(f)
            for row in reader:
                abnormal_raw = str(row.get('is_abnormal', '')).strip().lower()
                abnormal_val = abnormal_raw in ('true', '1')
                try:
                    test_val = float(row.get('test_value', 0.0))
                except (ValueError, TypeError):
                    test_val = 0.0

                record = {
                    "patient_id": str(row.get('patient_id', '')).strip(),
                    "test_name": str(row.get('test_name', '')).strip(),
                    "test_value": test_val,
                    "units": str(row.get('units', '')).strip(),
                    "is_abnormal": abnormal_val
                }
                records.append(record)
    except FileNotFoundError:
        return []
    return records


def filter_abnormal_records(records: list) -> list:
    """
    Given list of record dicts, returns list of records where is_abnormal is True.
    """
    return [rec for rec in records if rec.get('is_abnormal') is True]


def export_summary_json(records: list, output_json_path: str) -> bool:
    """
    Computes summary stats:
    - total_records: int
    - abnormal_count: int
    - avg_test_value: float rounded to 2 decimal places (0.0 if empty)
    - test_counts: dict counting occurrences per test_name

    Writes dict to JSON file at output_json_path with indent=2.
    Returns True on success, False on IOError/PermissionError/OSError.
    """
    try:
        total_records = len(records)
        abnormal_count = sum(1 for rec in records if rec.get('is_abnormal') is True)
        
        if total_records > 0:
            avg_test_value = round(sum(rec.get('test_value', 0.0) for rec in records) / total_records, 2)
        else:
            avg_test_value = 0.0
            
        test_counts = {}
        for rec in records:
            name = rec.get('test_name', '')
            test_counts[name] = test_counts.get(name, 0) + 1
            
        summary = {
            "total_records": total_records,
            "abnormal_count": abnormal_count,
            "avg_test_value": avg_test_value,
            "test_counts": test_counts
        }
        
        with open(output_json_path, mode='w', encoding='utf-8') as f:
            json.dump(summary, f, indent=2)
        return True
    except (IOError, OSError, PermissionError):
        return False


def append_audit_log(log_file_path: str, log_message: str) -> bool:
    """
    Appends line "[AUDIT] <log_message>\n" to text file at log_file_path.
    Uses with open(..., 'a').
    Returns True on success, False on IOError/PermissionError/OSError.
    """
    try:
        with open(log_file_path, mode='a', encoding='utf-8') as f:
            f.write(f"[AUDIT] {log_message}\n")
        return True
    except (IOError, OSError, PermissionError):
        return False
