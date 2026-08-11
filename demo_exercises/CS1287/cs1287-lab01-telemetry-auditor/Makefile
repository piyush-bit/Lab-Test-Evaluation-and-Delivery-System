.PHONY: test-public \
        test-submission-apparent-power \
        test-submission-voltage-status \
        test-submission-tariff-residential \
        test-submission-tariff-commercial \
        test-submission-telemetry-quality \
        clean

# Run the public test suite against student's telemetry_auditor.py
test-public:
	@python3 -m unittest public_test.py

# ---------------------------------------------------------------------------
# Grading targets — one per private test group (Strategy B: Inject & Restore)
# ---------------------------------------------------------------------------

define _grade
	@cp telemetry_auditor.py telemetry_auditor.py.bak
	@cp reference/telemetry_auditor.py telemetry_auditor.py
	@cp tests_private/test_private.py test_private_runner.py
	@python3 -m unittest test_private_runner.py -k $(1); EXIT=$$?; \
	  cp telemetry_auditor.py.bak telemetry_auditor.py; \
	  rm -f telemetry_auditor.py.bak test_private_runner.py; \
	  exit $$EXIT
endef

test-submission-apparent-power:
	$(call _grade,test_private_apparent_power)

test-submission-voltage-status:
	$(call _grade,test_private_voltage_status)

test-submission-tariff-residential:
	$(call _grade,test_private_tariff_residential)

test-submission-tariff-commercial:
	$(call _grade,test_private_tariff_commercial)

test-submission-telemetry-quality:
	$(call _grade,test_private_telemetry_quality)

clean:
	@rm -rf __pycache__ .pytest_cache *.pyc telemetry_auditor.py.bak test_private_runner.py
