.PHONY: test-public \
        test-submission-clean-text \
        test-submission-tokenize-stopwords \
        test-submission-generate-ngrams \
        test-submission-mask-pii-email \
        test-submission-mask-pii-phone \
        clean

# Run the public test suite against student's text_preprocessor.py
test-public:
	@python3 -m unittest public_test.py

# ---------------------------------------------------------------------------
# Grading targets — one per private test group (Strategy B: Inject & Restore)
# ---------------------------------------------------------------------------

define _grade
	@cp text_preprocessor.py text_preprocessor.py.bak
	@cp reference/text_preprocessor.py text_preprocessor.py
	@cp tests_private/test_private.py test_private_runner.py
	@python3 -m unittest test_private_runner.py -k $(1); EXIT=$$?; \
	  cp text_preprocessor.py.bak text_preprocessor.py; \
	  rm -f text_preprocessor.py.bak test_private_runner.py; \
	  exit $$EXIT
endef

test-submission-clean-text:
	$(call _grade,test_private_clean_text)

test-submission-tokenize-stopwords:
	$(call _grade,test_private_tokenize_stopwords)

test-submission-generate-ngrams:
	$(call _grade,test_private_generate_ngrams)

test-submission-mask-pii-email:
	$(call _grade,test_private_mask_pii_email)

test-submission-mask-pii-phone:
	$(call _grade,test_private_mask_pii_phone)

clean:
	@rm -rf __pycache__ .pytest_cache *.pyc text_preprocessor.py.bak test_private_runner.py
