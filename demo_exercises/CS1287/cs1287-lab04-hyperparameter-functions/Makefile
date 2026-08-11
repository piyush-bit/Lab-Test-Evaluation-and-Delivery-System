.PHONY: test-public \
        test-submission-feature-transform \
        test-submission-pipeline-metrics \
        test-submission-model-config \
        test-submission-lr-scheduler \
        clean

# Run public unit tests against student starter code
test-public:
	@python3 -m unittest public_test.py

# ---------------------------------------------------------------------------
# Grading targets — Strategy B: Inject Reference & Restore Starter
# ---------------------------------------------------------------------------

define _grade
	@cp pipeline_tuner.py pipeline_tuner.py.bak
	@cp reference/pipeline_tuner.py pipeline_tuner.py
	@cp tests_private/test_private.py test_private_runner.py
	@python3 -m unittest test_private_runner.py -k $(1); EXIT=$$?; \
	  cp pipeline_tuner.py.bak pipeline_tuner.py; \
	  rm -f pipeline_tuner.py.bak test_private_runner.py; \
	  exit $$EXIT
endef

test-submission-feature-transform:
	$(call _grade,test_private_feature_transform)

test-submission-pipeline-metrics:
	$(call _grade,test_private_pipeline_metrics)

test-submission-model-config:
	$(call _grade,test_private_model_config)

test-submission-lr-scheduler:
	$(call _grade,test_private_lr_scheduler)

clean:
	@rm -rf __pycache__ .pytest_cache *.pyc pipeline_tuner.py.bak test_private_runner.py
