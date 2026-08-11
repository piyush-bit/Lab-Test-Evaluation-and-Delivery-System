class StandardScaler:
    """
    Custom Feature Scaler implementing standard Z-score normalization.
    """
    def __init__(self, with_mean: bool = True, with_std: bool = True):
        # TODO: Set instance attributes:
        # self.with_mean = with_mean
        # self.with_std = with_std
        # self.mean_ = None
        # self.std_ = None
        # self.is_fitted_ = False
        pass

    def fit(self, data: list) -> 'StandardScaler':
        """
        Fit scaler to input numeric data by computing mean and population standard deviation.
        """
        # TODO: Raise ValueError("Data cannot be empty") if data is empty
        # TODO: Calculate sample mean and population standard deviation
        # TODO: If std == 0.0, set self.std_ = 1.0; else set self.std_ = std
        # TODO: Set self.mean_ and self.is_fitted_ = True
        # TODO: Return self
        return self

    def transform(self, data: list) -> list:
        """
        Transform data using computed mean and std.
        """
        # TODO: Raise RuntimeError("Scaler must be fitted before transform") if not self.is_fitted_
        # TODO: Apply scaling based on with_mean and with_std flags
        # TODO: Return list of floats rounded to 4 decimal places
        return []

    def fit_transform(self, data: list) -> list:
        """
        Fit scaler and transform data in a single step.
        """
        # TODO: Call self.fit(data) and return self.transform(data)
        return []

    def inverse_transform(self, scaled_data: list) -> list:
        """
        Reverse the scaling operation on scaled_data.
        """
        # TODO: Raise RuntimeError("Scaler must be fitted before transform") if not self.is_fitted_
        # TODO: Reverse scaling (multiply by std_ if with_std, add mean_ if with_mean)
        # TODO: Return list of floats rounded to 4 decimal places
        return []
