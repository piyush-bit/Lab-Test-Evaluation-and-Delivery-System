import unittest
import math
import pipeline_tuner


class TestPrivatePipelineTuner(unittest.TestCase):

    # ---------------------------------------------------------------------------
    # Group 1: Feature Transformation Tests (2 points)
    # ---------------------------------------------------------------------------

    def test_private_feature_transform_empty(self):
        res = pipeline_tuner.apply_feature_transformation([])
        self.assertEqual(res, [])

    def test_private_feature_transform_default_scaling(self):
        data = [0.0, 1.23456, -9.87654, 100.00001]
        res = pipeline_tuner.apply_feature_transformation(data)
        self.assertEqual(res, [0.0, 1.2346, -9.8765, 100.0])

    def test_private_feature_transform_custom_log(self):
        data = [1.0, 10.0, 100.0]
        res = pipeline_tuner.apply_feature_transformation(data, transform_fn=lambda x: math.log10(x))
        self.assertEqual(res, [0.0, 1.0, 2.0])

    def test_private_feature_transform_custom_polynomial(self):
        data = [-2.0, -1.0, 0.0, 1.5, 3.0]
        res = pipeline_tuner.apply_feature_transformation(data, transform_fn=lambda x: 2 * (x ** 3) - 0.5 * x + 1.25)
        self.assertEqual(res, [-13.75, -0.25, 1.25, 7.25, 53.75])

    # ---------------------------------------------------------------------------
    # Group 2: Pipeline Metrics Tests (3 points)
    # ---------------------------------------------------------------------------

    def test_private_pipeline_metrics_mismatch_and_empty(self):
        self.assertEqual(pipeline_tuner.calculate_pipeline_metrics([], []), {})
        self.assertEqual(pipeline_tuner.calculate_pipeline_metrics([1, 2, 3], [1, 2]), {})
        self.assertEqual(pipeline_tuner.calculate_pipeline_metrics([1, 2], [1, 2, 3]), {})

    def test_private_pipeline_metrics_case_insensitive_and_unknown(self):
        y_true = [10.0, 20.0, 30.0]
        y_pred = [12.0, 18.0, 33.0]
        # Request uppercase / mixed case and an unknown metric "r2"
        res = pipeline_tuner.calculate_pipeline_metrics(y_true, y_pred, "MSE", "MaE", "r2", "RMSE")
        self.assertIn("mse", res)
        self.assertIn("mae", res)
        self.assertIn("rmse", res)
        self.assertNotIn("r2", res)
        self.assertNotIn("MSE", res)
        # MSE = (4 + 4 + 9) / 3 = 17 / 3 = 5.666666... -> 5.6667
        self.assertEqual(res["mse"], 5.6667)
        # MAE = (2 + 2 + 3) / 3 = 7 / 3 = 2.333333... -> 2.3333
        self.assertEqual(res["mae"], 2.3333)
        # RMSE = sqrt(17/3) = 2.380476... -> 2.3805
        self.assertEqual(res["rmse"], 2.3805)

    def test_private_pipeline_metrics_perfect(self):
        y_true = [5.5, 10.0, -3.2]
        y_pred = [5.5, 10.0, -3.2]
        res = pipeline_tuner.calculate_pipeline_metrics(y_true, y_pred, "mse", "mae", "rmse")
        self.assertEqual(res, {"mse": 0.0, "mae": 0.0, "rmse": 0.0})

    def test_private_pipeline_metrics_subset(self):
        y_true = [1.0, 2.0, 3.0]
        y_pred = [1.5, 2.5, 3.5]
        res = pipeline_tuner.calculate_pipeline_metrics(y_true, y_pred, "rmse")
        self.assertEqual(res, {"rmse": 0.5})

    # ---------------------------------------------------------------------------
    # Group 3: Model Config Tests (2 points)
    # ---------------------------------------------------------------------------

    def test_private_model_config_casing_and_defaults(self):
        config = pipeline_tuner.build_model_config("lightgbm")
        self.assertEqual(config["model_name"], "LIGHTGBM")
        self.assertEqual(config["hyperparams"]["learning_rate"], 0.01)
        self.assertEqual(config["hyperparams"]["max_depth"], 5)
        self.assertEqual(config["hyperparams"]["use_gpu"], False)

    def test_private_model_config_overrides_and_extra(self):
        config = pipeline_tuner.build_model_config(
            "NeuralNet",
            learning_rate=0.001,
            use_gpu=True,
            hidden_layers=[64, 32],
            optimizer="adam"
        )
        self.assertEqual(config["model_name"], "NEURALNET")
        expected_hyperparams = {
            "learning_rate": 0.001,
            "max_depth": 5,
            "use_gpu": True,
            "hidden_layers": [64, 32],
            "optimizer": "adam"
        }
        self.assertEqual(config["hyperparams"], expected_hyperparams)

    # ---------------------------------------------------------------------------
    # Group 4: Learning Rate Scheduler Tests (3 points)
    # ---------------------------------------------------------------------------

    def test_private_lr_scheduler_closure_behavior(self):
        scheduler = pipeline_tuner.create_learning_rate_scheduler(0.05, 0.1)
        self.assertTrue(callable(scheduler))

        # epoch 0: 0.05 * e^0 = 0.05
        self.assertEqual(scheduler(0), 0.05)
        # epoch 1: 0.05 * e^-0.1 = 0.04524187... -> 0.045242
        self.assertEqual(scheduler(1), 0.045242)
        # epoch 5: 0.05 * e^-0.5 = 0.03032653... -> 0.030327
        self.assertEqual(scheduler(5), 0.030327)
        # epoch 20: 0.05 * e^-2.0 = 0.00676676... -> 0.006767
        self.assertEqual(scheduler(20), 0.006767)

    def test_private_lr_scheduler_different_params(self):
        scheduler = pipeline_tuner.create_learning_rate_scheduler(1.0, 0.01)
        # epoch 0: 1.0
        self.assertEqual(scheduler(0), 1.0)
        # epoch 100: 1.0 * e^-1.0 = 0.36787944... -> 0.367879
        self.assertEqual(scheduler(100), 0.367879)


if __name__ == "__main__":
    unittest.main()
