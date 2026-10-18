from datetime import date, timedelta

import numpy as np
import polars as pl

from churn import model


def synthetic(n_orgs: int, seed: int) -> tuple[pl.DataFrame, list[bool]]:
    """Churners use the product less and less; healthy orgs grow."""
    rng = np.random.default_rng(seed)
    rows, labels = [], []
    for i in range(n_orgs):
        churn = bool(i % 2)
        days = int(rng.integers(5, 15) if churn else rng.integers(20, 30))
        slope = -0.08 if churn else 0.03
        for d in range(days):
            tokens = max(0, int(1000 * (1 + slope * d) + rng.normal(0, 50)))
            rows.append((f"org{i:03d}", date(2026, 1, 1) + timedelta(days=d), tokens))
        labels.append(churn)
    return pl.DataFrame(rows, schema=model.DAILY, orient="row"), labels


def test_features():
    daily = pl.DataFrame(
        {"org": ["a"] * 4, "day": [date(2026, 1, d) for d in range(1, 5)], "tokens": [10, 10, 20, 20]}
    )
    f = model.features(daily)
    assert f.row(0) == ("a", 4, 15.0, 41 / 21)


def test_train_separates_churners():
    daily, labels = synthetic(60, seed=1)
    feats = model.features(daily)
    m = model.train(feats, labels)
    risk = model.churn_risk(m, feats)["churn_risk"].to_numpy()
    accuracy = ((risk > 0.5) == np.asarray(labels)).mean()
    assert accuracy >= 0.9
