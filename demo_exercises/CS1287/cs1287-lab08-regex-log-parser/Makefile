.PHONY: test-public \
        test-submission-extract-ipv4 \
        test-submission-parse-http \
        test-submission-detect-sqli \
        test-submission-redact-tokens \
        clean

# Run public tests against student's regex_log_parser.py
test-public:
	@python3 -m unittest public_test.py

# ---------------------------------------------------------------------------
# Grading targets — one per private test group (Strategy B: Inject & Restore)
# ---------------------------------------------------------------------------

define _grade
	@cp regex_log_parser.py regex_log_parser.py.bak
	@cp reference/regex_log_parser.py regex_log_parser.py
	@cp tests_private/test_private.py test_private_runner.py
	@python3 -m unittest test_private_runner.py -k $(1); EXIT=$$?; \
	  cp regex_log_parser.py.bak regex_log_parser.py; \
	  rm -f regex_log_parser.py.bak test_private_runner.py; \
	  exit $$EXIT
endef

test-submission-extract-ipv4:
	$(call _grade,test_private_extract_ipv4)

test-submission-parse-http:
	$(call _grade,test_private_parse_http)

test-submission-detect-sqli:
	$(call _grade,test_private_detect_sqli)

test-submission-redact-tokens:
	$(call _grade,test_private_redact_tokens)

clean:
	@rm -rf __pycache__ .pytest_cache *.pyc regex_log_parser.py.bak test_private_runner.py
