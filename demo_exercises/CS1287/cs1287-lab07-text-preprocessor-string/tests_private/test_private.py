import unittest
from text_preprocessor import (
    clean_text,
    tokenize_and_remove_stopwords,
    generate_ngrams,
    mask_sensitive_pii,
)


class TestPrivateTextPreprocessor(unittest.TestCase):

    # ---------------------------------------------------------------------------
    # Group 1: clean_text
    # ---------------------------------------------------------------------------
    def test_private_clean_text_all_punctuation(self):
        raw = "Hello! This, is a test: Python; is it? #1 $10 @user \"quoted\" 'single' word-hyphen."
        expected = "hello this is a test python is it 1 10 user quoted single word hyphen"
        self.assertEqual(clean_text(raw), expected)

    def test_private_clean_text_empty_and_whitespace(self):
        self.assertEqual(clean_text(""), "")
        self.assertEqual(clean_text("   \n\t   "), "")
        self.assertEqual(clean_text("!?,:;#$@\"'-"), "")

    def test_private_clean_text_multiple_spaces_and_newlines(self):
        raw = "   Data \t Science   \n  and \n\n Artificial   Intelligence  "
        expected = "data science and artificial intelligence"
        self.assertEqual(clean_text(raw), expected)

    def test_private_clean_text_casing(self):
        raw = "PyThOn ProGRamMiNG"
        expected = "python programming"
        self.assertEqual(clean_text(raw), expected)

    # ---------------------------------------------------------------------------
    # Group 2: tokenize_and_remove_stopwords
    # ---------------------------------------------------------------------------
    def test_private_tokenize_stopwords_case_insensitive(self):
        text = "NLP systems are FAST and Efficient!"
        stopwords = ["ARE", "And", "SYSTEMS"]
        expected = ["nlp", "fast", "efficient"]
        self.assertEqual(tokenize_and_remove_stopwords(text, stopwords), expected)

    def test_private_tokenize_stopwords_empty_inputs(self):
        self.assertEqual(tokenize_and_remove_stopwords("", ["the"]), [])
        self.assertEqual(
            tokenize_and_remove_stopwords("Hello world", []), ["hello", "world"]
        )

    def test_private_tokenize_stopwords_all_removed(self):
        text = "The a an in on"
        stopwords = ["the", "A", "AN", "In", "ON"]
        self.assertEqual(tokenize_and_remove_stopwords(text, stopwords), [])

    def test_private_tokenize_stopwords_no_match(self):
        text = "Machine Learning Model"
        stopwords = ["the", "is", "at"]
        expected = ["machine", "learning", "model"]
        self.assertEqual(tokenize_and_remove_stopwords(text, stopwords), expected)

    # ---------------------------------------------------------------------------
    # Group 3: generate_ngrams
    # ---------------------------------------------------------------------------
    def test_private_generate_ngrams_unigrams(self):
        tokens = ["natural", "language", "processing"]
        expected = ["natural", "language", "processing"]
        self.assertEqual(generate_ngrams(tokens, 1), expected)

    def test_private_generate_ngrams_trigrams(self):
        tokens = ["deep", "learning", "neural", "network", "architecture"]
        expected = [
            "deep learning neural",
            "learning neural network",
            "neural network architecture",
        ]
        self.assertEqual(generate_ngrams(tokens, 3), expected)

    def test_private_generate_ngrams_insufficient_tokens(self):
        tokens = ["word1", "word2"]
        self.assertEqual(generate_ngrams(tokens, 3), [])

    def test_private_generate_ngrams_invalid_n(self):
        tokens = ["a", "b", "c"]
        self.assertEqual(generate_ngrams(tokens, 0), [])
        self.assertEqual(generate_ngrams(tokens, -2), [])

    def test_private_generate_ngrams_exact_length(self):
        tokens = ["artificial", "intelligence"]
        expected = ["artificial intelligence"]
        self.assertEqual(generate_ngrams(tokens, 2), expected)

    # ---------------------------------------------------------------------------
    # Group 4: mask_sensitive_pii (Email)
    # ---------------------------------------------------------------------------
    def test_private_mask_pii_email_multiple(self):
        text = "Send queries to support@tech.com or john.doe_99@dev.org."
        expected = "Send queries to ***@tech.com or ***@dev.org."
        self.assertEqual(mask_sensitive_pii(text), expected)

    def test_private_mask_pii_email_subdomains(self):
        text = "Contact admin.user@sub.domain.co.uk immediately."
        expected = "Contact ***@sub.domain.co.uk immediately."
        self.assertEqual(mask_sensitive_pii(text), expected)

    def test_private_mask_pii_email_no_email(self):
        text = "No sensitive email addresses here!"
        self.assertEqual(mask_sensitive_pii(text), text)

    # ---------------------------------------------------------------------------
    # Group 5: mask_sensitive_pii (Phone & Combined)
    # ---------------------------------------------------------------------------
    def test_private_mask_pii_phone_multiple(self):
        text = "Call primary 9876543210 or alternate 1234567890."
        expected = "Call primary 98XXXXXX10 or alternate 12XXXXXX90."
        self.assertEqual(mask_sensitive_pii(text), expected)

    def test_private_mask_pii_phone_non_10_digits(self):
        text = "Order ID 12345 or transaction 12345678901234."
        self.assertEqual(mask_sensitive_pii(text), text)

    def test_private_mask_pii_phone_and_email_combined(self):
        text = "User bob@example.com reached via 9988776655."
        expected = "User ***@example.com reached via 99XXXXXX55."
        self.assertEqual(mask_sensitive_pii(text), expected)


if __name__ == "__main__":
    unittest.main()
