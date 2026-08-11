"""
Private test suite for cs1287-lab03 grading.
"""

import unittest
from log_sanitizer import (
    safe_cast_float,
    parse_boolean_flag,
    parse_timestamp_to_epoch,
    sanitize_sensor_record,
)


class TestPrivateFloatCasting(unittest.TestCase):

    def test_private_float_casting_precision(self):
        self.assertEqual(safe_cast_float("12.3456789"), 12.3457)
        self.assertEqual(safe_cast_float("-99.10001"), -99.1000)
        self.assertEqual(safe_cast_float(0), 0.0)
        self.assertEqual(safe_cast_float(3.14159265), 3.1416)

    def test_private_float_casting_fallbacks(self):
        self.assertEqual(safe_cast_float("corrupted_data", default_val=-1.0), -1.0)
        self.assertEqual(safe_cast_float(None, default_val=5.5), 5.5)
        self.assertEqual(safe_cast_float([1, 2, 3], default_val=0.0), 0.0)
        self.assertEqual(safe_cast_float({"temp": 20}, default_val=0.0), 0.0)


class TestPrivateBooleanFlags(unittest.TestCase):

    def test_private_boolean_flags_true_variants(self):
        for val in ["true", "TRUE", " True ", "1", "yes", "YES", "on", "ON", "y", "Y"]:
            self.assertTrue(parse_boolean_flag(val), f"Failed for {val}")
        self.assertTrue(parse_boolean_flag(1))
        self.assertTrue(parse_boolean_flag(1.0))
        self.assertTrue(parse_boolean_flag(-5))
        self.assertTrue(parse_boolean_flag(True))

    def test_private_boolean_flags_false_and_invalid(self):
        for val in ["false", "FALSE", "0", "no", "NO", "off", "OFF", "n", "N"]:
            self.assertFalse(parse_boolean_flag(val), f"Failed for {val}")
        self.assertFalse(parse_boolean_flag(0))
        self.assertFalse(parse_boolean_flag(0.0))
        self.assertFalse(parse_boolean_flag(False))
        # Invalid inputs
        for val in ["maybe", "2", "-1_str", "", None, [True], {"active": True}]:
            self.assertFalse(parse_boolean_flag(val), f"Failed for invalid {val}")


class TestPrivateTimestampEpoch(unittest.TestCase):

    def test_private_timestamp_epoch_numeric(self):
        self.assertEqual(parse_timestamp_to_epoch("1698765432"), 1698765432)
        self.assertEqual(parse_timestamp_to_epoch("1698765432.99"), 1698765432)
        self.assertEqual(parse_timestamp_to_epoch(1698765432), 1698765432)
        self.assertEqual(parse_timestamp_to_epoch(1698765432.4), 1698765432)

    def test_private_timestamp_epoch_iso(self):
        self.assertEqual(parse_timestamp_to_epoch("1970-01-01 00:00:00"), 0)
        self.assertEqual(parse_timestamp_to_epoch("2023-11-01 12:00:00"), 1698840000)

    def test_private_timestamp_epoch_invalid(self):
        self.assertEqual(parse_timestamp_to_epoch("2023/11/01 12:00:00"), -1)
        self.assertEqual(parse_timestamp_to_epoch("not-a-date"), -1)
        self.assertEqual(parse_timestamp_to_epoch(""), -1)
        self.assertEqual(parse_timestamp_to_epoch(None), -1)
        self.assertEqual(parse_timestamp_to_epoch([]), -1)


class TestPrivateRecordSanitization(unittest.TestCase):

    def test_private_record_sanitization_full(self):
        raw = {
            "sensor_id": " 204 ",
            "temperature": "24.56789",
            "is_active": "on",
            "timestamp": "2023-11-01 12:00:00",
        }
        sanitized = sanitize_sensor_record(raw)
        self.assertEqual(sanitized["sensor_id"], 204)
        self.assertEqual(sanitized["temperature"], 24.5679)
        self.assertIs(sanitized["is_active"], True)
        self.assertEqual(sanitized["timestamp"], 1698840000)

    def test_private_record_sanitization_missing_and_corrupt(self):
        raw = {
            "sensor_id": "invalid_id",
            "temperature": None,
            "is_active": "unknown_state",
            "timestamp": "invalid_time",
        }
        sanitized = sanitize_sensor_record(raw)
        self.assertEqual(sanitized["sensor_id"], -1)
        self.assertEqual(sanitized["temperature"], 0.0)
        self.assertIs(sanitized["is_active"], False)
        self.assertEqual(sanitized["timestamp"], -1)

    def test_private_record_sanitization_invalid_input_type(self):
        sanitized = sanitize_sensor_record(None)
        self.assertEqual(sanitized["sensor_id"], -1)
        self.assertEqual(sanitized["temperature"], 0.0)
        self.assertIs(sanitized["is_active"], False)
        self.assertEqual(sanitized["timestamp"], -1)


if __name__ == "__main__":
    unittest.main()
