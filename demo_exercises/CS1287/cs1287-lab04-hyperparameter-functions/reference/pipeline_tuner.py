"""
Reference Solution for CS1287 Lab 04: Machine Learning Pipeline & Hyperparameter Optimizer
"""

import math


def apply_feature_transformation(data: list, transform_fn=None) -> list:
    """
    Transforms numeric list `data` using `transform_fn` (a callable function or lambda).
    If `transform_fn` is None, defaults to `lambda x: round(x * 1.0, 4)`.
    Returns transformed list of floats rounded to 4 decimal places.
    """
    if transform_fn is None:
        transform_fn = lambda x: round(x * 1.0, 4)

    return [round(float(transform_fn(x)), 4) for x in data]


def calculate_pipeline_metrics(y_true: list, y_pred: list, *metric_names) -> dict:
    """
    Computes requested regression metrics specified in `*metric_names` ("mse", "mae", "rmse").
    Returns dict mapping lowercase metric name to calculated value rounded to 4 decimal places.
    If y_true and y_pred have different lengths or are empty, returns {}.
    """
    if not y_true or not y_pred or len(y_true) != len(y_pred):
        return {}

    n = len(y_true)
    results = {}

    for metric in metric_names:
        name = str(metric).lower()
        if name == "mse":
            mse = sum((yt - yp) ** 2 for yt, yp in zip(y_true, y_pred)) / n
            results[name] = round(float(mse), 4)
        elif name == "mae":
            mae = sum(abs(yt - yp) for yt, yp in zip(y_true, y_pred)) / n
            results[name] = round(float(mae), 4)
        elif name == "rmse":
            mse = sum((yt - yp) ** 2 for yt, yp in zip(y_true, y_pred)) / n
            rmse = math.sqrt(mse)
            results[name] = round(float(rmse), 4)

    return results


def build_model_config(model_name: str, **hyperparams) -> dict:
    """
    Builds model configuration dict with model_name (uppercase string) and hyperparams dict.
    Validates hyperparams against defaults:
      - learning_rate: 0.01
      - max_depth: 5
      - use_gpu: False
    Extra hyperparams passed override defaults or add new key-value pairs.
    """
    config_hyperparams = {
        "learning_rate": 0.01,
        "max_depth": 5,
        "use_gpu": False
    }
    config_hyperparams.update(hyperparams)

    return {
        "model_name": model_name.upper(),
        "hyperparams": config_hyperparams
    }


def create_learning_rate_scheduler(initial_lr: float, decay_rate: float):
    """
    Closure returning step(epoch: int) -> float.
    Formula: lr = initial_lr * math.exp(-decay_rate * epoch) rounded to 6 decimal places.
    """
    def step(epoch: int) -> float:
        lr = initial_lr * math.exp(-decay_rate * epoch)
        return round(float(lr), 6)

    return step
