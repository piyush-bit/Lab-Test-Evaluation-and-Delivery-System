# CS1287 Lab 07: NLP Text Cleaning & Tokenization Engine (Strings)

## Overview
In natural language processing (NLP) pipelines, raw text data must be preprocessed and sanitized before downstream feature extraction or modeling. In this lab, you will implement a core string preprocessing library in Python to perform text cleaning, stopword filtering, n-gram generation, and PII (Personally Identifiable Information) masking.

## Problem Specification

You are required to complete the implementation of functions in `text_preprocessor.py`:

### 1. `clean_text(text: str) -> str`
- Converts the input `text` to lowercase.
- Replaces all punctuation characters (`!,.:;?#$@"'-`) with spaces.
- Normalizes internal multiple spaces (and whitespace) into a single space.
- Strips leading and trailing whitespace.
- Returns the cleaned string.

**Example:**
```python
clean_text("  Hello, World!!  How are you? -- fine.  ")
# Output: "hello world how are you fine"
```

---

### 2. `tokenize_and_remove_stopwords(text: str, stopwords: list) -> list`
- First cleans `text` using `clean_text`.
- Splits the cleaned text into tokens (words).
- Filters out tokens present in the `stopwords` list using case-insensitive matching.
- Returns the list of remaining token strings.

**Example:**
```python
tokenize_and_remove_stopwords("The quick brown Fox jumps over the lazy Dog!", ["the", "OVER", "a"])
# Output: ["quick", "brown", "fox", "jumps", "lazy", "dog"]
```

---

### 3. `generate_ngrams(tokens: list, n: int) -> list`
- Accepts a list of token strings and an integer `n`.
- Generates a list of n-gram tuple strings joined by space.
- If `len(tokens) < n` or `n < 1`, returns an empty list `[]`.

**Example:**
```python
generate_ngrams(["machine", "learning", "model"], 2)
# Output: ["machine learning", "learning model"]

generate_ngrams(["a", "b", "c", "d"], 3)
# Output: ["a b c", "b c d"]
```

---

### 4. `mask_sensitive_pii(text: str) -> str`
- Sanitizes sensitive PII patterns in the string:
  - **Email addresses** (e.g. `user@example.com`): replaces the username portion before `@` with `***`, keeping `@domain.com` intact (e.g. `***@example.com`).
  - **Phone numbers** (10 digits e.g. `9876543210`): replaces the middle 6 digits with `XXXXXX`, retaining first 2 and last 2 digits (e.g. `98XXXXXX10`).
- Returns the sanitized text string.

**Example:**
```python
mask_sensitive_pii("Contact Alice at alice@company.org or call 9876543210 for details.")
# Output: "Contact Alice at ***@company.org or call 98XXXXXX10 for details."
```

---

## Directory Structure
```
cs1287-lab07-text-preprocessor-string/
├── manifest.json              # Exercise metadata & grading schema
├── README.md                  # Problem description and guide
├── Makefile                   # Local testing & grading automation
├── run                        # Executable CLI wrapper script
├── text_preprocessor.py       # Starter code for student implementation
├── public_test.py             # Public test suite
├── tests_private/
│   └── test_private.py        # Private test suite (evaluator only)
└── reference/
    └── text_preprocessor.py   # Instructor reference solution
```

## Running Tests
Run the public test suite locally using the wrapper script or Makefile:

```bash
./run public
# or
make test-public
```

Clean up cache files:
```bash
./run clean
# or
make clean
```
