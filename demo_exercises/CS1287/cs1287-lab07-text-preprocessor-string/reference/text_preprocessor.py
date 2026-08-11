"""
CS1287 - Lab 07: NLP Text Cleaning & Tokenization Engine (Strings)
Instructor Reference Solution
"""

import re

PUNCTUATION_CHARS = set('!,.:;?#$@"\'--')

def clean_text(text: str) -> str:
    """
    Converts text to lowercase, replaces punctuation characters (!,.:;?#$@"'-) with spaces,
    normalizes multiple internal spaces to a single space, and strips whitespace.
    """
    if not isinstance(text, str):
        text = str(text)
    
    text = text.lower()
    punct_table = str.maketrans({c: ' ' for c in PUNCTUATION_CHARS})
    cleaned = text.translate(punct_table)
    words = cleaned.split()
    return " ".join(words)


def tokenize_and_remove_stopwords(text: str, stopwords: list) -> list:
    """
    Cleans text using clean_text, splits into word tokens, and filters out
    stopwords (case-insensitive matching).
    """
    cleaned = clean_text(text)
    if not cleaned:
        return []
    
    tokens = cleaned.split(" ")
    stop_set = {str(w).lower() for w in stopwords}
    return [token for token in tokens if token not in stop_set]


def generate_ngrams(tokens: list, n: int) -> list:
    """
    Generates a list of n-gram tuple strings joined by space from a list of tokens.
    Returns [] if len(tokens) < n or n < 1.
    """
    if not tokens or n < 1 or len(tokens) < n:
        return []
    
    ngrams = []
    for i in range(len(tokens) - n + 1):
        ngram = " ".join(tokens[i : i + n])
        ngrams.append(ngram)
    return ngrams


def mask_sensitive_pii(text: str) -> str:
    """
    Masks email addresses (username -> ***) and 10-digit phone numbers
    (middle 6 digits -> XXXXXX) in the input text.
    """
    if not isinstance(text, str):
        text = str(text)
    
    # Mask email username part
    email_pattern = r'\b[A-Za-z0-9._%+-]+@([A-Za-z0-9.-]+\.[A-Za-z]{2,})\b'
    text = re.sub(email_pattern, r'***@\1', text)
    
    # Mask middle 6 digits of 10-digit phone numbers
    phone_pattern = r'\b(\d{2})\d{6}(\d{2})\b'
    text = re.sub(phone_pattern, r'\1XXXXXX\2', text)
    
    return text
