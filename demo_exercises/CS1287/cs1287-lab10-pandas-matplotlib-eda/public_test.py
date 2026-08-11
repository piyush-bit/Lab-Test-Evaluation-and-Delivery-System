import unittest
import os
import tempfile
import pandas as pd
from eda_dashboard import (
    create_sales_dataframe,
    compute_category_metrics,
    plot_revenue_by_category,
    plot_rating_vs_revenue
)


class TestPublicEDADashboard(unittest.TestCase):

    def setUp(self):
        self.sample_data = [
            {"order_id": "O1", "category": "Electronics", "revenue": 250.0, "rating": 4.5, "units": 2},
            {"order_id": "O2", "category": "Electronics", "revenue": 150.0, "rating": 4.0, "units": 1},
            {"order_id": "O3", "category": "Clothing", "revenue": 80.0, "rating": 4.2, "units": 3},
            {"order_id": "O4", "category": "Clothing", "revenue": 120.0, "rating": 3.8, "units": 2},
            {"order_id": "O5", "category": "Home", "revenue": None, "rating": 4.9, "units": None}
        ]
        self.temp_dir = tempfile.TemporaryDirectory()

    def tearDown(self):
        self.temp_dir.cleanup()

    def test_create_sales_dataframe_basic(self):
        df = create_sales_dataframe(self.sample_data)
        self.assertIsInstance(df, pd.DataFrame)
        self.assertEqual(list(df.columns), ["order_id", "category", "revenue", "rating", "units"])
        self.assertEqual(len(df), 5)
        self.assertEqual(df.loc[4, "revenue"], 0.0)
        self.assertEqual(df.loc[4, "units"], 0)
        self.assertEqual(df["revenue"].dtype, float)
        self.assertEqual(df["rating"].dtype, float)
        self.assertEqual(df["units"].dtype, int)

    def test_compute_category_metrics_basic(self):
        df = create_sales_dataframe(self.sample_data)
        metrics_df = compute_category_metrics(df)
        self.assertIsInstance(metrics_df, pd.DataFrame)
        self.assertEqual(list(metrics_df.columns), ["category", "total_revenue", "average_rating", "total_units"])
        self.assertEqual(len(metrics_df), 3)
        # Top category should be Electronics (total_revenue = 400.0)
        self.assertEqual(metrics_df.loc[0, "category"], "Electronics")
        self.assertEqual(metrics_df.loc[0, "total_revenue"], 400.0)
        self.assertEqual(metrics_df.loc[0, "average_rating"], 4.25)
        self.assertEqual(metrics_df.loc[0, "total_units"], 3)

    def test_plot_revenue_by_category_basic(self):
        df = create_sales_dataframe(self.sample_data)
        metrics_df = compute_category_metrics(df)
        output_path = os.path.join(self.temp_dir.name, "revenue_bar.png")
        result = plot_revenue_by_category(metrics_df, output_path)
        self.assertTrue(result)
        self.assertTrue(os.path.exists(output_path))
        self.assertGreater(os.path.getsize(output_path), 0)

    def test_plot_rating_vs_revenue_basic(self):
        df = create_sales_dataframe(self.sample_data)
        output_path = os.path.join(self.temp_dir.name, "rating_scatter.png")
        result = plot_rating_vs_revenue(df, output_path)
        self.assertTrue(result)
        self.assertTrue(os.path.exists(output_path))
        self.assertGreater(os.path.getsize(output_path), 0)


if __name__ == "__main__":
    unittest.main()
