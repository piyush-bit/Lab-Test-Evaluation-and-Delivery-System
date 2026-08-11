# Lab 04: Machine Learning Pipeline & Hyperparameter Optimizer

## Background
In modern machine learning workflows, building flexible and reusable pipelines requires leveraging Python's rich function capabilities. This lab focuses on functional programming constructs in Python, including user-defined functions, default argument values, variable positional arguments (`*args`), variable keyword arguments (`**kwargs`), lambda functions, and higher-order functions/closures.

You will implement utility functions for a machine learning experiment pipeline in `pipeline_tuner.py`.

---

## Technical Specifications

### 1. `apply_feature_transformation(data: list, transform_fn=None) -> list`
Transforms numeric data using a caller-provided transformation function or a default scaling transformation.

- **Parameters**:
  - `data`: A `list` of numeric values (integers or floats).
  - `transform_fn`: Optional callable function or lambda. Default is `None`.
- **Behavior**:
  - If `transform_fn` is `None`, default to a scaling lambda function: `lambda x: round(x * 1.0, 4)`.
  - Apply `transform_fn` to each item in `data`.
  - Return a list of transformed values as floats rounded to 4 decimal places (`round(..., 4)`).
  - If `data` is empty, return `[]`.

---

### 2. `calculate_pipeline_metrics(y_true: list, y_pred: list, *metric_names) -> dict`
Calculates regression evaluation metrics requested dynamically via `*metric_names`.

- **Parameters**:
  - `y_true`: List of actual numeric target values.
  - `y_pred`: List of predicted numeric target values.
  - `*metric_names`: Variable positional string arguments indicating metrics to compute (case-insensitive).
- **Supported Metrics**:
  - `"mse"`: Mean Squared Error $\text{MSE} = \frac{1}{n} \sum_{i=1}^n (y_i - \hat{y}_i)^2$
  - `"mae"`: Mean Absolute Error $\text{MAE} = \frac{1}{n} \sum_{i=1}^n |y_i - \hat{y}_i|$
  - `"rmse"`: Root Mean Squared Error $\text{RMSE} = \sqrt{\text{MSE}}$
- **Behavior**:
  - If `y_true` and `y_pred` have different lengths or either is empty, return `{}`.
  - Compute each supported metric requested in `*metric_names` (ignoring any unrecognized metric names).
  - Return a dictionary mapping lowercase metric names to calculated values rounded to 4 decimal places.

---

### 3. `build_model_config(model_name: str, **hyperparams) -> dict`
Constructs a model configuration dictionary, validating default hyperparameters and incorporating custom parameters.

- **Parameters**:
  - `model_name`: Model name string (e.g. `"xgboost"`).
  - `**hyperparams`: Variable keyword arguments representing model parameters.
- **Default Hyperparameters**:
  - `learning_rate`: `0.01`
  - `max_depth`: `5`
  - `use_gpu`: `False`
- **Behavior**:
  - Convert `model_name` to uppercase string.
  - Override default hyperparameters with any provided in `**hyperparams`, and add any extra hyperparameter key-value pairs.
  - Return a dict with structure:
    ```python
    {
        "model_name": str,      # Uppercase string
        "hyperparams": dict     # Validated dictionary of hyperparameters
    }
    ```

---

### 4. `create_learning_rate_scheduler(initial_lr: float, decay_rate: float)`
Creates an exponential learning rate decay scheduler closure.

- **Parameters**:
  - `initial_lr`: Initial learning rate (`float`).
  - `decay_rate`: Exponential decay rate parameter (`float`).
- **Behavior**:
  - Returns a step closure function `step(epoch: int) -> float`.
  - The returned function computes: $\text{lr} = \text{initial\_lr} \times e^{-\text{decay\_rate} \times \text{epoch}}$ using `math.exp`.
  - Returns decay-adjusted learning rate rounded to 6 decimal places (`round(..., 6)`).

---

## Directory Structure
```
cs1287-lab04-hyperparameter-functions/
├── manifest.json              (Machine-readable lab metadata & evaluation rules)
├── README.md                  (Problem description & instructions)
├── Makefile                   (Build & test commands)
├── run                        (Executable wrapper script)
├── pipeline_tuner.py          (Student starter code)
├── public_test.py             (Public test suite)
├── tests_private/
│   └── test_private.py        (Private test suite for grading)
└── reference/
    └── pipeline_tuner.py      (Instructor reference solution)
```

---

## Execution & Testing
Run public tests locally:
```bash
./run public
# or
make test-public
```

Clean temporary files:
```bash
./run clean
```
