"""Hashed TF vectors for documents, and cosine-similarity queries over them."""

from dataclasses import dataclass

import numpy as np
import polars as pl
from sklearn.feature_extraction.text import HashingVectorizer

from pathroute_common import schema

DOCS = {"id": pl.String, "kind": pl.String, "text": pl.String}
DIMENSIONS = 2**12


@dataclass
class Index:
    ids: list[str]
    vectors: np.ndarray  # rows are L2-normalised


def _vectorizer() -> HashingVectorizer:
    return HashingVectorizer(n_features=DIMENSIONS, alternate_sign=False, norm="l2")


def build(docs: pl.DataFrame) -> Index:
    """Indexes every non-empty document."""
    d = schema.require(docs, DOCS).filter(pl.col("text").str.strip_chars() != "")
    return Index(ids=d["id"].to_list(), vectors=_vectorizer().transform(d["text"].to_list()).toarray())


def query(index: Index, text: str, k: int = 5) -> list[tuple[str, float]]:
    """The k most similar documents with a positive score, best first."""
    if not index.ids:
        return []
    q = _vectorizer().transform([text]).toarray()[0]
    scores = index.vectors @ q
    best = np.argsort(-scores, kind="stable")[:k]
    return [(index.ids[i], float(scores[i])) for i in best if scores[i] > 0]
