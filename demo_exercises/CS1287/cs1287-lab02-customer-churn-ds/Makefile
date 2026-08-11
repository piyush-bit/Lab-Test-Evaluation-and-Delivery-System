.PHONY: test-public \
        test-submission-extract-unique \
        test-submission-aggregate-metrics \
        test-submission-filter-churn \
        test-submission-merge-profiles \
        clean

# Run the public test suite against student's churn_analyzer.py
test-public:
	@python3 -m unittest public_test.py

# ---------------------------------------------------------------------------
# Grading targets — one per private test group (Strategy B: Inject & Restore)
# ---------------------------------------------------------------------------

define _grade
	@cp churn_analyzer.py churn_analyzer.py.bak
	@cp reference/churn_analyzer.py churn_analyzer.py
	@cp tests_private/test_private.py test_private_runner.py
	@python3 -m unittest test_private_runner.py -k $(1); EXIT=$$?; \
	  cp churn_analyzer.py.bak churn_analyzer.py; \
	  rm -f churn_analyzer.py.bak test_private_runner.py; \
	  exit $$EXIT
endef

test-submission-extract-unique:
	$(call _grade,test_private_extract_unique)

test-submission-aggregate-metrics:
	$(call _grade,test_private_aggregate_metrics)

test-submission-filter-churn:
	$(call _grade,test_private_filter_churn)

test-submission-merge-profiles:
	$(call _grade,test_private_merge_profiles)

clean:
	@rm -rf __pycache__ .pytest_cache *.pyc churn_analyzer.py.bak test_private_runner.py
