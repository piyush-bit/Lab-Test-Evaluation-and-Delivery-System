import os
import json
import tempfile
import unittest
import dataset_file_processor as dfp


class TestPrivateDatasetFileProcessor(unittest.TestCase):

    def setUp(self):
        self.temp_dir = tempfile.TemporaryDirectory()

    def tearDown(self):
        self.temp_dir.cleanup()

    # ---------------------------------------------------------------------------
    # Private Test Group 1: read_lab_results_csv
    # ---------------------------------------------------------------------------
    def test_private_read_csv_boolean_variants_and_whitespace(self):
        csv_path = os.path.join(self.temp_dir.name, "lab_variants.csv")
        csv_content = (
            "patient_id,test_name,test_value,units,is_abnormal\n"
            " P101 , Glucose , 120.456 , mg/dL , TRUE \n"
            "P102, HbA1c ,6.8, % , 1\n"
            "P103, Sodium ,138.0, mEq/L , FALSE\n"
            "P104, Potassium ,4.2, mEq/L , 0\n"
        )
        with open(csv_path, "w", encoding="utf-8") as f:
            f.write(csv_content)

        records = dfp.read_lab_results_csv(csv_path)
        self.assertEqual(len(records), 4)

        self.assertEqual(records[0], {
            "patient_id": "P101",
            "test_name": "Glucose",
            "test_value": 120.456,
            "units": "mg/dL",
            "is_abnormal": True
        })
        self.assertEqual(records[1], {
            "patient_id": "P102",
            "test_name": "HbA1c",
            "test_value": 6.8,
            "units": "%",
            "is_abnormal": True
        })
        self.assertEqual(records[2], {
            "patient_id": "P103",
            "test_name": "Sodium",
            "test_value": 138.0,
            "units": "mEq/L",
            "is_abnormal": False
        })
        self.assertEqual(records[3], {
            "patient_id": "P104",
            "test_name": "Potassium",
            "test_value": 4.2,
            "units": "mEq/L",
            "is_abnormal": False
        })

    def test_private_read_csv_header_only(self):
        csv_path = os.path.join(self.temp_dir.name, "empty_lab.csv")
        with open(csv_path, "w", encoding="utf-8") as f:
            f.write("patient_id,test_name,test_value,units,is_abnormal\n")

        records = dfp.read_lab_results_csv(csv_path)
        self.assertEqual(records, [])

    def test_private_read_csv_nonexistent_file(self):
        csv_path = os.path.join(self.temp_dir.name, "does_not_exist_12345.csv")
        records = dfp.read_lab_results_csv(csv_path)
        self.assertEqual(records, [])

    # ---------------------------------------------------------------------------
    # Private Test Group 2: filter_abnormal_records
    # ---------------------------------------------------------------------------
    def test_private_filter_abnormal_mixed(self):
        records = [
            {"patient_id": "P1", "test_name": "Glucose", "test_value": 140.0, "units": "mg/dL", "is_abnormal": True},
            {"patient_id": "P2", "test_name": "Glucose", "test_value": 95.0, "units": "mg/dL", "is_abnormal": False},
            {"patient_id": "P3", "test_name": "ALT", "test_value": 55.0, "units": "U/L", "is_abnormal": True},
            {"patient_id": "P4", "test_name": "AST", "test_value": 22.0, "units": "U/L", "is_abnormal": False}
        ]
        abnormal = dfp.filter_abnormal_records(records)
        self.assertEqual(len(abnormal), 2)
        self.assertEqual([r["patient_id"] for r in abnormal], ["P1", "P3"])

    def test_private_filter_abnormal_none_or_all(self):
        no_abnormal = [
            {"patient_id": "P1", "test_name": "TSH", "test_value": 2.5, "units": "mIU/L", "is_abnormal": False}
        ]
        self.assertEqual(dfp.filter_abnormal_records(no_abnormal), [])

        all_abnormal = [
            {"patient_id": "P1", "test_name": "TSH", "test_value": 12.5, "units": "mIU/L", "is_abnormal": True},
            {"patient_id": "P2", "test_name": "TSH", "test_value": 0.1, "units": "mIU/L", "is_abnormal": True}
        ]
        self.assertEqual(len(dfp.filter_abnormal_records(all_abnormal)), 2)

    def test_private_filter_abnormal_empty_input(self):
        self.assertEqual(dfp.filter_abnormal_records([]), [])

    # ---------------------------------------------------------------------------
    # Private Test Group 3: export_summary_json
    # ---------------------------------------------------------------------------
    def test_private_export_json_valid_data(self):
        records = [
            {"patient_id": "P1", "test_name": "Glucose", "test_value": 100.333, "units": "mg/dL", "is_abnormal": False},
            {"patient_id": "P2", "test_name": "Glucose", "test_value": 140.666, "units": "mg/dL", "is_abnormal": True},
            {"patient_id": "P3", "test_name": "HbA1c", "test_value": 8.0, "units": "%", "is_abnormal": True},
            {"patient_id": "P4", "test_name": "Cholesterol", "test_value": 210.0, "units": "mg/dL", "is_abnormal": True}
        ]
        json_path = os.path.join(self.temp_dir.name, "summary_private.json")
        status = dfp.export_summary_json(records, json_path)
        self.assertTrue(status)

        with open(json_path, "r", encoding="utf-8") as f:
            content = f.read()
            data = json.loads(content)

        # Ensure indenting formatting
        self.assertIn("\n", content)
        self.assertEqual(data["total_records"], 4)
        self.assertEqual(data["abnormal_count"], 3)
        # sum = 100.333 + 140.666 + 8.0 + 210.0 = 458.999; 458.999 / 4 = 114.74975 -> rounded to 2 decimals = 114.75
        self.assertEqual(data["avg_test_value"], 114.75)
        self.assertEqual(data["test_counts"], {
            "Glucose": 2,
            "HbA1c": 1,
            "Cholesterol": 1
        })

    def test_private_export_json_empty_records(self):
        json_path = os.path.join(self.temp_dir.name, "summary_empty.json")
        status = dfp.export_summary_json([], json_path)
        self.assertTrue(status)

        with open(json_path, "r", encoding="utf-8") as f:
            data = json.load(f)

        self.assertEqual(data["total_records"], 0)
        self.assertEqual(data["abnormal_count"], 0)
        self.assertEqual(data["avg_test_value"], 0.0)
        self.assertEqual(data["test_counts"], {})

    def test_private_export_json_invalid_path(self):
        invalid_path = os.path.join(self.temp_dir.name, "non_existent_dir_9999", "out.json")
        status = dfp.export_summary_json([], invalid_path)
        self.assertFalse(status)

    # ---------------------------------------------------------------------------
    # Private Test Group 4: append_audit_log
    # ---------------------------------------------------------------------------
    def test_private_append_log_sequential(self):
        log_path = os.path.join(self.temp_dir.name, "audit_private.log")
        entries = ["System boot", "CSV loaded: 50 records", "Summary exported"]

        for entry in entries:
            res = dfp.append_audit_log(log_path, entry)
            self.assertTrue(res)

        with open(log_path, "r", encoding="utf-8") as f:
            lines = f.readlines()

        expected = [f"[AUDIT] {entry}\n" for entry in entries]
        self.assertEqual(lines, expected)

    def test_private_append_log_invalid_path(self):
        invalid_log_path = os.path.join(self.temp_dir.name, "non_existent_dir_9999", "audit.log")
        res = dfp.append_audit_log(invalid_log_path, "Test message")
        self.assertFalse(res)


if __name__ == "__main__":
    unittest.main()
