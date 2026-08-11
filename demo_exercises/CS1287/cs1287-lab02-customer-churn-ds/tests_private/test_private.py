import unittest
from churn_analyzer import (
    extract_unique_customers,
    aggregate_customer_metrics,
    filter_high_value_churn_risks,
    merge_customer_profiles,
)


class TestPrivateExtractUnique(unittest.TestCase):

    def test_private_extract_unique_empty(self):
        self.assertEqual(extract_unique_customers([]), set())

    def test_private_extract_unique_single(self):
        txs = [{"customer_id": "C999", "amount": 10.0}]
        self.assertEqual(extract_unique_customers(txs), {"C999"})

    def test_private_extract_unique_duplicates(self):
        txs = [{"customer_id": f"C{i%5}", "amount": float(i)} for i in range(50)]
        expected = {f"C{i}" for i in range(5)}
        self.assertEqual(extract_unique_customers(txs), expected)


class TestPrivateAggregateMetrics(unittest.TestCase):

    def test_private_aggregate_metrics_empty(self):
        self.assertEqual(aggregate_customer_metrics([]), {})

    def test_private_aggregate_metrics_rounding(self):
        txs = [
            {"customer_id": "C101", "amount": 10.333},
            {"customer_id": "C101", "amount": 20.666},
            {"customer_id": "C101", "amount": 5.001},
        ]
        res = aggregate_customer_metrics(txs)
        self.assertEqual(res["C101"]["total_spent"], 36.00)
        self.assertEqual(res["C101"]["transaction_count"], 3)
        self.assertEqual(res["C101"]["avg_spend"], 12.00)

    def test_private_aggregate_metrics_multiple_customers(self):
        txs = [
            {"customer_id": "C1", "amount": 100.0},
            {"customer_id": "C2", "amount": 200.0},
            {"customer_id": "C1", "amount": 50.0},
        ]
        res = aggregate_customer_metrics(txs)
        self.assertEqual(res["C1"]["total_spent"], 150.0)
        self.assertEqual(res["C1"]["avg_spend"], 75.0)
        self.assertEqual(res["C2"]["total_spent"], 200.0)
        self.assertEqual(res["C2"]["avg_spend"], 200.0)


class TestPrivateFilterChurn(unittest.TestCase):

    def test_private_filter_churn_none_qualify(self):
        metrics = {"C1": {"total_spent": 50.0, "transaction_count": 1, "avg_spend": 50.0}}
        inactivity = {"C1": 10}
        self.assertEqual(filter_high_value_churn_risks(metrics, inactivity, 100.0, 30), [])

    def test_private_filter_churn_boundary(self):
        metrics = {
            "C1": {"total_spent": 100.0, "transaction_count": 1, "avg_spend": 100.0},
            "C2": {"total_spent": 99.99, "transaction_count": 1, "avg_spend": 99.99},
        }
        inactivity = {"C1": 30, "C2": 30}
        res = filter_high_value_churn_risks(metrics, inactivity, 100.0, 30)
        self.assertEqual(res, ["C1"])

    def test_private_filter_churn_sorting_and_missing(self):
        metrics = {
            "C30": {"total_spent": 500.0, "transaction_count": 5, "avg_spend": 100.0},
            "C10": {"total_spent": 500.0, "transaction_count": 5, "avg_spend": 100.0},
            "C20": {"total_spent": 500.0, "transaction_count": 5, "avg_spend": 100.0},
            "C40": {"total_spent": 500.0, "transaction_count": 5, "avg_spend": 100.0},
        }
        inactivity = {"C30": 40, "C10": 50, "C20": 45}  # C40 is not in inactivity dict
        res = filter_high_value_churn_risks(metrics, inactivity, 100.0, 30)
        self.assertEqual(res, ["C10", "C20", "C30"])


class TestPrivateMergeProfiles(unittest.TestCase):

    def test_private_merge_profiles_mismatch(self):
        p1 = ("C1", ["a"], {"k": 1})
        p2 = ("C2", ["b"], {"k": 2})
        self.assertIsNone(merge_customer_profiles(p1, p2))

    def test_private_merge_profiles_empty_tags_pref(self):
        p1 = ("C1", [], {})
        p2 = ("C1", ["tag1"], {"a": 1})
        res = merge_customer_profiles(p1, p2)
        self.assertEqual(res, ("C1", ["tag1"], {"a": 1}))

    def test_private_merge_profiles_overwriting(self):
        p1 = ("C1", ["b", "a"], {"setting": "off", "color": "blue"})
        p2 = ("C1", ["c", "a"], {"setting": "on", "font": "arial"})
        res = merge_customer_profiles(p1, p2)
        expected_tags = ["a", "b", "c"]
        expected_pref = {"setting": "on", "color": "blue", "font": "arial"}
        self.assertEqual(res, ("C1", expected_tags, expected_pref))


if __name__ == "__main__":
    unittest.main()
