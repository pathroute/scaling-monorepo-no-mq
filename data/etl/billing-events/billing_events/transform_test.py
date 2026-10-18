from datetime import date, datetime

import polars as pl
import pytest

from billing_events import transform
from pathroute_common.schema import SchemaError


def events(rows):
    return pl.DataFrame(rows, schema=transform.EVENTS, orient="row")


def test_daily_revenue():
    df = events([
        ("acme", "charge", 1000, datetime(2026, 1, 1, 9)),
        ("acme", "charge", 500, datetime(2026, 1, 1, 17)),
        ("acme", "refund", 200, datetime(2026, 1, 1, 18)),
        ("acme", "charge", 300, datetime(2026, 1, 2, 9)),
        ("globex", "charge", 50, datetime(2026, 1, 1, 9)),
        ("globex", "chargeback", 50, datetime(2026, 1, 1, 9)),
        ("globex", "charge", 0, datetime(2026, 1, 1, 9)),
    ])
    out = transform.daily_revenue(df)
    assert out.rows() == [
        ("acme", date(2026, 1, 1), 1300, 2, 1),
        ("acme", date(2026, 1, 2), 300, 1, 0),
        ("globex", date(2026, 1, 1), 50, 1, 0),
    ]


def test_rejects_bad_schema():
    with pytest.raises(SchemaError):
        transform.daily_revenue(pl.DataFrame({"org": ["a"]}))
