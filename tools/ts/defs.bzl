"""ts_package: the standard BUILD for a libs/ts or web package.

Compiles src/ (tests included) with tsc into dist/, runs the *.test.ts(x) files with Jest
(jsdom, native ESM), and for libs publishes :pkg, which other packages link through
"workspace:*" in their package.json.
"""

load("@aspect_rules_jest//jest:defs.bzl", "jest_test")
load("@aspect_rules_js//js:defs.bzl", "js_library")
load("@aspect_rules_js//npm:defs.bzl", "npm_package")
load("@aspect_rules_ts//ts:defs.bzl", "ts_project")
load("@npm//:vite/package_json.bzl", vite_bin = "bin")

# Available to every package's tests.
TEST_DEPS = [
    "//:node_modules/@jest/globals",
    "//:node_modules/@testing-library/dom",
    "//:node_modules/@testing-library/jest-dom",
    "//:node_modules/@testing-library/react",
    "//:node_modules/@testing-library/user-event",
    "//:node_modules/jest-environment-jsdom",
]

def ts_package(deps = [], publish = False):
    """Defines :ts, :test and (publish = True) :pkg for the current package.

    Args:
        deps: npm and workspace packages the sources import, e.g. "//:node_modules/react"
            or ":node_modules/@suite/ui-kit".
        publish: also define :pkg so other packages can depend on this one.
    """
    # tsc reads "type": "module" from it.
    js_library(
        name = "package_json",
        srcs = ["package.json"],
    )

    ts_project(
        name = "ts",
        srcs = native.glob(["src/**/*.ts", "src/**/*.tsx"], allow_empty = True),
        root_dir = "src",
        out_dir = "dist",
        declaration = True,
        tsconfig = "//:tsconfig",
        # TypeScript 7 (native) type-checks and emits in one fast pass.
        transpiler = "tsc",
        deps = [":package_json"] + deps + TEST_DEPS,
    )

    if native.glob(["src/**/*.test.ts", "src/**/*.test.tsx"], allow_empty = True):
        jest_test(
            name = "test",
            size = "small",
            config = "//tools/ts:jest_config",
            data = [
                ":ts",
                "package.json",
                "//tools/ts:jest_setup",
            ] + deps + TEST_DEPS,
            node_modules = "//:node_modules",
            node_options = ["--experimental-vm-modules"],
        )

    if publish:
        npm_package(
            name = "pkg",
            srcs = [
                ":ts",
                "package.json",
            ],
            exclude_srcs_patterns = ["**/*.test.*"],
            visibility = ["//visibility:public"],
        )

def web_app(deps = []):
    """A web/ app: ts_package plus :bundle, the Vite production build (bazel build :bundle).

    Args:
        deps: as for ts_package.
    """
    ts_package(deps = deps)

    up = "/".join([".."] * len(native.package_name().split("/")))
    vite_bin.vite(
        name = "bundle",
        srcs = native.glob(
            ["src/**/*.ts", "src/**/*.tsx"],
            exclude = ["src/**/*.test.ts", "src/**/*.test.tsx"],
            allow_empty = True,
        ) + [
            "index.html",
            "package.json",
            "//tools/ts:vite_config",
        ] + deps,
        args = [
            "build",
            "--config",
            up + "/tools/ts/vite.config.mjs",
        ],
        chdir = native.package_name(),
        out_dirs = ["bundle"],
        # Vite 8's bundler (Rolldown, native) resolves symlinks itself; let Node do the same
        # so both see one path for every input (see tools/ts/vite.config.mjs).
        patch_node_fs = False,
    )
