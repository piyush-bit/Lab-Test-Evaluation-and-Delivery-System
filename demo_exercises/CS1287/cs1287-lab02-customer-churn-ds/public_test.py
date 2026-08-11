import unittest
from churn_analyzer import (
    extract_unique_customers,
    aggregate_customer_metrics,
    filter_high_value_churn_risks,
    merge_customer_profiles,
)


class TestPublicChurnAnalyzer(unittest.TestCase):

    def test_extract_unique_customers_basic(self):
        txs = [
            {"customer_id": "C101", "amount": 49.99},
            {"customer_id": "C102", "amount": 15.00},
            {"customer_id": "C101", "amount": 25.50},
        ]
        result = extract_unique_customers(txs)
        self.assertEqual(result, {"C101", "C102"})

    def test_extract_unique_customers_empty(self):
        self.assertEqual(extract_unique_customers([]), set())

    def test_aggregate_customer_metrics_basic(self):
        txs = [
            {"customer_id": "C101", "amount": 50.00},
            {"customer_id": "C101", "amount": 25.50},
            {"customer_id": "C102", "amount": 100.00},
        ]
        res = aggregate_customer_metrics(txs)
        self.assertIn("C101", res)
        self.assertIn("C102", res)
        self.assertEqual(res["C101"]["total_spent"], 75.50)
        self.assertEqual(res["C101"]["transaction_count"], 2)
        self.assertEqual(res["C101"]["avg_spend"], 37.75)
        self.assertEqual(res["C102"]["total_spent"], 100.00)
        self.assertEqual(res["C102"]["transaction_count"], 1)
        self.assertEqual(res["C102"]["avg_spend"], 100.00)

    def test_filter_high_value_churn_risks_basic(self):
        metrics = {
            "C101": {"total_spent": 500.00, "transaction_count": 5, "avg_spend": 100.00},
            "C102": {"total_spent": 50.00, "transaction_count": 1, "avg_spend": 50.00},
            "C103": {"total_spent": 300.00, "transaction_count": 3, "avg_spend": 100.00},
        }
        inactivity = {"C101": 60, "C102": 90, "C103": 10}
        risks = filter_high_value_churn_risks(metrics, inactivity, min_spend=100.0, max_activity_days=30)
        self.assertEqual(risks, ["C101"])

    def test_merge_customer_profiles_matching(self):
        p_a = ("C101", ["vip", "early_adopter"], {"email_opt_in": True, "theme": "light"})
        p_b = ("C101", ["mobile", "vip"], {"theme": "dark", "sms_opt_in": False})
        merged = merge_customer_profiles(p_a, p_b)
        expected_tags = ["early_adopter", "mobile", "vip"]
        expected_pref = {"email_opt_in": True, "theme": "dark", "sms_opt_in": False}
        self.assertEqual(merged, ("C101", expected_tags, expected_pref))

    def test_merge_customer_profiles_mismatch(self):
        p_a = ("C101", ["vip"], {})
        p_b = ("C102", ["vip"], {})
        self.assertIsNone(merge_customer_profiles(p_a, p_b))


if __name__ == "__main__":
    unittest.main()
