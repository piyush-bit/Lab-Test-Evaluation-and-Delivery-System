"""
Web Security & Log Audit Parser (Regular Expressions)
CS1287 - Lab 08 - Reference Solution
"""

import re


def extract_valid_ipv4(log_text: str) -> list:
    """
    Extracts all unique, valid IPv4 addresses (0.0.0.0 to 255.255.255.255)
    from the provided log text.

    Returns:
        list: Sorted list of unique valid IPv4 address strings.
    """
    pattern = r'(?<!\d)(?<!\d\.)(?:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)\.){3}(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(?!\d|\.\d)'
    matches = re.findall(pattern, log_text)
    return sorted(list(set(matches)))


def parse_http_access_line(log_line: str) -> dict:
    """
    Parses a single Common/Combined Log Format HTTP access log line.
    Format example:
    '192.168.1.1 - - [10/Aug/2026:13:45:30] "GET /api/v1/predict HTTP/1.1" 200 1024'

    Returns:
        dict: Parsed fields with keys 'ip', 'timestamp', 'method', 'path', 'status_code', 'response_size'.
              Returns {} if the log_line does not match the expected format.
    """
    pattern = r'^\s*(\S+)\s+\S+\s+\S+\s+\[([^\]]+)\]\s+"([A-Z]+)\s+(\S+)\s+HTTP/[0-9.]+"\s+(\d+)\s+(\d+)\s*$'
    match = re.match(pattern, log_line)
    if not match:
        return {}
    ip, timestamp, method, path, status_code, response_size = match.groups()
    return {
        "ip": ip,
        "timestamp": timestamp,
        "method": method,
        "path": path,
        "status_code": int(status_code),
        "response_size": int(response_size),
    }


def detect_sql_injection(log_text: str) -> list:
    """
    Scans multiline log text for lines containing potential SQL injection attack signatures.
    Signatures (case-insensitive):
    - OR '1'='1'
    - OR 1=1
    - UNION SELECT
    - DROP TABLE
    - --
    - SELECT * FROM

    Returns:
        list: List of full line strings from log_text that contain any SQL injection signature.
    """
    pattern = r"(?i)(OR\s+['\"]?1['\"]?\s*=\s*['\"]?1['\"]?|UNION\s+SELECT|DROP\s+TABLE|--|SELECT\s+\*\s+FROM)"
    detected = []
    for line in log_text.splitlines():
        if re.search(pattern, line):
            detected.append(line)
    return detected


def redact_api_tokens(log_text: str) -> str:
    """
    Redacts sensitive API tokens in log text.
    Patterns to redact:
    - api_key=[A-Za-z0-9_\-]+ -> replaced with api_key=[REDACTED]
    - Bearer\s+[A-Za-z0-9_\-\.]+ -> replaced with Bearer [REDACTED]

    Returns:
        str: Log text with token values replaced by [REDACTED].
    """
    text = re.sub(r'(api_key=)[A-Za-z0-9_\-]+', r'\1[REDACTED]', log_text)
    text = re.sub(r'(Bearer\s+)[A-Za-z0-9_\-\.]+', r'\1[REDACTED]', text)
    return text
