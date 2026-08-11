import unittest
from feature_scaler import StandardScaler


class TestStandardScalerPrivate(unittest.TestCase):

    # -------------------------------------------------------------------------
    # Target 1: test-submission-fit-attributes
    # -------------------------------------------------------------------------
    def test_private_fit_attributes_basic(self):
        scaler = StandardScaler()
        data = [2.0, 4.0, 4.0, 4.0, 5.0, 5.0, 7.0, 9.0]
        scaler.fit(data)
        self.assertTrue(scaler.is_fitted_)
        self.assertAlmostEqual(scaler.mean_, 5.0, places=4)
        # Mean = 40/8 = 5.0
        # Sq diffs: 9 + 1 + 1 + 1 + 0 + 0 + 4 + 16 = 32
        # Variance = 32 / 8 = 4.0, std = 2.0
        self.assertAlmostEqual(scaler.std_, 2.0, places=4)

    def test_private_fit_attributes_zero_std(self):
        scaler = StandardScaler()
        scaler.fit([7.0, 7.0, 7.0, 7.0])
        self.assertTrue(scaler.is_fitted_)
        self.assertEqual(scaler.mean_, 7.0)
        self.assertEqual(scaler.std_, 1.0)

    def test_private_fit_attributes_empty_data(self):
        scaler = StandardScaler()
        with self.assertRaises(ValueError):
            scaler.fit([])

    # -------------------------------------------------------------------------
    # Target 2: test-submission-transform-standard
    # -------------------------------------------------------------------------
    def test_private_transform_standard_values(self):
        scaler = StandardScaler()
        data = [2.0, 4.0, 4.0, 4.0, 5.0, 5.0, 7.0, 9.0]
        scaler.fit(data)
        transformed = scaler.transform(data)
        # Mean 5.0, Std 2.0
        # (x - 5.0) / 2.0 => [-1.5, -0.5, -0.5, -0.5, 0.0, 0.0, 1.0, 2.0]
        expected = [-1.5, -0.5, -0.5, -0.5, 0.0, 0.0, 1.0, 2.0]
        self.assertEqual(transformed, expected)

    def test_private_transform_standard_unfitted(self):
        scaler = StandardScaler()
        with self.assertRaises(RuntimeError):
            scaler.transform([1.0, 2.0, 3.0])

    # -------------------------------------------------------------------------
    # Target 3: test-submission-fit-transform
    # -------------------------------------------------------------------------
    def test_private_fit_transform_equivalence(self):
        data = [100.0, 200.0, 300.0]
        scaler1 = StandardScaler()
        t1 = scaler1.fit_transform(data)

        scaler2 = StandardScaler()
        scaler2.fit(data)
        t2 = scaler2.transform(data)

        self.assertEqual(t1, t2)
        self.assertTrue(scaler1.is_fitted_)

    # -------------------------------------------------------------------------
    # Target 4: test-submission-inverse-transform
    # -------------------------------------------------------------------------
    def test_private_inverse_transform_reconstruction(self):
        scaler = StandardScaler()
        data = [2.0, 4.0, 4.0, 4.0, 5.0, 5.0, 7.0, 9.0]
        scaled = scaler.fit_transform(data)
        reconstructed = scaler.inverse_transform(scaled)
        self.assertEqual(reconstructed, [2.0, 4.0, 4.0, 4.0, 5.0, 5.0, 7.0, 9.0])

    def test_private_inverse_transform_unfitted(self):
        scaler = StandardScaler()
        with self.assertRaises(RuntimeError):
            scaler.inverse_transform([0.0, 1.0])

    # -------------------------------------------------------------------------
    # Target 5: test-submission-partial-scaling-flags
    # -------------------------------------------------------------------------
    def test_private_partial_scaling_flags_with_mean_false(self):
        scaler = StandardScaler(with_mean=False, with_std=True)
        data = [2.0, 4.0, 4.0, 4.0, 5.0, 5.0, 7.0, 9.0]
        scaler.fit(data)
        # Std is 2.0. Mean is ignored during transform.
        # x / 2.0 => [1.0, 2.0, 2.0, 2.0, 2.5, 2.5, 3.5, 4.5]
        transformed = scaler.transform(data)
        self.assertEqual(transformed, [1.0, 2.0, 2.0, 2.0, 2.5, 2.5, 3.5, 4.5])
        inv = scaler.inverse_transform(transformed)
        self.assertEqual(inv, [2.0, 4.0, 4.0, 4.0, 5.0, 5.0, 7.0, 9.0])

    def test_private_partial_scaling_flags_with_std_false(self):
        scaler = StandardScaler(with_mean=True, with_std=False)
        data = [2.0, 4.0, 4.0, 4.0, 5.0, 5.0, 7.0, 9.0]
        scaler.fit(data)
        # Mean is 5.0. Std is ignored during transform.
        # x - 5.0 => [-3.0, -1.0, -1.0, -1.0, 0.0, 0.0, 2.0, 4.0]
        transformed = scaler.transform(data)
        self.assertEqual(transformed, [-3.0, -1.0, -1.0, -1.0, 0.0, 0.0, 2.0, 4.0])
        inv = scaler.inverse_transform(transformed)
        self.assertEqual(inv, [2.0, 4.0, 4.0, 4.0, 5.0, 5.0, 7.0, 9.0])

    def test_private_partial_scaling_flags_both_false(self):
        scaler = StandardScaler(with_mean=False, with_std=False)
        data = [1.5, 2.5, 3.5]
        scaler.fit(data)
        transformed = scaler.transform(data)
        self.assertEqual(transformed, [1.5, 2.5, 3.5])
        inv = scaler.inverse_transform(transformed)
        self.assertEqual(inv, [1.5, 2.5, 3.5])


if __name__ == "__main__":
    unittest.main()
