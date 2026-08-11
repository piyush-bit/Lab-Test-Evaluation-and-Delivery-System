# Lab 08: Web Security & Log Audit Parser (Regular Expressions)

## Overview
In web security and infrastructure auditing, parsing log files to identify valid network entities, audit HTTP requests, detect malicious SQL injection attempts, and redact sensitive API tokens is a critical daily task.

In this lab, you will implement four Python functions in `regex_log_parser.py` using the `re` module to parse, validate, detect, and redact information in security and HTTP access logs.

---

## 2. File to Edit & Contract

> **Target File:** `regex_log_parser.py`
>
> All function implementations and logic must be written inside `regex_log_parser.py`.

---

## 3. Functions to Implement

### 1. `extract_valid_ipv4(log_text: str) -> list`
- Scans `log_text` for valid IPv4 addresses (`0.0.0.0` to `255.255.255.255`).
- Each octet must be between `0` and `255`.
- Addresses must not be part of invalid strings like `256.1.1.1` or `1.2.3.4.5`.
- Returns a **sorted list of unique** IPv4 address strings.

### 2. `parse_http_access_line(log_line: str) -> dict`
- Parses a standard HTTP access log line.
- Expected line format example:
  `'192.168.1.1 - - [10/Aug/2026:13:45:30] "GET /api/v1/predict HTTP/1.1" 200 1024'`
- Extracts components using regex matching and returns a dictionary with the following keys:
  - `"ip"` (`str`): Client IP address (e.g., `'192.168.1.1'`)
  - `"timestamp"` (`str`): Timestamp string inside brackets (e.g., `'10/Aug/2026:13:45:30'`)
  - `"method"` (`str`): HTTP method (e.g., `'GET'`, `'POST'`)
  - `"path"` (`str`): Requested resource path (e.g., `'/api/v1/predict'`)
  - `"status_code"` (`int`): HTTP status code (e.g., `200`)
  - `"response_size"` (`int`): Response body size in bytes (e.g., `1024`)
- Returns `{}` if `log_line` does not match the expected format or is malformed.

### 3. `detect_sql_injection(log_text: str) -> list`
- Scans multiline `log_text` for lines containing potential SQL injection attack patterns.
- Case-insensitive signatures to detect:
  - `OR '1'='1'`
  - `OR 1=1`
  - `UNION SELECT`
  - `DROP TABLE`
  - `--`
  - `SELECT * FROM`
- Returns a **list of full line strings** where any SQL injection signature was detected.

### 4. `redact_api_tokens(log_text: str) -> str`
- Replaces sensitive API tokens in `log_text` with `[REDACTED]`.
- Key patterns to redact:
  - `api_key=[A-Za-z0-9_\-]+` $\rightarrow$ `api_key=[REDACTED]`
  - `Bearer\s+[A-Za-z0-9_\-\.]+` $\rightarrow$ `Bearer [REDACTED]`
- Returns the modified log text with tokens sanitized.

---

## Local Verification & Testing

Run public unit tests locally during development:

```bash
make test-public
# or
./run public
```

Clean temporary files:

```bash
make clean
# or
./run clean
```
