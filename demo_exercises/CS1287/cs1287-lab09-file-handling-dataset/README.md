# CS1287 Lab 09: Medical Lab Test CSV & JSON Data Processor

## Overview
In this lab, you will implement a Python module `dataset_file_processor.py` that processes medical laboratory test records using Python's file handling mechanisms (CSV reading, text appending, and JSON serialization).

## Requirements

You must implement the following functions in `dataset_file_processor.py`:

### 1. `read_lab_results_csv(file_path: str) -> list`
Reads a CSV file containing lab results with the header:
`patient_id,test_name,test_value,units,is_abnormal`

- Parses each row into a dictionary with appropriate type conversions:
  - `patient_id` (`str`): Stripped patient ID string.
  - `test_name` (`str`): Stripped test name string.
  - `test_value` (`float`): Test value converted to float.
  - `units` (`str`): Stripped measurement units string.
  - `is_abnormal` (`bool`): `True` if the column value (case-insensitive) is `"true"` or `"1"`, `False` otherwise.
- Returns a list of record dictionaries:
  ```python
  [
      {
          "patient_id": "P101",
          "test_name": "Glucose",
          "test_value": 125.5,
          "units": "mg/dL",
          "is_abnormal": True
      },
      ...
  ]
  ```
- **Error Handling**: If `file_path` does not exist (`FileNotFoundError`), catch the error and return an empty list `[]`.

### 2. `filter_abnormal_records(records: list) -> list`
- Accepts a list of record dictionaries (as returned by `read_lab_results_csv`).
- Returns a new list containing only records where `is_abnormal` is `True`.

### 3. `export_summary_json(records: list, output_json_path: str) -> bool`
- Calculates summary statistics for the provided record list:
  - `total_records` (`int`): Total count of records.
  - `abnormal_count` (`int`): Number of records where `is_abnormal` is `True`.
  - `avg_test_value` (`float`): Average of all `test_value` fields rounded to 2 decimal places (`round(avg, 2)`). If `records` is empty, this should be `0.0`.
  - `test_counts` (`dict`): Frequency dictionary mapping each `test_name` (`str`) to its occurrence count (`int`).
- Writes the summary dictionary to `output_json_path` formatted with JSON indentation (`indent=2`).
- Returns `True` if written successfully, or `False` if an `IOError`, `PermissionError`, or `OSError` occurs during file operations.

### 4. `append_audit_log(log_file_path: str, log_message: str) -> bool`
- Appends a line formatted exactly as `"[AUDIT] <log_message>\n"` to the text file at `log_file_path`.
- Uses `with open(..., 'a')` for clean file management.
- Returns `True` on success, or `False` if an `IOError`, `PermissionError`, or `OSError` occurs.

## Execution and Testing

- **Run public tests locally:**
  ```bash
  make test-public
  ```
  or
  ```bash
  ./run
  ```

- **Clean temporary build files:**
  ```bash
  make clean
  ```
