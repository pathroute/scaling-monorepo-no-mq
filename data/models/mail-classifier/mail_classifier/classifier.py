"""Classifies email subjects into mail-tracker's categories with TF-IDF + logistic regression."""

from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.linear_model import LogisticRegression
from sklearn.pipeline import Pipeline, make_pipeline

# Must match services/mail-tracker/api.Categories.
CATEGORIES = ("sales", "support", "billing", "other")

# A tiny seed set so the model works before real labels exist.
SEED = [
    ("Quick question about pricing for your team plan", "sales"),
    ("Can we schedule a demo next week", "sales"),
    ("Proposal and quote for 50 seats", "sales"),
    ("Login page shows an error after update", "support"),
    ("App crashes when I upload a file", "support"),
    ("How do I reset my password", "support"),
    ("Invoice 1042 is overdue", "billing"),
    ("Your payment failed, update your card", "billing"),
    ("Refund for duplicate charge", "billing"),
    ("Lunch on Friday?", "other"),
    ("Newsletter: what we shipped in March", "other"),
    ("Out of office until Monday", "other"),
]


def train(examples: list[tuple[str, str]] = SEED, seed: int = 0) -> Pipeline:
    for _, label in examples:
        if label not in CATEGORIES:
            raise ValueError(f"unknown category {label!r}")
    texts, labels = zip(*examples)
    model = make_pipeline(TfidfVectorizer(ngram_range=(1, 2), sublinear_tf=True), LogisticRegression(C=10, random_state=seed))
    model.fit(texts, labels)
    return model


def classify(model: Pipeline, subjects: list[str]) -> list[str]:
    return list(model.predict(subjects))
