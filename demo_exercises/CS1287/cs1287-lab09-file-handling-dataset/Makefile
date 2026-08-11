.PHONY: test-public \
        test-submission-read-csv \
        test-submission-filter-abnormal \
        test-submission-export-json \
        test-submission-append-log \
        clean

# Run public tests against student's dataset_file_processor.py
test-public:
	@python3 -m unittest public_test.py

# ---------------------------------------------------------------------------
# Grading targets — one per private test group (Strategy B: Inject & Restore)
# ---------------------------------------------------------------------------

define _grade
	@cp dataset_file_processor.py dataset_file_processor.py.bak
	@cp reference/dataset_file_processor.py dataset_file_processor.py
	@cp tests_private/test_private.py test_private_runner.py
	@python3 -m unittest test_private_runner.py -k $(1); EXIT=$$?; \
	  cp dataset_file_processor.py.bak dataset_file_processor.py; \
	  rm -f dataset_file_processor.py.bak test_private_runner.py; \
	  exit $$EXIT
endef

test-submission-read-csv:
	$(call _grade,test_private_read_csv)

test-submission-filter-abnormal:
	$(call _grade,test_private_filter_abnormal)

test-submission-export-json:
	$(call _grade,test_private_export_json)

test-submission-append-log:
	$(call _grade,test_private_append_log)

clean:
	@rm -rf __pycache__ .pytest_cache *.pyc dataset_file_processor.py.bak test_private_runner.py
