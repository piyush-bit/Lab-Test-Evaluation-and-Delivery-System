# CS1287 Lab 10: E-Commerce Sales EDA & Visualization Dashboard

## Problem Overview

In this lab, you will build an Exploratory Data Analysis (EDA) and visualization workflow for e-commerce sales data using Python's **Pandas** and **Matplotlib** libraries.

You will ingest raw transaction records, clean missing numeric data, aggregate performance metrics by product category, and generate standard bar and scatter plot visualizations.

---

## Starter Code & File Structure

Your solution must be implemented in `eda_dashboard.py`.

```
cs1287-lab10-pandas-matplotlib-eda/
├── manifest.json              # Exercise metadata and grading spec
├── README.md                  # Problem description and documentation
├── Makefile                   # Build & test rules
├── run                        # Helper execution script
├── eda_dashboard.py           # Starter file with function stubs (EDIT THIS)
├── public_test.py             # Public unit test suite
├── reference/
│   └── eda_dashboard.py       # Reference solution (for evaluation)
└── tests_private/
    └── test_private.py        # Private unit test suite
```

---

## Required Functions

Implement the following functions in `eda_dashboard.py`:

### 1. `create_sales_dataframe(raw_data: list) -> pandas.DataFrame`
- Ingests a list of dictionary objects representing raw sales records.
- Each dictionary contains: `"order_id"`, `"category"`, `"revenue"`, `"rating"`, and `"units"`.
- Converts the list into a `pandas.DataFrame` with exact columns: `["order_id", "category", "revenue", "rating", "units"]`.
- Fills missing values (`NaN` / `None`):
  - `"revenue"` missing values filled with `0.0`
  - `"rating"` missing values filled with `0.0`
  - `"units"` missing values filled with `0`
- Casts `"revenue"` and `"rating"` columns to `float` and `"units"` to `int`.
- Returns an empty `pandas.DataFrame` with the correct column structure and dtypes if `raw_data` is empty or `None`.

### 2. `compute_category_metrics(df: pandas.DataFrame) -> pandas.DataFrame`
- Groups the input DataFrame by `"category"`.
- Calculates aggregated metrics:
  - `"total_revenue"`: Sum of `"revenue"` rounded to 2 decimal places.
  - `"average_rating"`: Mean of `"rating"` rounded to 2 decimal places.
  - `"total_units"`: Sum of `"units"` (as integer).
- Returns the aggregated DataFrame with reset index containing columns `["category", "total_revenue", "average_rating", "total_units"]`.
- Sorts the resulting DataFrame by `"total_revenue"` in **descending** order.
- Returns an empty DataFrame with expected columns if `df` is empty.

### 3. `plot_revenue_by_category(category_df: pandas.DataFrame, output_image_path: str) -> bool`
- Creates a bar plot using `matplotlib.pyplot` plotting `"total_revenue"` for each `"category"`.
- Sets plot title to `"Total Revenue by Category"`.
- Sets x-axis label to `"Category"` and y-axis label to `"Revenue ($)"`.
- Saves the plot to `output_image_path` using `plt.savefig(output_image_path)`.
- Closes the plot figure with `plt.close()`.
- Returns `True` upon successful completion.

### 4. `plot_rating_vs_revenue(df: pandas.DataFrame, output_image_path: str) -> bool`
- Creates a scatter plot using `matplotlib.pyplot` with `"rating"` on the x-axis and `"revenue"` on the y-axis.
- Sets plot title to `"Rating vs Revenue Scatter"`.
- Sets x-axis label to `"Rating"` and y-axis label to `"Revenue ($)"`.
- Saves the plot to `output_image_path` using `plt.savefig(output_image_path)`.
- Closes the plot figure with `plt.close()`.
- Returns `True` upon successful completion.

---

## Testing & Verification

Run the public test suite locally using Make or the `./run` wrapper:

```bash
# Using helper script
./run public

# Or using make directly
make test-public
```

Clean temporary files with:
```bash
./run clean
```
