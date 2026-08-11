import unittest
from regex_log_parser import (
    extract_valid_ipv4,
    parse_http_access_line,
    detect_sql_injection,
    redact_api_tokens,
)


class TestPrivateRegexLogParser(unittest.TestCase):

    # -------------------------------------------------------------------------
    # 1. extract_valid_ipv4
    # -------------------------------------------------------------------------

    def test_private_extract_ipv4_valid_boundaries_and_sorting(self):
        log = (
            "Connected clients: 172.16.0.1, 10.0.0.1, 0.0.0.0, 255.255.255.255, 172.16.0.1. "
            "Internal Gateway 192.168.1.100."
        )
        res = extract_valid_ipv4(log)
        expected = ["0.0.0.0", "10.0.0.1", "172.16.0.1", "192.168.1.100", "255.255.255.255"]
        self.assertEqual(res, expected)

    def test_private_extract_ipv4_invalid_octets_and_decimals(self):
        log = (
            "Invalid IPs: 256.0.0.1, 192.168.1.999, 1.2.3.4.5, 300.1.1.1, 192.168.1.256. "
            "Valid IP: 127.0.0.1."
        )
        res = extract_valid_ipv4(log)
        self.assertEqual(res, ["127.0.0.1"])

    def test_private_extract_ipv4_no_matches(self):
        log = "No IP addresses here, just numbers 12345 and text 999.888.777.666."
        res = extract_valid_ipv4(log)
        self.assertEqual(res, [])

    # -------------------------------------------------------------------------
    # 2. parse_http_access_line
    # -------------------------------------------------------------------------

    def test_private_parse_http_valid_methods_and_paths(self):
        line1 = '10.0.0.1 - - [11/Aug/2026:08:00:00] "POST /api/v2/users HTTP/1.1" 201 512'
        res1 = parse_http_access_line(line1)
        self.assertEqual(
            res1,
            {
                "ip": "10.0.0.1",
                "timestamp": "11/Aug/2026:08:00:00",
                "method": "POST",
                "path": "/api/v2/users",
                "status_code": 201,
                "response_size": 512,
            },
        )

        line2 = '172.16.2.45 - - [11/Aug/2026:08:05:12] "DELETE /api/v2/items/99 HTTP/2.0" 204 0'
        res2 = parse_http_access_line(line2)
        self.assertEqual(
            res2,
            {
                "ip": "172.16.2.45",
                "timestamp": "11/Aug/2026:08:05:12",
                "method": "DELETE",
                "path": "/api/v2/items/99",
                "status_code": 204,
                "response_size": 0,
            },
        )

    def test_private_parse_http_malformed_variations(self):
        lines = [
            '10.0.0.1 - [11/Aug/2026:08:00:00] "GET /api HTTP/1.1" 200 100',  # missing column
            '10.0.0.1 - - 11/Aug/2026:08:00:00 "GET /api HTTP/1.1" 200 100',   # missing brackets
            '10.0.0.1 - - [11/Aug/2026:08:00:00] GET /api HTTP/1.1 200 100',     # missing quotes
            '10.0.0.1 - - [11/Aug/2026:08:00:00] "GET /api HTTP/1.1" OK 100',    # non-numeric status
            "",
        ]
        for line in lines:
            self.assertEqual(parse_http_access_line(line), {}, f"Failed for line: {line}")

    # -------------------------------------------------------------------------
    # 3. detect_sql_injection
    # -------------------------------------------------------------------------

    def test_private_detect_sqli_all_signatures_case_insensitive(self):
        log = (
            "Line 1: GET /search?q=test HTTP/1.1\n"
            "Line 2: GET /login?user=admin' oR '1'='1 HTTP/1.1\n"
            "Line 3: POST /data payload: 1 OR 1=1\n"
            "Line 4: GET /api?id=1 UnIoN SeLeCt username, password FROM users\n"
            "Line 5: POST /admin action: drOp TabLe logs\n"
            "Line 6: GET /comments?id=5-- HTTP/1.1\n"
            "Line 7: GET /items?q=sElEcT * FrOm products HTTP/1.1\n"
            "Line 8: GET /normal/page HTTP/1.1"
        )
        matches = detect_sql_injection(log)
        self.assertEqual(len(matches), 6)
        self.assertEqual(matches[0], "Line 2: GET /login?user=admin' oR '1'='1 HTTP/1.1")
        self.assertEqual(matches[1], "Line 3: POST /data payload: 1 OR 1=1")
        self.assertEqual(matches[2], "Line 4: GET /api?id=1 UnIoN SeLeCt username, password FROM users")
        self.assertEqual(matches[3], "Line 5: POST /admin action: drOp TabLe logs")
        self.assertEqual(matches[4], "Line 6: GET /comments?id=5-- HTTP/1.1")
        self.assertEqual(matches[5], "Line 7: GET /items?q=sElEcT * FrOm products HTTP/1.1")

    def test_private_detect_sqli_clean_multiline_logs(self):
        log = (
            "INFO 2026-08-11 System status normal\n"
            "DEBUG Processing order #10052\n"
            "INFO User logout completed\n"
        )
        self.assertEqual(detect_sql_injection(log), [])

    # -------------------------------------------------------------------------
    # 4. redact_api_tokens
    # -------------------------------------------------------------------------

    def test_private_redact_tokens_multiple_keys_and_bearers(self):
        log = (
            "Request 1: GET /data?api_key=secret-key-999&user=john\n"
            "Request 2: Authorization: Bearer eyJhbGciOiJIUzI1Ni.abc-123_XYZ\n"
            "Request 3: api_key=another_key-12345 along with Bearer my.token-val\n"
            "Request 4: No tokens present here."
        )
        redacted = redact_api_tokens(log)
        self.assertNotIn("secret-key-999", redacted)
        self.assertNotIn("eyJhbGciOiJIUzI1Ni.abc-123_XYZ", redacted)
        self.assertNotIn("another_key-12345", redacted)
        self.assertNotIn("my.token-val", redacted)

        self.assertIn("api_key=[REDACTED]", redacted)
        self.assertIn("Bearer [REDACTED]", redacted)
        self.assertIn("Request 4: No tokens present here.", redacted)


if __name__ == "__main__":
    unittest.main()
