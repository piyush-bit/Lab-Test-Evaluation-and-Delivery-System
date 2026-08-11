# CS1287 Lab 05: Custom Feature Scaler & Vector Normalizer (OOP)

## 1. Problem Overview

In machine learning and data engineering pipelines, feature scaling is a fundamental preprocessing step. Unscaled numerical features with varying magnitudes can distort algorithm training, leading to slow convergence or biased model parameters.

In this lab exercise, you will practice **Object-Oriented Programming (OOP)** in Python by building a custom `StandardScaler` class from scratch. Your class will encapsulate scaling configuration parameters (`with_mean`, `with_std`), learned state attributes (`mean_`, `std_`, `is_fitted_`), and transformation logic (`fit`, `transform`, `fit_transform`, `inverse_transform`).

---

## 2. Requirements & Specification

All code must be implemented within `feature_scaler.py`.

### Class: `StandardScaler`

#### Method 1: `__init__(self, with_mean: bool = True, with_std: bool = True)`
* **Description:** Initializes the `StandardScaler` instance.
* **Parameters:**
  * `with_mean` (bool, default `True`): If `True`, center the data before scaling by subtracting the mean.
  * `with_std` (bool, default `True`): If `True`, scale the data to unit variance by dividing by the standard deviation.
* **Instance Attributes Initialized:**
  * `self.with_mean` (bool): Stored configuration.
  * `self.with_std` (bool): Stored configuration.
  * `self.mean_` (float or `None`): Calculated sample mean after fitting; default `None`.
  * `self.std_` (float or `None`): Calculated population standard deviation after fitting; default `None`.
  * `self.is_fitted_` (bool): State flag indicating whether `fit` has been executed; default `False`.

---

#### Method 2: `fit(self, data: list) -> 'StandardScaler'`
* **Description:** Computes the mean $\mu$ and population standard deviation $\sigma$ from `data`.
* **Mathematical Formulas:**
  $$\mu = \frac{1}{N} \sum_{i=1}^{N} x_i$$
  $$\sigma = \sqrt{\frac{1}{N} \sum_{i=1}^{N} (x_i - \mu)^2}$$
* **Behavior & Constraints:**
  * If `data` is empty (`len(data) == 0`), raise `ValueError("Data cannot be empty")`.
  * Store $\mu$ as a float in `self.mean_`.
  * Store $\sigma$ as a float in `self.std_`. If $\sigma == 0.0$, set `self.std_ = 1.0` to prevent division by zero during transformation.
  * Set `self.is_fitted_ = True`.
  * Return `self` (to allow method chaining).

---

#### Method 3: `transform(self, data: list) -> list`
* **Description:** Scales numeric data using stored `mean_` and `std_`.
* **Behavior & Constraints:**
  * If `self.is_fitted_` is `False`, raise `RuntimeError("Scaler must be fitted before transform")`.
  * For each value $x$ in `data`:
    * If `self.with_mean` is `True`: $x' = x - \text{mean\_}$
    * If `self.with_std` is `True`: $x'' = \frac{x'}{\text{std\_}}$
  * Return a list of scaled `float` values, each rounded to **4 decimal places** using `round(val, 4)`.

---

#### Method 4: `fit_transform(self, data: list) -> list`
* **Description:** Fits the scaler to `data` and returns the transformed data in a single step.
* **Behavior:** Calls `self.fit(data)` and returns `self.transform(data)`.

---

#### Method 5: `inverse_transform(self, scaled_data: list) -> list`
* **Description:** Reverses the transformation on `scaled_data` back to original scale.
* **Behavior & Constraints:**
  * If `self.is_fitted_` is `False`, raise `RuntimeError("Scaler must be fitted before transform")`.
  * For each value $y$ in `scaled_data`:
    * If `self.with_std` is `True`: $y' = y \times \text{std\_}$
    * If `self.with_mean` is `True`: $y'' = y' + \text{mean\_}$
  * Return a list of unscaled `float` values rounded to **4 decimal places** using `round(val, 4)`.

---

## 3. Usage Example

```python
from feature_scaler import StandardScaler

scaler = StandardScaler(with_mean=True, with_std=True)

data = [10.0, 20.0, 30.0, 40.0, 50.0]

# Fit and transform
scaled = scaler.fit_transform(data)
print("Scaled:", scaled)
# Output: [-1.4142, -0.7071, 0.0, 0.7071, 1.4142]

# Reconstruct original data
original = scaler.inverse_transform(scaled)
print("Original:", original)
# Output: [10.0, 20.0, 30.0, 40.0, 50.0]
```

---

## 4. Testing & Verification

Run tests using the following commands:

```bash
# Run public unit tests
make test-public

# Run via script wrapper
./run public

# Clean cache and backup artifacts
make clean
```
