import unittest
from feature_scaler import StandardScaler


class TestStandardScalerPublic(unittest.TestCase):
    def test_fit_and_attributes(self):
        scaler = StandardScaler()
        self.assertFalse(scaler.is_fitted_)
        self.assertIsNone(scaler.mean_)
        self.assertIsNone(scaler.std_)

        data = [10.0, 20.0, 30.0, 40.0, 50.0]
        scaler.fit(data)

        self.assertTrue(scaler.is_fitted_)
        self.assertAlmostEqual(scaler.mean_, 30.0, places=4)
        # Variance of [10,20,30,40,50] is 200.0, std is sqrt(200) ~ 14.1421356
        self.assertAlmostEqual(scaler.std_, 14.1421356, places=4)

    def test_transform(self):
        scaler = StandardScaler()
        data = [10.0, 20.0, 30.0, 40.0, 50.0]
        scaler.fit(data)
        scaled = scaler.transform(data)

        # Expected: [-1.4142, -0.7071, 0.0, 0.7071, 1.4142]
        expected = [-1.4142, -0.7071, 0.0, 0.7071, 1.4142]
        self.assertEqual(scaled, expected)

    def test_unfitted_raises_runtime_error(self):
        scaler = StandardScaler()
        with self.assertRaises(RuntimeError):
            scaler.transform([1.0, 2.0])

    def test_empty_data_raises_value_error(self):
        scaler = StandardScaler()
        with self.assertRaises(ValueError):
            scaler.fit([])


if __name__ == "__main__":
    unittest.main()
