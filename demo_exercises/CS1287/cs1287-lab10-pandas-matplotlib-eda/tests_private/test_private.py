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


class TestPrivateCreateDataFrame(unittest.TestCase):

    def test_private_create_dataframe_empty_input(self):
        df_empty = create_sales_dataframe([])
        self.assertIsInstance(df_empty, pd.DataFrame)
        self.assertEqual(list(df_empty.columns), ["order_id", "category", "revenue", "rating", "units"])
        self.assertEqual(len(df_empty), 0)
        self.assertEqual(df_empty["revenue"].dtype, float)
        self.assertEqual(df_empty["rating"].dtype, float)
        self.assertEqual(df_empty["units"].dtype, int)

    def test_private_create_dataframe_missing_keys_and_nulls(self):
        raw = [
            {"order_id": "X1", "category": "Books"},  # missing revenue, rating, units
            {"order_id": "X2", "category": "Books", "revenue": None, "rating": 4.111, "units": None},
            {"order_id": "X3", "category": "Toys", "revenue": 19.99, "rating": None, "units": 5}
        ]
        df = create_sales_dataframe(raw)
        self.assertEqual(len(df), 3)
        self.assertEqual(df.loc[0, "revenue"], 0.0)
        self.assertEqual(df.loc[0, "rating"], 0.0)
        self.assertEqual(df.loc[0, "units"], 0)

        self.assertEqual(df.loc[1, "revenue"], 0.0)
        self.assertEqual(df.loc[1, "units"], 0)

        self.assertEqual(df.loc[2, "rating"], 0.0)
        self.assertEqual(df.loc[2, "revenue"], 19.99)
        self.assertEqual(df.loc[2, "units"], 5)

        self.assertEqual(df["revenue"].dtype, float)
        self.assertEqual(df["rating"].dtype, float)
        self.assertEqual(df["units"].dtype, int)


class TestPrivateCategoryMetrics(unittest.TestCase):

    def test_private_category_metrics_sorting_and_rounding(self):
        raw = [
            {"order_id": "1", "category": "CategoryA", "revenue": 100.126, "rating": 4.555, "units": 2},
            {"order_id": "2", "category": "CategoryA", "revenue": 200.201, "rating": 3.444, "units": 3},
            {"order_id": "3", "category": "CategoryB", "revenue": 500.500, "rating": 5.0, "units": 1},
            {"order_id": "4", "category": "CategoryC", "revenue": 50.000, "rating": 2.0, "units": 10}
        ]
        df = create_sales_dataframe(raw)
        metrics = compute_category_metrics(df)

        self.assertEqual(list(metrics.columns), ["category", "total_revenue", "average_rating", "total_units"])
        self.assertEqual(len(metrics), 3)

        # Descending sort order by total_revenue: CategoryB (500.5), CategoryA (300.33), CategoryC (50.0)
        self.assertEqual(metrics.loc[0, "category"], "CategoryB")
        self.assertEqual(metrics.loc[0, "total_revenue"], 500.5)
        self.assertEqual(metrics.loc[0, "average_rating"], 5.0)

        self.assertEqual(metrics.loc[1, "category"], "CategoryA")
        self.assertEqual(metrics.loc[1, "total_revenue"], 300.33)  # 100.126 + 200.201 = 300.327 -> 300.33
        self.assertEqual(metrics.loc[1, "average_rating"], 4.0)    # (4.555 + 3.444)/2 = 3.9995 -> 4.0
        self.assertEqual(metrics.loc[1, "total_units"], 5)

        self.assertEqual(metrics.loc[2, "category"], "CategoryC")
        self.assertEqual(metrics.loc[2, "total_revenue"], 50.0)

    def test_private_category_metrics_empty(self):
        empty_df = pd.DataFrame()
        res = compute_category_metrics(empty_df)
        self.assertEqual(list(res.columns), ["category", "total_revenue", "average_rating", "total_units"])
        self.assertEqual(len(res), 0)


