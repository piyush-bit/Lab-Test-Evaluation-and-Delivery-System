"""
CS1287 Lab 02: Customer Churn Analytics & Feature Aggregator

In this lab, you will implement core data processing utilities for customer churn analytics
using Python's fundamental data structures (Lists, Dictionaries, Sets, and Tuples).
"""


def extract_unique_customers(transactions: list) -> set:
    """
    Extracts unique customer IDs from a list of transaction dictionaries.

    :param transactions: List of transaction dictionaries, e.g., [{"customer_id": "C101", "amount": 49.99}, ...]
    :return: A set of unique customer ID strings.
    """
    # TODO: Implement function logic
    return set()


def aggregate_customer_metrics(transactions: list) -> dict:
    """
    Aggregates financial and transaction count metrics for each customer.

    :param transactions: List of transaction dictionaries, e.g., [{"customer_id": "C101", "amount": 49.99}, ...]
    :return: Dictionary mapping customer_id -> {"total_spent": float, "transaction_count": int, "avg_spend": float}.
             Float values must be rounded to 2 decimal places. Returns {} if transactions is empty.
    """
    # TODO: Implement function logic
    return {}


def filter_high_value_churn_risks(metrics: dict, inactivity_days_dict: dict, min_spend: float, max_activity_days: int) -> list:
    """
    Identifies high-value customers who are at risk of churn based on spending and inactivity thresholds.

    :param metrics: Dictionary of metrics per customer (output of aggregate_customer_metrics)
    :param inactivity_days_dict: Dictionary mapping customer_id -> integer days since last activity
    :param min_spend: Minimum total spending threshold (inclusive)
    :param max_activity_days: Inactivity threshold in days (inclusive)
    :return: Alphabetically sorted list of customer IDs meeting total_spent >= min_spend and inactivity_days >= max_activity_days.
    """
    # TODO: Implement function logic
    return []


def merge_customer_profiles(profile_tuple_a: tuple, profile_tuple_b: tuple) -> tuple:
    """
    Merges two profile records for the same customer.

    Profile tuple format: (customer_id, list_of_tags, preference_dict)

    :param profile_tuple_a: First customer profile tuple
    :param profile_tuple_b: Second customer profile tuple
    :return: Merged tuple (customer_id, merged_tags_list, merged_preference_dict)
             where merged_tags_list contains unique tags sorted alphabetically,
             and merged_preference_dict merges preferences (profile_b overwrites profile_a on conflicts).
             Returns None if customer_ids do not match.
    """
    # TODO: Implement function logic
    return ()
