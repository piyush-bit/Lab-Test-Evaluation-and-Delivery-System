"""
CS1287: Machine Learning Using Python Lab
Experiment 1: Basics of Python Programming Language

Task: Smart Grid Telemetry Quality & Tariff Auditor
"""

def compute_apparent_power(voltage: float, current: float) -> float:
    """
    Calculates apparent electrical power S in kVA:
    Apparent Power (kVA) = (voltage * current) / 1000
    
    Returns result rounded to 2 decimal places.
    If voltage < 0 or current < 0, return -1.0.
    """
    # TODO: Implement apparent power calculation with validation and rounding
    return 0.0


def classify_voltage_status(voltage: float) -> str:
    """
    Classifies nominal 230V single-phase voltage status:
    - 'INVALID': voltage <= 0.0
    - 'UNDER_VOLTAGE': 0.0 < voltage < 207.0
    - 'NORMAL': 207.0 <= voltage <= 253.0
    - 'OVER_VOLTAGE': voltage > 253.0
    """
    # TODO: Implement voltage classification logic using conditional statements
    return "INVALID"


def calculate_tariff_bill(units_kwh: float, customer_type: str) -> float:
    """
    Computes energy billing total based on consumed units (kWh) and customer type.
    
    Residential Pricing:
      - First 100 kWh: $0.10 / kWh
      - Next 200 kWh (101-300 kWh): $0.15 / kWh
      - Above 300 kWh: $0.20 / kWh
      - Minimum Base Charge: $5.00
      
    Commercial Pricing:
      - Flat Rate: $0.25 / kWh
      - Tax Surcharge: 15% added to usage bill
      - Minimum Base Charge: $25.00
      
    Validation:
      - If units_kwh < 0 or customer_type not in ["RESIDENTIAL", "COMMERCIAL"], return -1.0.
      
    Returns final bill rounded to 2 decimal places.
    """
    # TODO: Implement tariff calculation for residential and commercial customers
    return -1.0


def evaluate_telemetry_quality(voltage: float, power_factor: float, frequency: float) -> dict:
    """
    Evaluates telemetry data quality score (0 to 100) and ML readiness flag.
    Returns: {"score": int, "is_valid_for_ml": bool}
    
    Base Score: 100
    Penalties:
      - Voltage: INVALID -> -100; UNDER_VOLTAGE or OVER_VOLTAGE -> -30
      - Power Factor: PF < 0.0 or PF > 1.0 -> -100; 0.0 <= PF < 0.85 -> -25
      - Frequency: outside [49.5, 50.5] Hz -> -20
      
    Score is clamped at 0 minimum.
    is_valid_for_ml is True if score >= 70 else False.
    """
    # TODO: Implement telemetry quality evaluation score and ML validity flag
    return {"score": 0, "is_valid_for_ml": False}
