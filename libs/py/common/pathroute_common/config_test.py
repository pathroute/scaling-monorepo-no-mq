from pathlib import Path

import pytest

from pathroute_common import config


def test_defaults():
    s = config.load({})
    assert s == config.Settings(env="local", data_dir=Path("data"), batch_size=10000)


def test_overrides():
    s = config.load({"DATA_ENV": "prod", "DATA_DIR": "/d", "BATCH_SIZE": "5"})
    assert (s.env, s.data_dir, s.batch_size) == ("prod", Path("/d"), 5)


@pytest.mark.parametrize("env", [{"DATA_ENV": "dev"}, {"BATCH_SIZE": "0"}])
def test_rejects_bad_values(env):
    with pytest.raises(ValueError):
        config.load(env)
