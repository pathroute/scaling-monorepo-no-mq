"""py_pytest_test: a py_test that runs its srcs with pytest. Gazelle maps py_test to it."""

load("@rules_python//python:py_test.bzl", "py_test")

def py_pytest_test(name, srcs, deps = [], args = [], **kwargs):
    kwargs.pop("main", None)
    py_test(
        name = name,
        srcs = srcs + ["//tools/python:pytest_main.py"],
        main = "//tools/python:pytest_main.py",
        args = args + ["$(rootpath %s)" % s for s in srcs],
        deps = deps if "@pypi//pytest" in deps else deps + ["@pypi//pytest"],
        **kwargs
    )
