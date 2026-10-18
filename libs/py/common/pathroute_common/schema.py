"""Checks that a DataFrame has the columns and types a job expects."""

import polars as pl

Schema = dict[str, pl.DataType | type[pl.DataType]]


class SchemaError(ValueError):
    pass


def problems(df: pl.DataFrame, schema: Schema) -> list[str]:
    """Lists missing columns and type mismatches; extra columns are allowed."""
    out = []
    for col, want in schema.items():
        if col not in df.columns:
            out.append(f"missing column {col}")
        elif df.schema[col] != want:
            out.append(f"column {col} is {df.schema[col]}, want {want}")
    return out


def require(df: pl.DataFrame, schema: Schema) -> pl.DataFrame:
    """Returns df narrowed to the schema's columns, or raises SchemaError."""
    if errs := problems(df, schema):
        raise SchemaError("; ".join(errs))
    return df.select(list(schema))
