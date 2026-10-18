"""Reading and writing tables. Writes are atomic so a crashed job never leaves half a file."""

import os
import tempfile
from pathlib import Path

import polars as pl


def read_ndjson(path: Path) -> pl.DataFrame:
    return pl.read_ndjson(path)


def write_parquet(df: pl.DataFrame, path: Path) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    fd, tmp = tempfile.mkstemp(dir=path.parent, suffix=".tmp")
    os.close(fd)
    try:
        df.write_parquet(tmp)
        os.replace(tmp, path)
    except BaseException:
        os.unlink(tmp)
        raise