class TestPrivatePlotRevenue(unittest.TestCase):

    def setUp(self):
        self.temp_dir = tempfile.TemporaryDirectory()

    def tearDown(self):
        self.temp_dir.cleanup()

    def test_private_plot_revenue_execution(self):
        metrics_df = pd.DataFrame({
            "category": ["A", "B"],
            "total_revenue": [150.0, 300.0],
            "average_rating": [4.0, 4.5],
            "total_units": [5, 10]
        })
        out_path = os.path.join(self.temp_dir.name, "bar.png")
        res = plot_revenue_by_category(metrics_df, out_path)
        self.assertTrue(res)
        self.assertTrue(os.path.exists(out_path))
        self.assertGreater(os.path.getsize(out_path), 0)

    def test_private_plot_revenue_empty_df(self):
        empty_metrics = pd.DataFrame(columns=["category", "total_revenue", "average_rating", "total_units"])
        out_path = os.path.join(self.temp_dir.name, "empty_bar.png")
        res = plot_revenue_by_category(empty_metrics, out_path)
        self.assertTrue(res)
        self.assertTrue(os.path.exists(out_path))
        self.assertGreater(os.path.getsize(out_path), 0)


class TestPrivatePlotRating(unittest.TestCase):

    def setUp(self):
        self.temp_dir = tempfile.TemporaryDirectory()

    def tearDown(self):
        self.temp_dir.cleanup()

    def test_private_plot_rating_execution(self):
        df = pd.DataFrame({
            "order_id": ["1", "2"],
            "category": ["A", "B"],
            "revenue": [100.0, 200.0],
            "rating": [3.5, 4.8],
            "units": [1, 2]
        })
        out_path = os.path.join(self.temp_dir.name, "scatter.png")
        res = plot_rating_vs_revenue(df, out_path)
        self.assertTrue(res)
        self.assertTrue(os.path.exists(out_path))
        self.assertGreater(os.path.getsize(out_path), 0)

    def test_private_plot_rating_empty_df(self):
        empty_df = pd.DataFrame(columns=["order_id", "category", "revenue", "rating", "units"])
        out_path = os.path.join(self.temp_dir.name, "empty_scatter.png")
        res = plot_rating_vs_revenue(empty_df, out_path)
        self.assertTrue(res)
        self.assertTrue(os.path.exists(out_path))
        self.assertGreater(os.path.getsize(out_path), 0)


class TestPrivateEDAIntegration(unittest.TestCase):

    def setUp(self):
        self.temp_dir = tempfile.TemporaryDirectory()

    def tearDown(self):
        self.temp_dir.cleanup()

    def test_private_eda_integration_pipeline(self):
        raw = [
            {"order_id": "P1", "category": "Home", "revenue": 45.50, "rating": 4.1, "units": 1},
            {"order_id": "P2", "category": "Electronics", "revenue": 999.99, "rating": 4.8, "units": 1},
            {"order_id": "P3", "category": "Home", "revenue": 104.50, "rating": 4.3, "units": 2},
            {"order_id": "P4", "category": "Electronics", "revenue": 500.00, "rating": 4.6, "units": 3},
            {"order_id": "P5", "category": "Books", "revenue": 15.00, "rating": 4.0, "units": 5}
        ]
        df = create_sales_dataframe(raw)
        self.assertEqual(len(df), 5)

        metrics = compute_category_metrics(df)
        self.assertEqual(len(metrics), 3)
        self.assertEqual(list(metrics["category"]), ["Electronics", "Home", "Books"])
        self.assertAlmostEqual(metrics.loc[0, "total_revenue"], 1499.99)
        self.assertAlmostEqual(metrics.loc[1, "total_revenue"], 150.00)
        self.assertAlmostEqual(metrics.loc[2, "total_revenue"], 15.00)

        bar_file = os.path.join(self.temp_dir.name, "pipeline_bar.png")
        scatter_file = os.path.join(self.temp_dir.name, "pipeline_scatter.png")

        self.assertTrue(plot_revenue_by_category(metrics, bar_file))
        self.assertTrue(plot_rating_vs_revenue(df, scatter_file))

        self.assertTrue(os.path.exists(bar_file))
        self.assertTrue(os.path.exists(scatter_file))


if __name__ == "__main__":
    unittest.main()
