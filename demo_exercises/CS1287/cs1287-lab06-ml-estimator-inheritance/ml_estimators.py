"""
Student Starter Code: Machine Learning Estimator Hierarchy (Inheritance)
CS1287 - Lab 06
"""


class BaseEstimator:
    """Base class for all machine learning estimators."""

    def __init__(self, model_name: str, random_state: int = 42):
        """
        Initialize base estimator attributes.
        
        TODO: Store model_name, random_state, and set is_fitted_ to False.
        """
        # TODO: Implement __init__
        self.model_name = ""
        self.random_state = 0
        self.is_fitted_ = False

    def fit(self, X: list, y: list = None):
        """
        Abstract fit method.
        
        TODO: Raise NotImplementedError with an appropriate message.
        """
        # TODO: Implement fit interface method
        pass

    def predict(self, X: list):
        """
        Abstract predict method.
        
        TODO: Raise NotImplementedError with an appropriate message.
        """
        # TODO: Implement predict interface method
        pass

    def get_params(self) -> dict:
        """
        Return parameters and state dictionary.
        
        TODO: Return dict with keys 'model_name', 'random_state', 'is_fitted'.
        """
        # TODO: Implement get_params
        return {}


class LinearRegressionEstimator(BaseEstimator):
    """1D Linear Regression Estimator deriving from BaseEstimator."""

    def __init__(self, fit_intercept: bool = True, random_state: int = 42):
        """
        Initialize Linear Regression estimator.
        
        TODO: Call super().__init__ with "LinearRegression" and random_state.
        TODO: Set fit_intercept, slope_ (default None), and intercept_ (default None).
        """
        # TODO: Implement __init__ using super()
        super().__init__("", 0)
        self.fit_intercept = True
        self.slope_ = None
        self.intercept_ = None

    def fit(self, X: list, y: list):
        """
        Fit 1D Linear Regression model y = m * X + c.
        
        TODO: Compute slope m and intercept c.
        TODO: If fit_intercept is False, c should be 0.0.
        TODO: Set slope_, intercept_, and is_fitted_ = True.
        TODO: Return self.
        """
        # TODO: Implement fit algorithm
        return self

    def predict(self, X: list) -> list:
        """
        Predict target values for input X.
        
        TODO: Raise RuntimeError if is_fitted_ is False.
        TODO: Return list of y_hat values rounded to 4 decimal places.
        """
        # TODO: Implement predict
        return []


class ThresholdClassifierEstimator(BaseEstimator):
    """Binary Threshold Classifier Estimator deriving from BaseEstimator."""

    def __init__(self, threshold: float = 0.5, random_state: int = 42):
        """
        Initialize Threshold Classifier estimator.
        
        TODO: Call super().__init__ with "ThresholdClassifier" and random_state.
        TODO: Store threshold attribute.
        """
        # TODO: Implement __init__ using super()
        super().__init__("", 0)
        self.threshold = 0.5

    def fit(self, X: list, y: list = None):
        """
        Fit Threshold Classifier.
        
        TODO: Set is_fitted_ to True and return self.
        """
        # TODO: Implement fit
        return self

    def predict(self, X: list) -> list:
        """
        Predict binary labels for input X.
        
        TODO: Raise RuntimeError if is_fitted_ is False.
        TODO: Return list of 1s (if x >= threshold) and 0s otherwise.
        """
        # TODO: Implement predict
        return []
