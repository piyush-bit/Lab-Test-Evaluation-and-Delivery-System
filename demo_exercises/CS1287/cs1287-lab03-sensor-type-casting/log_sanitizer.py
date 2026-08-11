"""
CS1287 - Lab 03: Heterogeneous IoT Sensor Log Sanitizer

This module provides data type conversion and sanitization functions for IoT
sensor logs collected from heterogeneous field devices.
"""

from datetime import datetime, timezone


def safe_cast_float(val, default_val: float = 0.0) -> float:
    """
    Attempts to cast `val` (string, int, float) to float rounded to 4 decimal places.
    If `val` is None or casting raises ValueError/TypeError, returns `default_val`.

    :param val: Input value of arbitrary type.
    :param default_val: Fallback float value if conversion fails.
    :return: Floated and rounded value or default_val.
    """
    # TODO: Implement safe float casting with 4 decimal places rounding
    return default_val


def parse_boolean_flag(val) -> bool:
    """
    Casts various representations to a boolean flag:
      - True representations: "true", "1", "yes", "on", "y" (case-insensitive),
        or int/float non-zero (1, 1.0, etc.), or boolean True.
      - False representations: "false", "0", "no", "off", "n",
        int/float 0, or boolean False.
      - Any invalid or unrecognized string/type returns False.

    :param val: Input value of arbitrary type.
    :return: Parsed boolean value.
    """
    # TODO: Implement boolean flag parsing according to spec
    return False


def parse_timestamp_to_epoch(timestamp_str: str) -> int:
    """
    Parses string timestamps to an integer Unix epoch timestamp (seconds):
      - Numeric string (e.g. "1698765432" or "1698765432.50") or numeric value -> cast to int.
      - ISO format string "YYYY-MM-DD HH:MM:SS" -> parse using datetime.strptime (UTC)
        and return integer Unix timestamp.
      - If parsing fails or input is invalid format/type, return -1.

    :param timestamp_str: Timestamp input (string or numeric).
    :return: Integer epoch timestamp in seconds, or -1 on failure.
    """
    # TODO: Implement timestamp to Unix epoch parsing
    return -1


def sanitize_sensor_record(raw_record: dict) -> dict:
    """
    Given a raw sensor log dictionary, returns a sanitized dictionary with strict types:
      {"sensor_id": int, "temperature": float, "is_active": bool, "timestamp": int}

    Uses safe_cast_float, parse_boolean_flag, and parse_timestamp_to_epoch.
    sensor_id is cast to int (default -1 if invalid or missing).

    :param raw_record: Raw dictionary received from IoT device.
    :return: Sanitized dictionary with strictly-typed fields.
    """
    # TODO: Implement complete dictionary sanitization using helper functions
    return {
        "sensor_id": -1,
        "temperature": 0.0,
        "is_active": False,
        "timestamp": -1
    }
