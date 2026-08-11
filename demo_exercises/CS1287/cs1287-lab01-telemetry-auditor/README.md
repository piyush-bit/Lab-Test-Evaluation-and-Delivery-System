# CS1287 Lab 01: Smart Grid Telemetry Quality & Tariff Auditor

## 1. Problem Overview

You are working as a Data Engineer at a Smart Grid Utilities enterprise. Thousands of IoT smart meters stream operational parameters—such as voltage, current, power factor, and frequency—in real time. Before ingesting this telemetry data into downstream Machine Learning models (such as load forecasting and anomaly detection), your team must validate data quality, sanitize invalid readings, and calculate customer billing tariffs based on energy consumption.

In this exercise, you will implement core Python routines to compute grid physics parameters, classify nominal operational statuses, calculate tiered utility tariffs, and evaluate telemetry quality scores.

---

## 2. Requirements & Contract

All functions must be implemented in `telemetry_auditor.py`.

### Task 1: `compute_apparent_power(voltage: float, current: float) -> float`

Calculates the apparent electrical power $S$ in kilovolt-amperes (kVA):
$$\text{Apparent Power (kVA)} = \frac{\text{voltage (V)} \times \text{current (A)}}{1000}$$

* **Return:** Result rounded to 2 decimal places using `round(val, 2)`.
* **Validation:** If `voltage < 0` or `current < 0`, return `-1.0` to signal invalid electrical measurements.

---

### Task 2: `classify_voltage_status(voltage: float) -> str`

Classifies single-phase nominal AC grid voltage status (nominal 230V):

| Voltage Range ($V$) | Status String | Description |
| :--- | :--- | :--- |
| $V \le 0.0$ | `"INVALID"` | Physical fault or disconnected sensor |
| $0.0 < V < 207.0$ | `"UNDER_VOLTAGE"` | Drop greater than 10% below 230V nominal |
| $207.0 \le V \le 253.0$ | `"NORMAL"` | Within standard ±10% nominal tolerance |
| $V > 253.0$ | `"OVER_VOLTAGE"` | Surge greater than 10% above nominal |

---

### Task 3: `calculate_tariff_bill(units_kwh: float, customer_type: str) -> float`

Computes the energy billing total based on consumed units ($kWh$) and tariff category (`"RESIDENTIAL"` or `"COMMERCIAL"`):

#### Residential Tiered Pricing:
* **Tier 1 (First 100 kWh):** $\$0.10 / \text{kWh}$
* **Tier 2 (Next 200 kWh, i.e., 101–300 kWh):** $\$0.15 / \text{kWh}$
* **Tier 3 (Above 300 kWh):** $\$0.20 / \text{kWh}$
* **Minimum Base Charge:** $\$5.00$ (If calculated usage bill $< 5.00$, total bill is $\$5.00$).

#### Commercial Flat Pricing:
* **Flat Rate:** $\$0.25 / \text{kWh}$
* **Commercial Tax Surcharge:** $15\%$ added to the usage bill ($1.15 \times \text{usage}$).
* **Minimum Base Charge:** $\$25.00$ (Enforced after applying tax surcharge).

#### Error Handling:
* If `units_kwh < 0` or `customer_type` is not exactly `"RESIDENTIAL"` or `"COMMERCIAL"`, return `-1.0`.
* Return the final bill rounded to 2 decimal places.

---

### Task 4: `evaluate_telemetry_quality(voltage: float, power_factor: float, frequency: float) -> dict`

Evaluates telemetry quality for Machine Learning ingestion readiness. Returns a dictionary:
```json
{
  "score": int,
  "is_valid_for_ml": bool
}
```

#### Scoring Rules (Base Score = 100):
1. **Voltage Penalties:**
   * If `classify_voltage_status(voltage)` is `"INVALID"`, deduct **100 points**.
   * If `"UNDER_VOLTAGE"` or `"OVER_VOLTAGE"`, deduct **30 points**.
2. **Power Factor (PF) Penalties:**
   * Standard valid PF range is $0.0 \le \text{PF} \le 1.0$.
   * If $\text{PF} < 0.0$ or $\text{PF} > 1.0$, deduct **100 points** (out of bounds).
   * If $0.0 \le \text{PF} < 0.85$ (poor grid efficiency), deduct **25 points**.
3. **Frequency Penalties:**
   * Nominal AC frequency is $50.0 \text{ Hz} \pm 0.5 \text{ Hz}$ ($49.5 \le \text{frequency} \le 50.5$).
   * If frequency is outside $[49.5, 50.5]$, deduct **20 points**.

#### Output Constraints:
* The final score cannot fall below 0 (clamp minimum at `0`).
* `is_valid_for_ml` is `True` if `score >= 70`, otherwise `False`.

---

## 3. Quick Start & Testing

Test your code locally using:

```bash
# Run public test suite
make test-public

# Or using the run wrapper script
./run
```
