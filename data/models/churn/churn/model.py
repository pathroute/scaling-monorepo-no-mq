"""Churn model: per-org usage features → logistic regression."""

import numpy as np
import polars as pl
from sklearn.linear_model import LogisticRegression
from sklearn.pipeline import Pipeline, make_pipeline
from sklearn.preprocessing import StandardScaler

from pathroute_common import schema

DAILY = {"org": pl.String, "day": pl.Date, "tokens": pl.Int64}
FEATURES = ["active_days", "mean_tokens", "trend"]


def features(daily: pl.DataFrame) -> pl.DataFrame:
    """Per org: days active, mean daily tokens, and second-half / first-half token ratio."""
    d = schema.require(daily, DAILY).sort("org", "day")
    half = pl.col("tokens").len() // 2
    return (
        d.group_by("org", maintain_order=True)
        .agg(
            active_days=pl.len(),
            mean_tokens=pl.col("tokens").mean(),
            trend=(pl.col("tokens").slice(half).sum() + 1) / (pl.col("tokens").head(half).sum() + 1),
        )
        .sort("org")
    )


def train(feats: pl.DataFrame, churned: list[bool], seed: int = 0) -> Pipeline:
    model = make_pipeline(StandardScaler(), LogisticRegression(random_state=seed))
    model.fit(feats.select(FEATURES).to_numpy(), np.asarray(churned))
    return model


def churn_risk(model: Pipeline, feats: pl.DataFrame) -> pl.DataFrame:
    """Adds a churn_risk probability column."""
    p = model.predict_proba(feats.select(FEATURES).to_numpy())[:, 1]
    return feats.with_columns(churn_risk=pl.Series(p))
