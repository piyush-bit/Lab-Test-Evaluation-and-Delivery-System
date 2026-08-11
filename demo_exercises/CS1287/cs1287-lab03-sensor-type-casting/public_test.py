"""
Public unit tests for cs1287-lab03: Heterogeneous IoT Sensor Log Sanitizer.
"""

import unittest
from log_sanitizer import (
    safe_cast_float,
    parse_boolean_flag,
    parse_timestamp_to_epoch,
    sanitize_sensor_record,
)


class TestLogSanitizerPublic(unittest.TestCase):

    def test_safe_cast_float_basic(self):
        self.assertEqual(safe_cast_float("36.54321"), 36.5432)
        self.assertEqual(safe_cast_float(42), 42.0)
        self.assertEqual(safe_cast_float("invalid", default_val=10.0), 10.0)
        self.assertEqual(safe_cast_float(None), 0.0)

    def test_parse_boolean_flag_basic(self):
        self.assertTrue(parse_boolean_flag("YES"))
        self.assertTrue(parse_boolean_flag("1"))
        self.assertTrue(parse_boolean_flag(True))
        self.assertFalse(parse_boolean_flag("no"))
        self.assertFalse(parse_boolean_flag("0"))
        self.assertFalse(parse_boolean_flag("unknown"))

    def test_parse_timestamp_to_epoch_basic(self):
        self.assertEqual(parse_timestamp_to_epoch("1698765432"), 1698765432)
        self.assertEqual(parse_timestamp_to_epoch("1698765432.50"), 1698765432)
        self.assertEqual(parse_timestamp_to_epoch("2023-11-01 12:00:00"), 1698840000)
        self.assertEqual(parse_timestamp_to_epoch("bad-timestamp"), -1)

    def test_sanitize_sensor_record_basic(self):
        raw = {
            "sensor_id": "101",
            "temperature": "36.54321",
            "is_active": "YES",
            "timestamp": "1698765432",
        }
        sanitized = sanitize_sensor_record(raw)
        self.assertEqual(sanitized["sensor_id"], 101)
        self.assertEqual(sanitized["temperature"], 36.5432)
        self.assertIs(sanitized["is_active"], True)
        self.assertEqual(sanitized["timestamp"], 1698765432)
        self.assertIsInstance(sanitized["sensor_id"], int)
        self.assertIsInstance(sanitized["temperature"], float)
        self.assertIsInstance(sanitized["is_active"], bool)
        self.assertIsInstance(sanitized["timestamp"], int)


if __name__ == "__main__":
    unittest.main()
