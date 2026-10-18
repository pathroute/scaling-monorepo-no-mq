"""Settings read from the environment, shared by every data job."""

import os
from collections.abc import Mapping
from dataclasses import dataclass
from pathlib import Path


@dataclass(frozen=True)
class Settings:
    env: str
    data_dir: Path
    batch_size: int


def load(environ: Mapping[str, str] | None = None) -> Settings:
    """Reads DATA_ENV, DATA_DIR and BATCH_SIZE, with local defaults."""
    environ = os.environ if environ is None else environ
    env = environ.get("DATA_ENV", "local")
    if env not in {"local", "staging", "prod"}:
        raise ValueError(f"unknown DATA_ENV {env!r}")
    batch = int(environ.get("BATCH_SIZE", "10000"))
    if batch <= 0:
        raise ValueError("BATCH_SIZE must be positive")
    return Settings(env=env, data_dir=Path(environ.get("DATA_DIR", "data")), batch_size=batch)
