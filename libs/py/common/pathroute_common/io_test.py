import json
import logging

import polars as pl

from pathroute_common import io, log


def test_parquet_round_trip(tmp_path):
    df = pl.DataFrame({"a": [1, 2]})
    out = tmp_path / "nested" / "t.parquet"
    io.write_parquet(df, out)
    assert pl.read_parquet(out).equals(df)
    assert [p.name for p in out.parent.iterdir()] == ["t.parquet"]


def test_read_ndjson(tmp_path):
    p = tmp_path / "e.ndjson"
    p.write_text('{"a": 1}\n{"a": 2}\n')
    assert io.read_ndjson(p)["a"].to_list() == [1, 2]


def test_json_log_line():
    rec = logging.LogRecord("job", logging.INFO, __file__, 1, "done %d", (3,), None)
    rec.fields = {"rows": 3}
    assert json.loads(log.JSONFormatter().format(rec)) == {"level": "INFO", "msg": "done 3", "job": "job", "rows": 3}
    assert log.get_logger("x") is log.get_logger("x")
