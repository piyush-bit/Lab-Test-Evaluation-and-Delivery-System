"""
Reference Solution: Smart Grid Telemetry Quality & Tariff Auditor
CS1287 - Machine Learning Using Python Lab (Experiment 1)
"""

def compute_apparent_power(voltage: float, current: float) -> float:
    if voltage < 0 or current < 0:
        return -1.0
    power_kva = (voltage * current) / 1000.0
    return round(power_kva, 2)


def classify_voltage_status(voltage: float) -> str:
    if voltage <= 0.0:
        return "INVALID"
    elif voltage < 207.0:
        return "UNDER_VOLTAGE"
    elif voltage <= 253.0:
        return "NORMAL"
    else:
        return "OVER_VOLTAGE"


def calculate_tariff_bill(units_kwh: float, customer_type: str) -> float:
    if units_kwh < 0 or customer_type not in ["RESIDENTIAL", "COMMERCIAL"]:
        return -1.0
    
    if customer_type == "RESIDENTIAL":
        if units_kwh <= 100:
            usage_bill = units_kwh * 0.10
        elif units_kwh <= 300:
            usage_bill = (100 * 0.10) + ((units_kwh - 100) * 0.15)
        else:
            usage_bill = (100 * 0.10) + (200 * 0.15) + ((units_kwh - 300) * 0.20)
        
        final_bill = max(usage_bill, 5.00)
        return round(final_bill, 2)
        
    else:  # COMMERCIAL
        usage_bill = units_kwh * 0.25
        total_bill = usage_bill * 1.15  # 15% surcharge
        final_bill = max(total_bill, 25.00)
        return round(final_bill, 2)


def evaluate_telemetry_quality(voltage: float, power_factor: float, frequency: float) -> dict:
    score = 100
    
    # 1. Voltage Penalties
    v_status = classify_voltage_status(voltage)
    if v_status == "INVALID":
        score -= 100
    elif v_status in ["UNDER_VOLTAGE", "OVER_VOLTAGE"]:
        score -= 30
        
    # 2. Power Factor Penalties
    if power_factor < 0.0 or power_factor > 1.0:
        score -= 100
    elif power_factor < 0.85:
        score -= 25
        
    # 3. Frequency Penalties
    if frequency < 49.5 or frequency > 50.5:
        score -= 20
        
    # Clamp score to minimum of 0
    final_score = max(0, score)
    is_valid = final_score >= 70
    
    return {"score": final_score, "is_valid_for_ml": is_valid}
