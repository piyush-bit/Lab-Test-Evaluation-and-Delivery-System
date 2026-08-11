.PHONY: test-public \
        test-submission-fit-attributes \
        test-submission-transform-standard \
        test-submission-fit-transform \
        test-submission-inverse-transform \
        test-submission-partial-scaling-flags \
        clean

# Run public test suite against student's feature_scaler.py
test-public:
	@python3 -m unittest public_test.py

# ---------------------------------------------------------------------------
# Grading targets — one per private test group (Strategy B: Inject & Restore)
# ---------------------------------------------------------------------------

define _grade
	@cp feature_scaler.py feature_scaler.py.bak
	@cp reference/feature_scaler.py feature_scaler.py
	@cp tests_private/test_private.py test_private_runner.py
	@python3 -m unittest test_private_runner.py -k $(1); EXIT=$$?; \
	  cp feature_scaler.py.bak feature_scaler.py; \
	  rm -f feature_scaler.py.bak test_private_runner.py; \
	  exit $$EXIT
endef

test-submission-fit-attributes:
	$(call _grade,test_private_fit_attributes)

test-submission-transform-standard:
	$(call _grade,test_private_transform_standard)

test-submission-fit-transform:
	$(call _grade,test_private_fit_transform)

test-submission-inverse-transform:
	$(call _grade,test_private_inverse_transform)

test-submission-partial-scaling-flags:
	$(call _grade,test_private_partial_scaling_flags)

clean:
	@rm -rf __pycache__ .pytest_cache *.pyc feature_scaler.py.bak test_private_runner.py
