import unittest
import telemetry_auditor

class TestPublicTelemetryAuditor(unittest.TestCase):

    def test_public_compute_apparent_power(self):
        """Test basic apparent power calculation."""
        # 230V, 10A -> (230 * 10) / 1000 = 2.3 kVA
        self.assertEqual(telemetry_auditor.compute_apparent_power(230.0, 10.0), 2.3)
        # Invalid negative voltage
        self.assertEqual(telemetry_auditor.compute_apparent_power(-230.0, 10.0), -1.0)

    def test_public_classify_voltage_status(self):
        """Test classification for nominal, under, over, and invalid voltages."""
        self.assertEqual(telemetry_auditor.classify_voltage_status(230.0), "NORMAL")
        self.assertEqual(telemetry_auditor.classify_voltage_status(190.0), "UNDER_VOLTAGE")
        self.assertEqual(telemetry_auditor.classify_voltage_status(260.0), "OVER_VOLTAGE")
        self.assertEqual(telemetry_auditor.classify_voltage_status(0.0), "INVALID")

    def test_public_calculate_tariff_bill(self):
        """Test residential and commercial tariff calculations."""
        # Residential: 150 kWh -> 100 * 0.10 + 50 * 0.15 = 10.00 + 7.50 = 17.50
        self.assertEqual(telemetry_auditor.calculate_tariff_bill(150.0, "RESIDENTIAL"), 17.50)
        # Commercial: 100 kWh -> 100 * 0.25 * 1.15 = 28.75
        self.assertEqual(telemetry_auditor.calculate_tariff_bill(100.0, "COMMERCIAL"), 28.75)
        # Invalid inputs
        self.assertEqual(telemetry_auditor.calculate_tariff_bill(-10.0, "RESIDENTIAL"), -1.0)
        self.assertEqual(telemetry_auditor.calculate_tariff_bill(100.0, "INVALID_TYPE"), -1.0)

    def test_public_evaluate_telemetry_quality(self):
        """Test telemetry data quality score calculation."""
        # Perfect nominal values (V=230, PF=0.95, Freq=50.0) -> Score: 100, Valid: True
        res_perfect = telemetry_auditor.evaluate_telemetry_quality(230.0, 0.95, 50.0)
        self.assertEqual(res_perfect["score"], 100)
        self.assertTrue(res_perfect["is_valid_for_ml"])

        # Under voltage (-30 points) -> Score: 70, Valid: True
        res_undervoltage = telemetry_auditor.evaluate_telemetry_quality(200.0, 0.95, 50.0)
        self.assertEqual(res_undervoltage["score"], 70)
        self.assertTrue(res_undervoltage["is_valid_for_ml"])


if __name__ == "__main__":
    unittest.main()
