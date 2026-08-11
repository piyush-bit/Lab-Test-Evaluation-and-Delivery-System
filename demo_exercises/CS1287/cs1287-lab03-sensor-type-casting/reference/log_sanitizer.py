"""
CS1287 - Lab 03: Heterogeneous IoT Sensor Log Sanitizer Reference Solution
"""

from datetime import datetime, timezone


def safe_cast_float(val, default_val: float = 0.0) -> float:
    """
    Attempts to cast `val` (string, int, float) to float rounded to 4 decimal places.
    If `val` is None or casting raises ValueError/TypeError, returns `default_val`.
    """
    if val is None:
        return float(default_val)
    try:
        res = float(val)
        return round(res, 4)
    except (ValueError, TypeError, OverflowError):
        return float(default_val)


def parse_boolean_flag(val) -> bool:
    """
    Casts various representations to a boolean flag according to specification.
    """
    if isinstance(val, bool):
        return val
    if isinstance(val, (int, float)):
        return val != 0
    if isinstance(val, str):
        s = val.strip().lower()
        if s in {"true", "1", "yes", "on", "y"}:
            return True
        if s in {"false", "0", "no", "off", "n"}:
            return False
        return False
    return False


def parse_timestamp_to_epoch(timestamp_str: str) -> int:
    """
    Parses string timestamps to integer Unix epoch timestamp (seconds).
    Supports numeric strings and ISO format "YYYY-MM-DD HH:MM:SS". Returns -1 on failure.
    """
    if timestamp_str is None:
        return -1
    if isinstance(timestamp_str, (int, float)):
        return int(timestamp_str)
    if not isinstance(timestamp_str, str):
        return -1
    s = timestamp_str.strip()
    try:
        return int(float(s))
    except ValueError:
        pass

    try:
        dt = datetime.strptime(s, "%Y-%m-%d %H:%M:%S")
        return int(dt.replace(tzinfo=timezone.utc).timestamp())
    except (ValueError, TypeError):
        return -1


def sanitize_sensor_record(raw_record: dict) -> dict:
    """
    Sanitizes raw sensor record dict into strict data types:
    - sensor_id: int (default -1)
    - temperature: float (safe_cast_float)
    - is_active: bool (parse_boolean_flag)
    - timestamp: int (parse_timestamp_to_epoch)
    """
    if not isinstance(raw_record, dict):
        return {
            "sensor_id": -1,
            "temperature": 0.0,
            "is_active": False,
            "timestamp": -1
        }

    def safe_cast_int(val, default_val: int = -1) -> int:
        if val is None:
            return default_val
        try:
            return int(float(val))
        except (ValueError, TypeError, OverflowError):
            return default_val

    sanitized = dict(raw_record)
    sanitized["sensor_id"] = safe_cast_int(raw_record.get("sensor_id"), -1)
    sanitized["temperature"] = safe_cast_float(raw_record.get("temperature"), 0.0)
    sanitized["is_active"] = parse_boolean_flag(raw_record.get("is_active"))
    sanitized["timestamp"] = parse_timestamp_to_epoch(raw_record.get("timestamp"))
    return sanitized
