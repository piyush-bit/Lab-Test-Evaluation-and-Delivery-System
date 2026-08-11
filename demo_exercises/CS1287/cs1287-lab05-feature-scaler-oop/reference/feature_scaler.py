import math


class StandardScaler:
    """
    StandardScaler standardizes features by removing the mean and scaling
    to unit variance using population standard deviation.
    """
    def __init__(self, with_mean: bool = True, with_std: bool = True):
        self.with_mean = with_mean
        self.with_std = with_std
        self.mean_ = None
        self.std_ = None
        self.is_fitted_ = False

    def fit(self, data: list) -> 'StandardScaler':
        """
        Compute the mean and standard deviation of the input data.
        """
        if not data:
            raise ValueError("Data cannot be empty")

        n = len(data)
        self.mean_ = float(sum(data) / n)

        variance = sum((x - self.mean_) ** 2 for x in data) / n
        std = math.sqrt(variance)

        if std == 0.0:
            self.std_ = 1.0
        else:
            self.std_ = float(std)

        self.is_fitted_ = True
        return self

    def transform(self, data: list) -> list:
        """
        Perform standardization by centering and scaling the input data.
        """
        if not self.is_fitted_:
            raise RuntimeError("Scaler must be fitted before transform")

        res = []
        for x in data:
            val = float(x)
            if self.with_mean:
                val -= self.mean_
            if self.with_std:
                val /= self.std_
            res.append(round(val, 4))
        return res

    def fit_transform(self, data: list) -> list:
        """
        Fit to data, then transform it.
        """
        return self.fit(data).transform(data)

    def inverse_transform(self, scaled_data: list) -> list:
        """
        Scale back the data to original representation.
        """
        if not self.is_fitted_:
            raise RuntimeError("Scaler must be fitted before transform")

        res = []
        for y in scaled_data:
            val = float(y)
            if self.with_std:
                val *= self.std_
            if self.with_mean:
                val += self.mean_
            res.append(round(val, 4))
        return res
