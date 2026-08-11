"""
CS1287 Lab 02: Customer Churn Analytics & Feature Aggregator - Reference Solution
"""


def extract_unique_customers(transactions: list) -> set:
    """
    Extracts unique customer IDs from a list of transaction dictionaries.
    """
    return {tx["customer_id"] for tx in transactions if "customer_id" in tx}


def aggregate_customer_metrics(transactions: list) -> dict:
    """
    Aggregates financial and transaction count metrics for each customer.
    """
    if not transactions:
        return {}

    totals = {}
    counts = {}

    for tx in transactions:
        cid = tx["customer_id"]
        amt = tx["amount"]
        totals[cid] = totals.get(cid, 0.0) + amt
        counts[cid] = counts.get(cid, 0) + 1

    result = {}
    for cid in totals:
        total_spent = round(totals[cid], 2)
        count = counts[cid]
        avg_spend = round(total_spent / count, 2)
        result[cid] = {
            "total_spent": total_spent,
            "transaction_count": count,
            "avg_spend": avg_spend
        }
    return result


def filter_high_value_churn_risks(metrics: dict, inactivity_days_dict: dict, min_spend: float, max_activity_days: int) -> list:
    """
    Identifies high-value customers who are at risk of churn based on spending and inactivity thresholds.
    """
    risks = []
    for cid, data in metrics.items():
        if cid in inactivity_days_dict:
            total_spent = data.get("total_spent", 0.0)
            inactivity = inactivity_days_dict[cid]
            if total_spent >= min_spend and inactivity >= max_activity_days:
                risks.append(cid)
    return sorted(risks)


def merge_customer_profiles(profile_tuple_a: tuple, profile_tuple_b: tuple) -> tuple:
    """
    Merges two profile records for the same customer.
    """
    id_a, tags_a, pref_a = profile_tuple_a
    id_b, tags_b, pref_b = profile_tuple_b

    if id_a != id_b:
        return None

    merged_tags = sorted(list(set(tags_a).union(set(tags_b))))
    merged_pref = dict(pref_a)
    merged_pref.update(pref_b)

    return (id_a, merged_tags, merged_pref)
