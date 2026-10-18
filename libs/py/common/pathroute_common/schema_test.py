import polars as pl
import pytest

from pathroute_common import schema

SCHEMA = {"org": pl.String, "amount": pl.Int64}


def test_require_narrows_columns():
    df = pl.DataFrame({"org": ["a"], "amount": [1], "extra": [True]})
    assert schema.require(df, SCHEMA).columns == ["org", "amount"]


def test_problems():
    df = pl.DataFrame({"org": [1]})
    assert schema.problems(df, SCHEMA) == ["column org is Int64, want String", "missing column amount"]
    with pytest.raises(schema.SchemaError):
        schema.require(df, SCHEMA)
