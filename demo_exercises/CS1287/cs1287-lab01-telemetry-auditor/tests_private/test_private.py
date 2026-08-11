import unittest
import telemetry_auditor

class TestPrivateTelemetryAuditor(unittest.TestCase):

    def test_private_apparent_power(self):
        """Private 1: Comprehensive Apparent Power testing (2 pts)"""
        # Exact floating point precision & rounding
        self.assertEqual(telemetry_auditor.compute_apparent_power(235.45, 12.34), 2.91)
        self.assertEqual(telemetry_auditor.compute_apparent_power(0.0, 100.0), 0.0)
        self.assertEqual(telemetry_auditor.compute_apparent_power(230.0, 0.0), 0.0)
        
        # Negative parameters validation
        self.assertEqual(telemetry_auditor.compute_apparent_power(-230.0, 10.0), -1.0)
        self.assertEqual(telemetry_auditor.compute_apparent_power(230.0, -5.0), -1.0)
        self.assertEqual(telemetry_auditor.compute_apparent_power(-100.0, -10.0), -1.0)

    def test_private_voltage_status(self):
        """Private 2: Voltage classification boundary testing (2 pts)"""
        # Edge boundary values
        self.assertEqual(telemetry_auditor.classify_voltage_status(0.0), "INVALID")
        self.assertEqual(telemetry_auditor.classify_voltage_status(-10.0), "INVALID")
        self.assertEqual(telemetry_auditor.classify_voltage_status(0.001), "UNDER_VOLTAGE")
        self.assertEqual(telemetry_auditor.classify_voltage_status(206.99), "UNDER_VOLTAGE")
        self.assertEqual(telemetry_auditor.classify_voltage_status(207.0), "NORMAL")
        self.assertEqual(telemetry_auditor.classify_voltage_status(230.0), "NORMAL")
        self.assertEqual(telemetry_auditor.classify_voltage_status(253.0), "NORMAL")
        self.assertEqual(telemetry_auditor.classify_voltage_status(253.01), "OVER_VOLTAGE")
        self.assertEqual(telemetry_auditor.classify_voltage_status(500.0), "OVER_VOLTAGE")

    def test_private_tariff_residential(self):
        """Private 3: Residential tiered tariff calculation & min base charge (2 pts)"""
        # Minimum base charge ($5.00)
        self.assertEqual(telemetry_auditor.calculate_tariff_bill(0.0, "RESIDENTIAL"), 5.00)
        self.assertEqual(telemetry_auditor.calculate_tariff_bill(10.0, "RESIDENTIAL"), 5.00)  # 10 * 0.10 = $1.00 -> min $5.00

        # Tier 1 (up to 100 kWh)
        self.assertEqual(telemetry_auditor.calculate_tariff_bill(80.0, "RESIDENTIAL"), 8.00)
        self.assertEqual(telemetry_auditor.calculate_tariff_bill(100.0, "RESIDENTIAL"), 10.00)

        # Tier 2 (101 - 300 kWh)
        # 200 kWh -> 100*0.10 + 100*0.15 = 10 + 15 = 25.00
        self.assertEqual(telemetry_auditor.calculate_tariff_bill(200.0, "RESIDENTIAL"), 25.00)
        self.assertEqual(telemetry_auditor.calculate_tariff_bill(300.0, "RESIDENTIAL"), 40.00)  # 10 + 30 = 40.00

        # Tier 3 (> 300 kWh)
        # 400 kWh -> 10 + 30 + 100*0.20 = 60.00
        self.assertEqual(telemetry_auditor.calculate_tariff_bill(400.0, "RESIDENTIAL"), 60.00)

    def test_private_tariff_commercial(self):
        """Private 4: Commercial tariff flat rate + 15% surcharge + min charge & invalid handling (2 pts)"""
        # Minimum base charge ($25.00)
        self.assertEqual(telemetry_auditor.calculate_tariff_bill(0.0, "COMMERCIAL"), 25.00)
        self.assertEqual(telemetry_auditor.calculate_tariff_bill(20.0, "COMMERCIAL"), 25.00)  # 20 * 0.25 * 1.15 = 5.75 -> min $25.00

        # Regular Commercial usage
        # 200 kWh -> 200 * 0.25 = 50.0; 50.0 * 1.15 = 57.50
        self.assertEqual(telemetry_auditor.calculate_tariff_bill(200.0, "COMMERCIAL"), 57.50)

        # Invalid customer type or negative units
        self.assertEqual(telemetry_auditor.calculate_tariff_bill(-50.0, "COMMERCIAL"), -1.0)
        self.assertEqual(telemetry_auditor.calculate_tariff_bill(100.0, "residential"), -1.0)  # Case sensitive check
        self.assertEqual(telemetry_auditor.calculate_tariff_bill(100.0, "INDUSTRIAL"), -1.0)

    def test_private_telemetry_quality(self):
        """Private 5: Full scoring matrix, penalty stacking, clamping & ML validity flag (2 pts)"""
        # Perfect telemetry (100 pts)
        r1 = telemetry_auditor.evaluate_telemetry_quality(230.0, 0.95, 50.0)
        self.assertEqual(r1["score"], 100)
        self.assertTrue(r1["is_valid_for_ml"])

        # Multiple penalties: Over voltage (-30), low PF (-25), frequency fault (-20) -> Score: 100 - 75 = 25
        r2 = telemetry_auditor.evaluate_telemetry_quality(260.0, 0.80, 52.0)
        self.assertEqual(r2["score"], 25)
        self.assertFalse(r2["is_valid_for_ml"])

        # Threshold check: Score exactly 70 -> Valid is True
        # Under voltage (-30) -> Score: 70
        r3 = telemetry_auditor.evaluate_telemetry_quality(200.0, 0.90, 50.0)
        self.assertEqual(r3["score"], 70)
        self.assertTrue(r3["is_valid_for_ml"])

        # Threshold check: Score 65 -> Valid is False (low PF -25, frequency fault -20 -> score 55)
        r4 = telemetry_auditor.evaluate_telemetry_quality(230.0, 0.75, 49.0)
        self.assertEqual(r4["score"], 55)
        self.assertFalse(r4["is_valid_for_ml"])

        # Out of bounds PF (-100) + Invalid voltage (-100) -> Clamped at 0 score
        r5 = telemetry_auditor.evaluate_telemetry_quality(-10.0, 1.5, 50.0)
        self.assertEqual(r5["score"], 0)
        self.assertFalse(r5["is_valid_for_ml"])


if __name__ == "__main__":
    unittest.main()
