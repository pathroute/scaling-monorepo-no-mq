"""JSON logging, one object per line (same shape as the Go services' logs)."""

import json
import logging


class JSONFormatter(logging.Formatter):
    def format(self, record: logging.LogRecord) -> str:
        out = {"level": record.levelname, "msg": record.getMessage(), "job": record.name}
        out.update(getattr(record, "fields", {}))
        return json.dumps(out, default=str)


def get_logger(job: str) -> logging.Logger:
    """Returns a logger for job that writes JSON lines; safe to call repeatedly."""
    logger = logging.getLogger(job)
    if not logger.handlers:
        handler = logging.StreamHandler()
        handler.setFormatter(JSONFormatter())
        logger.addHandler(handler)
        logger.setLevel(logging.INFO)
    return logger
