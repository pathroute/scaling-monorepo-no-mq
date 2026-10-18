"""Aggregates ai-gateway request logs into daily per-org, per-model usage."""

import polars as pl

from pathroute_common import schema

USAGE = {
    "org": pl.String,
    "model": pl.String,
    "input_tokens": pl.Int64,
    "output_tokens": pl.Int64,
    "latency_ms": pl.Float64,
    "ts": pl.Datetime("us"),
}


def daily(usage: pl.DataFrame) -> pl.DataFrame:
    """One row per org, model and day: requests, tokens and p95 latency."""
    return (
        schema.require(usage, USAGE)
        .with_columns(day=pl.col("ts").dt.date())
        .group_by("org", "model", "day")
        .agg(
            requests=pl.len(),
            tokens=(pl.col("input_tokens") + pl.col("output_tokens")).sum(),
            p95_latency_ms=pl.col("latency_ms").quantile(0.95, interpolation="nearest"),
        )
        .sort("org", "model", "day")
    )


def top_orgs(rollup: pl.DataFrame, n: int) -> list[str]:
    """The n orgs with the most tokens across all models and days."""
    totals = rollup.group_by("org").agg(pl.col("tokens").sum()).sort(["tokens", "org"], descending=[True, False])
    return totals["org"].head(n).to_list()
