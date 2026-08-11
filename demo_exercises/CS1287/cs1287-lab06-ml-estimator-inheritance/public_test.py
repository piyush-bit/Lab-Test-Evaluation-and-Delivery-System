import unittest
from ml_estimators import BaseEstimator, LinearRegressionEstimator, ThresholdClassifierEstimator


class TestPublicMLEstimators(unittest.TestCase):
    """Public unit tests for ML Estimator Hierarchy exercise."""

    def test_base_estimator_interface(self):
        """Test BaseEstimator initialization, get_params, and abstract methods."""
        base = BaseEstimator(model_name="BaseModel", random_state=123)
        params = base.get_params()
        self.assertEqual(params.get("model_name"), "BaseModel")
        self.assertEqual(params.get("random_state"), 123)
        self.assertFalse(params.get("is_fitted"))

        with self.assertRaises(NotImplementedError):
            base.fit([1, 2, 3])

        with self.assertRaises(NotImplementedError):
            base.predict([1, 2, 3])

    def test_linear_regression_fit_and_predict(self):
        """Test LinearRegressionEstimator basic fit and predict with intercept."""
        X = [1.0, 2.0, 3.0, 4.0, 5.0]
        y = [2.0, 4.0, 6.0, 8.0, 10.0]
        est = LinearRegressionEstimator(fit_intercept=True)

        with self.assertRaises(RuntimeError):
            est.predict([1.0, 2.0])

        returned_self = est.fit(X, y)
        self.assertIs(returned_self, est)
        self.assertTrue(est.is_fitted_)
        self.assertAlmostEqual(est.slope_, 2.0, places=4)
        self.assertAlmostEqual(est.intercept_, 0.0, places=4)

        preds = est.predict([1.5, 2.5])
        self.assertEqual(preds, [3.0, 5.0])

    def test_threshold_classifier_fit_and_predict(self):
        """Test ThresholdClassifierEstimator basic fit and predict."""
        X = [0.2, 0.5, 0.8]
        clf = ThresholdClassifierEstimator(threshold=0.5)

        with self.assertRaises(RuntimeError):
            clf.predict(X)

        clf.fit(X)
        self.assertTrue(clf.is_fitted_)

        preds = clf.predict(X)
        self.assertEqual(preds, [0, 1, 1])

    def test_inheritance_hierarchy(self):
        """Test proper inheritance hierarchy and polymorphism."""
        lin = LinearRegressionEstimator()
        clf = ThresholdClassifierEstimator()

        self.assertTrue(isinstance(lin, BaseEstimator))
        self.assertTrue(isinstance(clf, BaseEstimator))
        self.assertEqual(lin.model_name, "LinearRegression")
        self.assertEqual(clf.model_name, "ThresholdClassifier")


if __name__ == "__main__":
    unittest.main()
