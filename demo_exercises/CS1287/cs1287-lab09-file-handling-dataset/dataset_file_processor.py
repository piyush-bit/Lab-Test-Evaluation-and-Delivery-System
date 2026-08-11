"""
Medical Lab Test CSV & JSON Data Processor
File Handling Lab Exercise
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
    # TODO: Read CSV file and parse rows into list of dictionaries
    return []


def filter_abnormal_records(records: list) -> list:
    """
    Given list of record dicts, returns list of records where is_abnormal is True.
    """
    # TODO: Filter and return records where is_abnormal is True
    return []


def export_summary_json(records: list, output_json_path: str) -> bool:
    """
    Computes summary stats:
    - total_records: int
    - abnormal_count: int
    - avg_test_value: float rounded to 2 decimal places (0.0 if empty)
    - test_counts: dict counting occurrences per test_name

    Writes dict to JSON file at output_json_path with indent=2.
    Returns True on success, False on IOError/PermissionError.
    """
    # TODO: Calculate summary statistics and export to JSON file
    return False


def append_audit_log(log_file_path: str, log_message: str) -> bool:
    """
    Appends line "[AUDIT] <log_message>\n" to text file at log_file_path.
    Uses with open(..., 'a').
    Returns True on success, False on IOError.
    """
    # TODO: Append formatted audit log entry to specified text file
    return False
