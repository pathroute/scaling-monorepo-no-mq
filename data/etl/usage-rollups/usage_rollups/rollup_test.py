from datetime import datetime

import polars as pl

from usage_rollups import rollup


def usage(rows):
    return pl.DataFrame(rows, schema=rollup.USAGE, orient="row")


def test_daily():
    rows = [("acme", "small", 10, 5, float(ms), datetime(2026, 1, 1, h)) for h, ms in enumerate(range(10, 210, 10))]
    rows.append(("acme", "large", 100, 100, 900.0, datetime(2026, 1, 1, 3)))
    out = rollup.daily(usage(rows))
    assert out.select("model", "requests", "tokens", "p95_latency_ms").rows() == [
        ("large", 1, 200, 900.0),
        ("small", 20, 300, 190.0),
    ]


def test_top_orgs():
    df = pl.DataFrame({"org": ["a", "b", "c", "b"], "tokens": [5, 3, 5, 4]})
    assert rollup.top_orgs(df, 2) == ["b", "a"]
