# CS1287 Lab 02: Customer Churn Analytics & Feature Aggregator

## 1. Problem Overview

You are working as a Data Engineer at an e-commerce platform. The marketing and customer retention teams require feature engineering pipelines to identify churn risks and merge user profiles. Raw transaction streams log customer purchasing behaviors, while user activity monitoring tracks days since last platform interaction.

In this exercise, you will implement core Python routines leveraging basic data structures (**Lists**, **Dictionaries**, **Sets**, and **Tuples**) to extract unique customer entities, aggregate aggregate financial metrics, filter high-value churn risks, and consolidate duplicate user profiles.

---

## 2. Requirements & Contract

All functions must be implemented in `churn_analyzer.py`.

### Task 1: `extract_unique_customers(transactions: list) -> set`

Extracts unique customer IDs from a raw list of transaction dictionaries.

* **Input:** `transactions` list where each element is a dictionary, e.g., `[{"customer_id": "C101", "amount": 49.99}, ...]`.
* **Return:** A Python `set` of string customer IDs.
* **Edge Cases:** If `transactions` is empty, return an empty set `set()`.

---

### Task 2: `aggregate_customer_metrics(transactions: list) -> dict`

Aggregates financial spend and transaction counts for each customer across all logged transactions.

* **Input:** `transactions` list of dictionaries containing keys `"customer_id"` and `"amount"`.
* **Return:** A `dict` mapping each `customer_id` (string) to a dictionary with:
  * `"total_spent"` (`float`): Sum of all transaction amounts for the customer, rounded to 2 decimal places.
  * `"transaction_count"` (`int`): Total count of transactions for the customer.
  * `"avg_spend"` (`float`): Average transaction amount (`total_spent / transaction_count`), rounded to 2 decimal places.
* **Edge Cases:** If `transactions` is empty, return an empty dictionary `{}`.

---

### Task 3: `filter_high_value_churn_risks(metrics: dict, inactivity_days_dict: dict, min_spend: float, max_activity_days: int) -> list`

Identifies high-value customers who have been inactive for an extended period and are at risk of churning.

* **Input:**
  * `metrics` (`dict`): Output mapping from `aggregate_customer_metrics`.
  * `inactivity_days_dict` (`dict`): Dict mapping `customer_id` (string) -> `inactivity_days` (`int`).
  * `min_spend` (`float`): Spending threshold.
  * `max_activity_days` (`int`): Inactivity threshold in days.
* **Return:** A Python `list` of customer IDs (strings) sorted in **alphabetical order** where:
  * `total_spent >= min_spend` AND
  * `inactivity_days >= max_activity_days`.
* **Edge Cases:** Only consider customers present in both `metrics` and `inactivity_days_dict`. If no customers qualify, return an empty list `[]`.

---

### Task 4: `merge_customer_profiles(profile_tuple_a: tuple, profile_tuple_b: tuple) -> tuple`

Consolidates two user profile records for the same customer into a single unified profile tuple.

* **Input:** Each profile tuple has the format `(customer_id, list_of_tags, preference_dict)`.
  * `customer_id` (`str`)
  * `list_of_tags` (`list` of strings)
  * `preference_dict` (`dict` of key-value user preferences)
* **Return:**
  * If `customer_id` in `profile_tuple_a` matches `profile_tuple_b`: return a merged tuple `(customer_id, merged_tags_list, merged_preference_dict)` where:
    * `merged_tags_list` is the union of tags from both profiles, formatted as a **sorted list of unique tag strings**.
    * `merged_preference_dict` is the merged dictionary where values from `profile_tuple_b` overwrite `profile_tuple_a` on key conflicts.
  * If `customer_id` does not match between the two profile tuples, return `None`.

---

## 3. Quick Start & Testing

Test your code locally using:

```bash
# Run public test suite
make test-public

# Or using the run wrapper script
./run
```
