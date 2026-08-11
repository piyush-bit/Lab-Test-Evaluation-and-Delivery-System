.PHONY: test-public \
        test-submission-create-dataframe \
        test-submission-category-metrics \
        test-submission-plot-revenue \
        test-submission-plot-rating \
        test-submission-eda-integration \
        clean

# Run the public test suite against student's eda_dashboard.py
test-public:
	@python3 -m unittest public_test.py

# ---------------------------------------------------------------------------
# Grading targets — one per private test group (Strategy B: Inject & Restore)
# ---------------------------------------------------------------------------

define _grade
	@cp eda_dashboard.py eda_dashboard.py.bak
	@cp reference/eda_dashboard.py eda_dashboard.py
	@cp tests_private/test_private.py test_private_runner.py
	@python3 -m unittest test_private_runner.py -k $(1); EXIT=$$?; \
	  cp eda_dashboard.py.bak eda_dashboard.py; \
	  rm -f eda_dashboard.py.bak test_private_runner.py; \
	  exit $$EXIT
endef

test-submission-create-dataframe:
	$(call _grade,test_private_create_dataframe)

test-submission-category-metrics:
	$(call _grade,test_private_category_metrics)

test-submission-plot-revenue:
	$(call _grade,test_private_plot_revenue)

test-submission-plot-rating:
	$(call _grade,test_private_plot_rating)

test-submission-eda-integration:
	$(call _grade,test_private_eda_integration)

clean:
	@rm -rf __pycache__ .pytest_cache *.pyc eda_dashboard.py.bak test_private_runner.py *.png
