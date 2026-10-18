import pytest

from mail_classifier import classifier


def test_classifies_seed_like_subjects():
    m = classifier.train()
    got = classifier.classify(m, ["Invoice overdue reminder", "Error when I log in", "Pricing for enterprise plan"])
    assert got == ["billing", "support", "sales"]


def test_rejects_unknown_category():
    with pytest.raises(ValueError):
        classifier.train([("hi", "spam")])
