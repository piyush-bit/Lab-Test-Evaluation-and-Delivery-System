.PHONY: test-public \
        test-submission-float-casting \
        test-submission-boolean-flags \
        test-submission-timestamp-epoch \
        test-submission-record-sanitization \
        clean

# Run the public test suite against student's log_sanitizer.py
test-public:
	@python3 -m unittest public_test.py

# ---------------------------------------------------------------------------
# Grading targets — one per private test group (Strategy B: Inject & Restore)
# ---------------------------------------------------------------------------

define _grade
	@cp log_sanitizer.py log_sanitizer.py.bak
	@cp reference/log_sanitizer.py log_sanitizer.py
	@cp tests_private/test_private.py test_private_runner.py
	@python3 -m unittest test_private_runner.py -k $(1); EXIT=$$?; \
	  cp log_sanitizer.py.bak log_sanitizer.py; \
	  rm -f log_sanitizer.py.bak test_private_runner.py; \
	  exit $$EXIT
endef

test-submission-float-casting:
	$(call _grade,test_private_float_casting)

test-submission-boolean-flags:
	$(call _grade,test_private_boolean_flags)

test-submission-timestamp-epoch:
	$(call _grade,test_private_timestamp_epoch)

test-submission-record-sanitization:
	$(call _grade,test_private_record_sanitization)

clean:
	@rm -rf __pycache__ .pytest_cache *.pyc log_sanitizer.py.bak test_private_runner.py
