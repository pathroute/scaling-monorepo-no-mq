"""Turns raw billing events into the daily_revenue warehouse table."""

import polars as pl

from pathroute_common import schema

EVENTS = {"org": pl.String, "kind": pl.String, "amount_cents": pl.Int64, "ts": pl.Datetime("us")}
KINDS = ("charge", "refund")


def daily_revenue(events: pl.DataFrame) -> pl.DataFrame:
    """Sums charges minus refunds per org and day; unknown kinds and non-positive amounts are dropped."""
    ev = schema.require(events, EVENTS).filter(pl.col("kind").is_in(KINDS) & (pl.col("amount_cents") > 0))
    signed = pl.when(pl.col("kind") == "refund").then(-pl.col("amount_cents")).otherwise(pl.col("amount_cents"))
    return (
        ev.with_columns(day=pl.col("ts").dt.date(), signed=signed)
        .group_by("org", "day")
        .agg(
            revenue_cents=pl.col("signed").sum(),
            charges=(pl.col("kind") == "charge").sum(),
            refunds=(pl.col("kind") == "refund").sum(),
        )
        .sort("org", "day")
    )
