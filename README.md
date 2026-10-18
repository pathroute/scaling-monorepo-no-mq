# Software factory experiment arm

This repo is one arm of a merge-strategy experiment. Pull requests are opened and merged by automated agents; please don't open PRs by hand.

The code is a small product suite built with Bazel: Go services (`services/`), React apps (`web/`), Python data jobs (`data/`), and shared libs (`libs/`). Run `bazel test //...` with [Bazelisk](https://github.com/bazelbuild/bazelisk); every toolchain is pinned and downloaded by Bazel. See [AGENTS.md](AGENTS.md) for how to work in it.

`.factory/prs/` holds one metadata file per factory PR, and `<project>/changes/` holds the factory's change notes.
