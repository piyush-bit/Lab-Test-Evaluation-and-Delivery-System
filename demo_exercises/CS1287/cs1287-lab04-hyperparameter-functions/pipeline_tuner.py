"""
CS1287 - Lab 04: Machine Learning Pipeline & Hyperparameter Optimizer

Implement the functions below to complete the ML pipeline utility functions.
Refer to README.md for complete specifications and requirements.
"""

import math


def apply_feature_transformation(data: list, transform_fn=None) -> list:
    """
    Transforms numeric list `data` using `transform_fn` (a callable function or lambda).

    :param data: List of numeric values (ints or floats).
    :param transform_fn: Optional callable function/lambda. Default is None.
           If None, default to a lambda scaling values: lambda x: round(x * 1.0, 4).
    :return: List of floats, each rounded to 4 decimal places.
    """
    # TODO: Implement feature transformation logic
    pass


def calculate_pipeline_metrics(y_true: list, y_pred: list, *metric_names) -> dict:
    """
    Computes requested regression metrics specified in *metric_names ("mse", "mae", "rmse").

    :param y_true: List of ground truth float/int targets.
    :param y_pred: List of predicted float/int targets.
    :param metric_names: Variable positional arguments for requested metric names (case-insensitive).
    :return: Dict mapping lowercase metric name to calculated float value rounded to 4 decimal places.
             Returns empty dict {} if y_true and y_pred have different lengths or are empty.
    """
    # TODO: Implement pipeline metric calculation logic
    pass


def build_model_config(model_name: str, **hyperparams) -> dict:
    """
    Builds model configuration dictionary with model_name (uppercase) and validated hyperparams.

    Defaults:
    - learning_rate: 0.01
    - max_depth: 5
    - use_gpu: False

    :param model_name: String name of the model.
    :param hyperparams: Variable keyword arguments overriding defaults or adding extra parameters.
    :return: Dict with keys "model_name" (uppercase str) and "hyperparams" (dict).
    """
    # TODO: Implement model config builder logic
    pass


def create_learning_rate_scheduler(initial_lr: float, decay_rate: float):
    """
    Creates a higher-order learning rate scheduler closure.

    Formula: lr = initial_lr * math.exp(-decay_rate * epoch)

    :param initial_lr: Initial learning rate float value.
    :param decay_rate: Exponential decay rate float value.
    :return: Function `step(epoch: int) -> float` returning learning rate rounded to 6 decimal places.
    """
    # TODO: Implement learning rate scheduler closure
    pass
