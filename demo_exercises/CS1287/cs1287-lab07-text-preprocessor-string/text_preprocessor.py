"""
CS1287 - Lab 07: NLP Text Cleaning & Tokenization Engine (Strings)

Implement the functions below to clean, tokenize, generate n-grams, and sanitize
text data for natural language processing tasks.
"""

def clean_text(text: str) -> str:
    """
    Converts text to lowercase, replaces punctuation characters (!,.:;?#$@"'-) with spaces,
    normalizes multiple internal spaces to a single space, and strips whitespace.
    """
    # TODO: Implement clean_text
    return ""


def tokenize_and_remove_stopwords(text: str, stopwords: list) -> list:
    """
    Cleans text using clean_text, splits into word tokens, and filters out
    stopwords (case-insensitive matching).
    """
    # TODO: Implement tokenize_and_remove_stopwords
    return []


def generate_ngrams(tokens: list, n: int) -> list:
    """
    Generates a list of n-gram tuple strings joined by space from a list of tokens.
    Returns [] if len(tokens) < n or n < 1.
    """
    # TODO: Implement generate_ngrams
    return []


def mask_sensitive_pii(text: str) -> str:
    """
    Masks email addresses (username -> ***) and 10-digit phone numbers
    (middle 6 digits -> XXXXXX) in the input text.
    """
    # TODO: Implement mask_sensitive_pii
    return ""
