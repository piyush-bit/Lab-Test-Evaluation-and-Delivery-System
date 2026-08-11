import unittest
from text_preprocessor import (
    clean_text,
    tokenize_and_remove_stopwords,
    generate_ngrams,
    mask_sensitive_pii,
)


class TestPublicTextPreprocessor(unittest.TestCase):

    def test_clean_text_basic(self):
        raw_text = "  Hello, World!!  How are you? -- fine.  "
        expected = "hello world how are you fine"
        self.assertEqual(clean_text(raw_text), expected)

    def test_tokenize_and_remove_stopwords_basic(self):
        raw_text = "The quick brown Fox jumps over the lazy Dog!"
        stopwords = ["the", "OVER", "a"]
        expected = ["quick", "brown", "fox", "jumps", "lazy", "dog"]
        self.assertEqual(tokenize_and_remove_stopwords(raw_text, stopwords), expected)

    def test_generate_ngrams_basic(self):
        tokens = ["machine", "learning", "model"]
        expected_bigrams = ["machine learning", "learning model"]
        self.assertEqual(generate_ngrams(tokens, 2), expected_bigrams)

    def test_mask_sensitive_pii_basic(self):
        text = "Contact Alice at alice@company.org or call 9876543210 for details."
        expected = "Contact Alice at ***@company.org or call 98XXXXXX10 for details."
        self.assertEqual(mask_sensitive_pii(text), expected)


if __name__ == "__main__":
    unittest.main()
