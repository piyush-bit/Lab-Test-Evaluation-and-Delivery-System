import unittest
from regex_log_parser import (
    extract_valid_ipv4,
    parse_http_access_line,
    detect_sql_injection,
    redact_api_tokens,
)


class TestPublicRegexLogParser(unittest.TestCase):

    def test_extract_valid_ipv4_basic(self):
        log = "Server IP 192.168.1.1 connected to 10.0.0.255. Invalid: 256.1.1.1 and 1.2.3.4.5"
        result = extract_valid_ipv4(log)
        self.assertEqual(result, ["10.0.0.255", "192.168.1.1"])

    def test_parse_http_access_line_basic(self):
        line = '192.168.1.1 - - [10/Aug/2026:13:45:30] "GET /api/v1/predict HTTP/1.1" 200 1024'
        expected = {
            "ip": "192.168.1.1",
            "timestamp": "10/Aug/2026:13:45:30",
            "method": "GET",
            "path": "/api/v1/predict",
            "status_code": 200,
            "response_size": 1024,
        }
        self.assertEqual(parse_http_access_line(line), expected)

    def test_parse_http_access_line_invalid(self):
        invalid_line = 'Malformed log line string without HTTP metadata'
        self.assertEqual(parse_http_access_line(invalid_line), {})

    def test_detect_sql_injection_basic(self):
        log = (
            "INFO User logged in\n"
            "WARN Query: SELECT * FROM users WHERE username = 'admin' OR '1'='1'\n"
            "INFO Page rendered\n"
            "ERROR Query: DROP TABLE accounts;\n"
        )
        detected = detect_sql_injection(log)
        self.assertEqual(len(detected), 2)
        self.assertIn("WARN Query: SELECT * FROM users WHERE username = 'admin' OR '1'='1'", detected)
        self.assertIn("ERROR Query: DROP TABLE accounts;", detected)

    def test_redact_api_tokens_basic(self):
        log = "Fetch data with api_key=secret-key-123 using Authorization: Bearer token_abc.123"
        redacted = redact_api_tokens(log)
        self.assertNotIn("secret-key-123", redacted)
        self.assertNotIn("token_abc.123", redacted)
        self.assertIn("api_key=[REDACTED]", redacted)
        self.assertIn("Bearer [REDACTED]", redacted)


if __name__ == "__main__":
    unittest.main()
