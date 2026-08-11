# Experiment 6: Machine Learning Estimator Hierarchy (Inheritance)

**Course:** CS1287  
**Lab ID:** `cs1287-lab06`  
**Language:** Python 3  
**Solution File:** `ml_estimators.py`  

---

## 1. Overview & Objectives

In this lab, you will explore Object-Oriented Programming (OOP) in Python by building a machine learning estimator hierarchy using inheritance, method overriding, and the `super()` function.

### Key Concepts Covered
- Base class declaration and shared state management (`BaseEstimator`).
- Abstract interface design (raising `NotImplementedError`).
- Derived classes (`LinearRegressionEstimator`, `ThresholdClassifierEstimator`).
- Calling parent initializers via `super().__init__(...)`.
- Method overriding (`fit` and `predict`).
- Polymorphic estimator behavior.

---

## 2. Requirements & Class Specifications

You must implement the following classes in `ml_estimators.py`:

### A. `BaseEstimator`
Base class for all estimators in the hierarchy.

- **`__init__(self, model_name: str, random_state: int = 42)`**
  - Sets `self.model_name = model_name`
  - Sets `self.random_state = random_state`
  - Sets `self.is_fitted_ = False`
- **`fit(self, X: list, y: list = None)`**
  - Abstract method. Raises `NotImplementedError("fit method must be implemented by subclasses.")`.
- **`predict(self, X: list)`**
  - Abstract method. Raises `NotImplementedError("predict method must be implemented by subclasses.")`.
- **`get_params(self) -> dict`**
  - Returns dictionary with keys: `{"model_name": self.model_name, "random_state": self.random_state, "is_fitted": self.is_fitted_}`.

### B. `LinearRegressionEstimator(BaseEstimator)`
1D Ordinary Least Squares Linear Regression model deriving from `BaseEstimator`.

- **`__init__(self, fit_intercept: bool = True, random_state: int = 42)`**
  - Invokes `super().__init__("LinearRegression", random_state)`.
  - Sets `self.fit_intercept = fit_intercept`.
  - Sets `self.slope_ = None` and `self.intercept_ = None`.
- **`fit(self, X: list, y: list)`**
  - Fits 1D linear regression model $y = m \cdot X + c$.
  - Computes slope:
    $$m = \frac{\sum (X_i - \bar{X})(y_i - \bar{y})}{\sum (X_i - \bar{X})^2}$$
  - Computes intercept:
    - If `fit_intercept` is `True`: $c = \bar{y} - m \cdot \bar{X}$
    - If `fit_intercept` is `False`: $c = 0.0$
  - Sets `self.slope_ = float(m)`, `self.intercept_ = float(c)`, and `self.is_fitted_ = True`.
  - Raises `ValueError` if $X$ has zero variance or if $X$ and $y$ are empty / unequal in length.
  - Returns `self`.
- **`predict(self, X: list) -> list`**
  - Raises `RuntimeError` if `self.is_fitted_` is `False`.
  - Computes predictions $\hat{y}_i = m \cdot X_i + c$ for each element in $X$.
  - Returns a list of floats rounded to 4 decimal places (`round(m * x + c, 4)`).

### C. `ThresholdClassifierEstimator(BaseEstimator)`
Binary classification model based on a threshold rule deriving from `BaseEstimator`.

- **`__init__(self, threshold: float = 0.5, random_state: int = 42)`**
  - Invokes `super().__init__("ThresholdClassifier", random_state)`.
  - Sets `self.threshold = float(threshold)`.
- **`fit(self, X: list, y: list = None)`**
  - Sets `self.is_fitted_ = True`.
  - Returns `self`.
- **`predict(self, X: list) -> list`**
  - Raises `RuntimeError` if `self.is_fitted_` is `False`.
  - Returns a list of binary integers: `1` if `x >= self.threshold` else `0`.

---

## 3. Testing & Execution Commands

### Local Testing (Student Feedback Loop)
Run public unit tests locally using the convenience script or Makefile:
```bash
./run
# or
make test-public
```

### Clean Temporary Files
```bash
./run clean
# or
make clean
```
