.PHONY: test-public \
        test-submission-base-estimator \
        test-submission-linear-fit \
        test-submission-linear-predict \
        test-submission-threshold-estimator \
        test-submission-inheritance-polymorphism \
        clean

# Run the public test suite against student's ml_estimators.py
test-public:
	@python3 -m unittest public_test.py

# ---------------------------------------------------------------------------
# Grading targets — one per private test group (Strategy B: Inject & Restore)
# ---------------------------------------------------------------------------

define _grade
	@cp ml_estimators.py ml_estimators.py.bak
	@cp reference/ml_estimators.py ml_estimators.py
	@cp tests_private/test_private.py test_private_runner.py
	@python3 -m unittest test_private_runner.py -k $(1); EXIT=$$?; \
	  cp ml_estimators.py.bak ml_estimators.py; \
	  rm -f ml_estimators.py.bak test_private_runner.py; \
	  exit $$EXIT
endef

test-submission-base-estimator:
	$(call _grade,test_private_base_estimator)

test-submission-linear-fit:
	$(call _grade,test_private_linear_fit)

test-submission-linear-predict:
	$(call _grade,test_private_linear_predict)

test-submission-threshold-estimator:
	$(call _grade,test_private_threshold_estimator)

test-submission-inheritance-polymorphism:
	$(call _grade,test_private_inheritance_polymorphism)

clean:
	@rm -rf __pycache__ .pytest_cache *.pyc ml_estimators.py.bak test_private_runner.py
