# Experiment 3: Heterogeneous IoT Sensor Log Sanitizer

## Overview
In distributed Internet of Things (IoT) monitoring systems, environmental sensors, smart meters, and edge gateways transmit telemetry records in disparate data formats. Due to hardware variations, firmware legacy issues, and transmission corruption, raw sensor logs often contain mismatched data types (e.g. numbers formatted as strings, boolean flags represented as `"YES"` or `"1"`, timestamps in mixed numeric or ISO date formats).

In this lab, you will build a Python log sanitizer to normalize heterogeneous IoT sensor log dictionaries into standardized, strictly-typed data structures.

## Learning Objectives
- Perform explicit data type conversions (`str`, `int`, `float`, `bool`) in Python.
- Implement defensive programming with `try-except` blocks for type safety.
- Parse ISO datetime strings into Unix epoch timestamps using standard libraries (`datetime`).
- Process and sanitize complex dictionary structures containing mixed telemetry fields.

## File Structure
```
cs1287-lab03-sensor-type-casting/
├── manifest.json              # Lab metadata and test suite definitions
├── README.md                  # Problem statement & instructions
├── Makefile                   # Test automation targets
├── run                        # Executable test runner wrapper
├── log_sanitizer.py           # Starter code (implement your solution here)
├── public_test.py             # Public unit tests
├── reference/
│   └── log_sanitizer.py       # Instructor reference solution
└── tests_private/
    └── test_private.py        # Private grading test suite
```

## Function Specifications

Implement the following functions in `log_sanitizer.py`:

### 1. `safe_cast_float(val, default_val: float = 0.0) -> float`
- Converts `val` (string, integer, float) to a float rounded to **4 decimal places**.
- If `val` is `None` or if casting raises a `ValueError` or `TypeError`, returns `default_val`.

### 2. `parse_boolean_flag(val) -> bool`
- Normalizes various truthy and falsy representations into a boolean `True` or `False`:
  - **True representations**: `"true"`, `"1"`, `"yes"`, `"on"`, `"y"` (case-insensitive, whitespace-trimmed), non-zero numbers (`1`, `1.0`, `-5`), or boolean `True`.
  - **False representations**: `"false"`, `"0"`, `"no"`, `"off"`, `"n"` (case-insensitive, whitespace-trimmed), zero numbers (`0`, `0.0`), or boolean `False`.
  - **Unrecognized/Invalid inputs**: Any other string, `None`, list, or dict returns `False`.

### 3. `parse_timestamp_to_epoch(timestamp_str: str) -> int`
- Converts string or numeric timestamps into an integer **Unix epoch timestamp (seconds)**:
  - **Numeric format** (e.g. `"1698765432"` or `"1698765432.50"` or integer/float) -> cast to integer seconds.
  - **ISO datetime format** (`"YYYY-MM-DD HH:MM:SS"`) -> parse using `datetime.strptime` (UTC timezone) and convert to integer Unix timestamp.
  - **Invalid format/type** -> returns `-1`.

### 4. `sanitize_sensor_record(raw_record: dict) -> dict`
- Accepts a raw sensor log dictionary `raw_record`.
- Returns a sanitized dictionary with strict data types:
  - `"sensor_id"`: `int` (default `-1` if missing or unparseable)
  - `"temperature"`: `float` (using `safe_cast_float`, default `0.0`)
  - `"is_active"`: `bool` (using `parse_boolean_flag`)
  - `"timestamp"`: `int` (using `parse_timestamp_to_epoch`, default `-1`)

## Example Usage

```python
from log_sanitizer import sanitize_sensor_record

raw_data = {
    "sensor_id": "101",
    "temperature": "36.54321",
    "is_active": "YES",
    "timestamp": "1698765432"
}

clean_data = sanitize_sensor_record(raw_data)
print(clean_data)
# Output:
# {'sensor_id': 101, 'temperature': 36.5432, 'is_active': True, 'timestamp': 1698765432}
```

## Running Tests

Execute the public test suite locally:
```bash
./run public
# Or using make directly:
make test-public
```
