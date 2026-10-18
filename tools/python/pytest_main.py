"""Entry point for py_pytest_test: runs pytest on the test files passed as arguments."""

import sys

import pytest

if __name__ == "__main__":
    # No cache dir: the test runs in a read-only sandbox.
    sys.exit(pytest.main(["-p", "no:cacheprovider", "-q", *sys.argv[1:]]))
