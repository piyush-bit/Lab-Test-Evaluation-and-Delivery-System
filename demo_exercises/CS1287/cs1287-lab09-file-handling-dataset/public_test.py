import os
import json
import tempfile
import unittest
import dataset_file_processor as dfp


class TestPublicDatasetFileProcessor(unittest.TestCase):

    def setUp(self):
        self.temp_dir = tempfile.TemporaryDirectory()

    def tearDown(self):
        self.temp_dir.cleanup()

    def test_read_lab_results_csv_basic(self):
        csv_path = os.path.join(self.temp_dir.name, "sample_lab.csv")
        csv_content = (
            "patient_id,test_name,test_value,units,is_abnormal\n"
            "P101,Glucose,140.5,mg/dL,True\n"
            "P102,Cholesterol,185.0,mg/dL,False\n"
        )
        with open(csv_path, "w", encoding="utf-8") as f:
            f.write(csv_content)

        records = dfp.read_lab_results_csv(csv_path)
        self.assertEqual(len(records), 2)
        self.assertEqual(records[0], {
            "patient_id": "P101",
            "test_name": "Glucose",
            "test_value": 140.5,
            "units": "mg/dL",
            "is_abnormal": True
        })
        self.assertEqual(records[1], {
            "patient_id": "P102",
            "test_name": "Cholesterol",
            "test_value": 185.0,
            "units": "mg/dL",
            "is_abnormal": False
        })

    def test_read_lab_results_csv_missing_file(self):
        records = dfp.read_lab_results_csv("non_existent_file_9999.csv")
        self.assertEqual(records, [])

    def test_filter_abnormal_records_basic(self):
        sample_records = [
            {"patient_id": "P1", "test_name": "Glucose", "test_value": 140.5, "units": "mg/dL", "is_abnormal": True},
            {"patient_id": "P2", "test_name": "Glucose", "test_value": 90.0, "units": "mg/dL", "is_abnormal": False},
            {"patient_id": "P3", "test_name": "HbA1c", "test_value": 7.2, "units": "%", "is_abnormal": True}
        ]
        filtered = dfp.filter_abnormal_records(sample_records)
        self.assertEqual(len(filtered), 2)
        self.assertEqual(filtered[0]["patient_id"], "P1")
        self.assertEqual(filtered[1]["patient_id"], "P3")

    def test_export_summary_json_basic(self):
        sample_records = [
            {"patient_id": "P1", "test_name": "Glucose", "test_value": 100.0, "units": "mg/dL", "is_abnormal": False},
            {"patient_id": "P2", "test_name": "Glucose", "test_value": 140.0, "units": "mg/dL", "is_abnormal": True},
            {"patient_id": "P3", "test_name": "HbA1c", "test_value": 7.5, "units": "%", "is_abnormal": True}
        ]
        json_path = os.path.join(self.temp_dir.name, "summary.json")
        result = dfp.export_summary_json(sample_records, json_path)
        self.assertTrue(result)
        self.assertTrue(os.path.exists(json_path))

        with open(json_path, "r", encoding="utf-8") as f:
            data = json.load(f)

        self.assertEqual(data["total_records"], 3)
        self.assertEqual(data["abnormal_count"], 2)
        self.assertEqual(data["avg_test_value"], 82.5)
        self.assertEqual(data["test_counts"], {"Glucose": 2, "HbA1c": 1})

    def test_append_audit_log_basic(self):
        log_path = os.path.join(self.temp_dir.name, "audit.log")
        res1 = dfp.append_audit_log(log_path, "Processor initialized")
        res2 = dfp.append_audit_log(log_path, "Batch processing complete")

        self.assertTrue(res1)
        self.assertTrue(res2)

        with open(log_path, "r", encoding="utf-8") as f:
            lines = f.readlines()

        self.assertEqual(lines, [
            "[AUDIT] Processor initialized\n",
            "[AUDIT] Batch processing complete\n"
        ])


if __name__ == "__main__":
    unittest.main()
