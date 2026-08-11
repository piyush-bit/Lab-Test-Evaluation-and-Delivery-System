"""
Reference Solution: Machine Learning Estimator Hierarchy (Inheritance)
CS1287 - Lab 06
"""


class BaseEstimator:
    """Base class for all machine learning estimators."""

    def __init__(self, model_name: str, random_state: int = 42):
        """Initialize base estimator attributes."""
        self.model_name = model_name
        self.random_state = random_state
        self.is_fitted_ = False

    def fit(self, X: list, y: list = None):
        """Abstract fit method to be implemented by derived classes."""
        raise NotImplementedError("fit method must be implemented by subclasses.")

    def predict(self, X: list):
        """Abstract predict method to be implemented by derived classes."""
        raise NotImplementedError("predict method must be implemented by subclasses.")

    def get_params(self) -> dict:
        """Return parameters and state dictionary."""
        return {
            "model_name": self.model_name,
            "random_state": self.random_state,
            "is_fitted": self.is_fitted_
        }


class LinearRegressionEstimator(BaseEstimator):
    """1D Linear Regression Estimator deriving from BaseEstimator."""

    def __init__(self, fit_intercept: bool = True, random_state: int = 42):
        """Initialize Linear Regression estimator."""
        super().__init__("LinearRegression", random_state)
        self.fit_intercept = fit_intercept
        self.slope_ = None
        self.intercept_ = None

    def fit(self, X: list, y: list):
        """
        Fit 1D Linear Regression model y = m * X + c.
        
        Computes slope m and intercept c.
        Sets slope_, intercept_, and is_fitted_ = True.
        Returns self.
        """
        if not X or not y or len(X) != len(y):
            raise ValueError("X and y must be non-empty lists of equal length.")

        n = len(X)
        mean_x = sum(X) / n
        mean_y = sum(y) / n

        num = sum((X[i] - mean_x) * (y[i] - mean_y) for i in range(n))
        den = sum((X[i] - mean_x) ** 2 for i in range(n))

        if den == 0:
            raise ValueError("Cannot fit linear regression with zero variance in X.")

        m = num / den

        if self.fit_intercept:
            c = mean_y - m * mean_x
        else:
            c = 0.0

        self.slope_ = float(m)
        self.intercept_ = float(c)
        self.is_fitted_ = True

        return self

    def predict(self, X: list) -> list:
        """
        Predict target values for input X.
        
        Raises RuntimeError if model is not fitted.
        Returns list of predicted values rounded to 4 decimal places.
        """
        if not self.is_fitted_:
            raise RuntimeError("Estimator is not fitted yet.")

        m = self.slope_
        c = self.intercept_
        return [round(m * x + c, 4) for x in X]


class ThresholdClassifierEstimator(BaseEstimator):
    """Binary Threshold Classifier Estimator deriving from BaseEstimator."""

    def __init__(self, threshold: float = 0.5, random_state: int = 42):
        """Initialize Threshold Classifier estimator."""
        super().__init__("ThresholdClassifier", random_state)
        self.threshold = float(threshold)

    def fit(self, X: list, y: list = None):
        """
        Fit Threshold Classifier (sets is_fitted_ to True).
        Returns self.
        """
        self.is_fitted_ = True
        return self

    def predict(self, X: list) -> list:
        """
        Predict binary labels for input X.
        
        Returns 1 if x >= threshold else 0.
        Raises RuntimeError if model is not fitted.
        """
        if not self.is_fitted_:
            raise RuntimeError("Estimator is not fitted yet.")

        return [1 if x >= self.threshold else 0 for x in X]
