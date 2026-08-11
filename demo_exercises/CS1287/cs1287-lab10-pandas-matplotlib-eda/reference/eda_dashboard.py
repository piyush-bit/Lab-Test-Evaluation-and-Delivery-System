import pandas as pd
import matplotlib
matplotlib.use('Agg')
import matplotlib.pyplot as plt


def create_sales_dataframe(raw_data: list) -> pd.DataFrame:
    """
    Converts raw list of dicts into a Pandas DataFrame.
    Fills missing numeric values with 0.0 or 0 and sets appropriate dtypes.
    """
    cols = ["order_id", "category", "revenue", "rating", "units"]
    if not raw_data:
        df = pd.DataFrame(columns=cols)
        df["revenue"] = df["revenue"].astype(float)
        df["rating"] = df["rating"].astype(float)
        df["units"] = df["units"].astype(int)
        return df

    df = pd.DataFrame(raw_data)
    for col in cols:
        if col not in df.columns:
            df[col] = None

    df = df[cols]
    df["revenue"] = df["revenue"].fillna(0.0).astype(float)
    df["rating"] = df["rating"].fillna(0.0).astype(float)
    df["units"] = df["units"].fillna(0).astype(int)
    return df


def compute_category_metrics(df: pd.DataFrame) -> pd.DataFrame:
    """
    Groups DataFrame by 'category' and aggregates total_revenue, average_rating, total_units.
    Returns DataFrame sorted by total_revenue descending with reset index.
    """
    target_cols = ["category", "total_revenue", "average_rating", "total_units"]
    if df is None or df.empty or "category" not in df.columns:
        return pd.DataFrame(columns=target_cols)

    grouped = df.groupby("category", as_index=False).agg(
        total_revenue=("revenue", "sum"),
        average_rating=("rating", "mean"),
        total_units=("units", "sum")
    )

    grouped["total_revenue"] = grouped["total_revenue"].round(2)
    grouped["average_rating"] = grouped["average_rating"].round(2)
    grouped["total_units"] = grouped["total_units"].astype(int)

    grouped = grouped.sort_values(by="total_revenue", ascending=False).reset_index(drop=True)
    return grouped[target_cols]


def plot_revenue_by_category(category_df: pd.DataFrame, output_image_path: str) -> bool:
    """
    Plots bar chart of total_revenue per category and saves to output_image_path.
    """
    plt.figure()
    if category_df is not None and not category_df.empty:
        plt.bar(category_df["category"].astype(str), category_df["total_revenue"])
    plt.title("Total Revenue by Category")
    plt.xlabel("Category")
    plt.ylabel("Revenue ($)")
    plt.savefig(output_image_path)
    plt.close()
    return True


def plot_rating_vs_revenue(df: pd.DataFrame, output_image_path: str) -> bool:
    """
    Plots scatter chart of rating vs revenue and saves to output_image_path.
    """
    plt.figure()
    if df is not None and not df.empty:
        plt.scatter(df["rating"], df["revenue"])
    plt.title("Rating vs Revenue Scatter")
    plt.xlabel("Rating")
    plt.ylabel("Revenue ($)")
    plt.savefig(output_image_path)
    plt.close()
    return True
