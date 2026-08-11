import unittest
from ml_estimators import BaseEstimator, LinearRegressionEstimator, ThresholdClassifierEstimator


class TestPrivateMLEstimators(unittest.TestCase):
    """Private test suite for ML Estimator Hierarchy exercise."""

    # -------------------------------------------------------------------
    # Group 1: BaseEstimator Tests
    # -------------------------------------------------------------------
    def test_private_base_estimator_init_and_defaults(self):
        est = BaseEstimator(model_name="CustomBase")
        self.assertEqual(est.model_name, "CustomBase")
        self.assertEqual(est.random_state, 42)
        self.assertFalse(est.is_fitted_)

    def test_private_base_estimator_get_params(self):
        est = BaseEstimator(model_name="TestModel", random_state=100)
        params = est.get_params()
        self.assertEqual(params, {
            "model_name": "TestModel",
            "random_state": 100,
            "is_fitted": False
        })
        est.is_fitted_ = True
        params_fitted = est.get_params()
        self.assertTrue(params_fitted["is_fitted"])

    def test_private_base_estimator_abstract_methods(self):
        est = BaseEstimator(model_name="Base")
        with self.assertRaises(NotImplementedError):
            est.fit([1, 2], [3, 4])
        with self.assertRaises(NotImplementedError):
            est.predict([1, 2])

    # -------------------------------------------------------------------
    # Group 2: LinearRegression Fit Tests
    # -------------------------------------------------------------------
    def test_private_linear_fit_with_intercept(self):
        X = [1, 2, 3, 4, 5]
        y = [2, 4, 5, 4, 5]
        est = LinearRegressionEstimator(fit_intercept=True)
        ret = est.fit(X, y)

        self.assertIs(ret, est)
        self.assertTrue(est.is_fitted_)
        self.assertAlmostEqual(est.slope_, 0.6, places=5)
        self.assertAlmostEqual(est.intercept_, 2.2, places=5)

    def test_private_linear_fit_without_intercept(self):
        X = [1, 2, 3, 4, 5]
        y = [2, 4, 5, 4, 5]
        est = LinearRegressionEstimator(fit_intercept=False)
        est.fit(X, y)

        self.assertTrue(est.is_fitted_)
        self.assertAlmostEqual(est.slope_, 0.6, places=5)
        self.assertEqual(est.intercept_, 0.0)

    def test_private_linear_fit_validation(self):
        est = LinearRegressionEstimator()
        with self.assertRaises(ValueError):
            est.fit([1, 1, 1], [2, 3, 4])

    # -------------------------------------------------------------------
    # Group 3: LinearRegression Predict Tests
    # -------------------------------------------------------------------
    def test_private_linear_predict_unfitted_raises(self):
        est = LinearRegressionEstimator()
        with self.assertRaises(RuntimeError):
            est.predict([1, 2, 3])

    def test_private_linear_predict_precision_rounding(self):
        X = [1, 2, 3, 4, 5]
        y = [2, 4, 5, 4, 5]
        est = LinearRegressionEstimator(fit_intercept=True)
        est.fit(X, y)

        preds = est.predict([1.23456, 2.34567])
        self.assertEqual(preds[0], 2.9407)
        self.assertEqual(preds[1], 3.6074)

    def test_private_linear_predict_no_intercept(self):
        X = [1, 2, 3, 4, 5]
        y = [2, 4, 5, 4, 5]
        est = LinearRegressionEstimator(fit_intercept=False)
        est.fit(X, y)

        preds = est.predict([10.0])
        self.assertEqual(preds, [6.0])

    # -------------------------------------------------------------------
    # Group 4: ThresholdClassifier Tests
    # -------------------------------------------------------------------
    def test_private_threshold_estimator_init(self):
        clf = ThresholdClassifierEstimator(threshold=0.75, random_state=99)
        self.assertEqual(clf.model_name, "ThresholdClassifier")
        self.assertEqual(clf.threshold, 0.75)
        self.assertEqual(clf.random_state, 99)
        self.assertFalse(clf.is_fitted_)

    def test_private_threshold_estimator_unfitted_raises(self):
        clf = ThresholdClassifierEstimator()
        with self.assertRaises(RuntimeError):
            clf.predict([0.1, 0.6])

    def test_private_threshold_estimator_fit_and_predict(self):
        clf = ThresholdClassifierEstimator(threshold=0.5)
        ret = clf.fit([0.1, 0.6])
        self.assertIs(ret, clf)
        self.assertTrue(clf.is_fitted_)

        test_data = [0.0, 0.4999, 0.5, 0.5001, 1.0]
        preds = clf.predict(test_data)
        self.assertEqual(preds, [0, 0, 1, 1, 1])

    # -------------------------------------------------------------------
    # Group 5: Inheritance & Polymorphism Tests
    # -------------------------------------------------------------------
    def test_private_inheritance_polymorphism_hierarchy(self):
        self.assertTrue(issubclass(LinearRegressionEstimator, BaseEstimator))
        self.assertTrue(issubclass(ThresholdClassifierEstimator, BaseEstimator))

    def test_private_inheritance_polymorphism_execution(self):
        estimators = [
            LinearRegressionEstimator(random_state=1),
            ThresholdClassifierEstimator(threshold=0.3, random_state=2)
        ]

        for est in estimators:
            self.assertIsInstance(est, BaseEstimator)
            params = est.get_params()
            self.assertIn("model_name", params)
            self.assertIn("random_state", params)
            self.assertIn("is_fitted", params)
            self.assertFalse(params["is_fitted"])

            est.fit([1.0, 2.0, 3.0], [2.0, 4.0, 6.0])
            self.assertTrue(est.is_fitted_)
            preds = est.predict([1.5, 2.5])
            self.assertEqual(len(preds), 2)


if __name__ == "__main__":
    unittest.main()
