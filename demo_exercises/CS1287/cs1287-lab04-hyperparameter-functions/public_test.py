import unittest
import pipeline_tuner


class TestPublicPipelineTuner(unittest.TestCase):

    def test_apply_feature_transformation_default(self):
        data = [1, 2.5, 3.14159, 4]
        res = pipeline_tuner.apply_feature_transformation(data)
        expected = [1.0, 2.5, 3.1416, 4.0]
        self.assertEqual(res, expected)

    def test_apply_feature_transformation_custom_lambda(self):
        data = [1, 2, 3, 4]
        res = pipeline_tuner.apply_feature_transformation(data, transform_fn=lambda x: x ** 2 + 0.1234)
        expected = [1.1234, 4.1234, 9.1234, 16.1234]
        self.assertEqual(res, expected)

    def test_calculate_pipeline_metrics_standard(self):
        y_true = [1.0, 2.0, 3.0, 4.0]
        y_pred = [1.1, 1.9, 3.2, 3.8]
        metrics = pipeline_tuner.calculate_pipeline_metrics(y_true, y_pred, "mse", "mae", "rmse")
        self.assertIn("mse", metrics)
        self.assertIn("mae", metrics)
        self.assertIn("rmse", metrics)
        self.assertEqual(metrics["mse"], 0.025)
        self.assertEqual(metrics["mae"], 0.15)
        self.assertEqual(metrics["rmse"], 0.1581)

    def test_calculate_pipeline_metrics_invalid(self):
        self.assertEqual(pipeline_tuner.calculate_pipeline_metrics([], [1.0]), {})
        self.assertEqual(pipeline_tuner.calculate_pipeline_metrics([1.0, 2.0], [1.0]), {})

    def test_build_model_config_defaults(self):
        config = pipeline_tuner.build_model_config("xgboost")
        expected = {
            "model_name": "XGBOOST",
            "hyperparams": {
                "learning_rate": 0.01,
                "max_depth": 5,
                "use_gpu": False
            }
        }
        self.assertEqual(config, expected)

    def test_build_model_config_custom_override(self):
        config = pipeline_tuner.build_model_config("random_forest", max_depth=10, n_estimators=100, use_gpu=True)
        expected = {
            "model_name": "RANDOM_FOREST",
            "hyperparams": {
                "learning_rate": 0.01,
                "max_depth": 10,
                "use_gpu": True,
                "n_estimators": 100
            }
        }
        self.assertEqual(config, expected)

    def test_create_learning_rate_scheduler(self):
        scheduler = pipeline_tuner.create_learning_rate_scheduler(0.1, 0.05)
        self.assertTrue(callable(scheduler))
        self.assertEqual(scheduler(0), 0.1)
        self.assertEqual(scheduler(10), 0.060653)


if __name__ == "__main__":
    unittest.main()
