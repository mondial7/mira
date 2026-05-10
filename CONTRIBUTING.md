# Contributing to mira

Thanks for taking the time to contribute! This document is intentionally
short — most things follow standard Go conventions.

## Code of conduct

By participating, you agree to abide by the
[Code of Conduct](CODE_OF_CONDUCT.md).

## Development setup

Requirements: **Go 1.25+** and `git`. No other tooling is required.

```sh
git clone https://github.com/mondial7/mira
cd mira
make test
go run .
```

## Running the test suite

```sh
go test ./... -race -cover
```

The TUI itself is hard to unit-test, but the `internal/listing` package
has a thorough scaffolded test suite — keep it that way. New behaviour
in `internal/listing` should come with a test.

## Lint and format

We rely on the standard Go toolchain:

```sh
gofmt -s -w .
go vet ./...
```

If you have [`golangci-lint`](https://golangci-lint.run/) installed, run it
too. CI runs the same checks and will block merging if anything fails.

## Commit messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add status bar
fix: do not crash on unreadable directories
docs: clarify gitignore behaviour
refactor(listing): collapse helpers
```

GoReleaser uses these prefixes to build the changelog automatically.

## Pull requests

1. Fork the repo and create a topic branch (`feat/your-thing`).
2. Make your changes — keep PRs focused; one logical change per PR.
3. Add tests where it makes sense.
4. Make sure `go test ./...`, `go vet ./...`, and `gofmt` are clean.
5. Open a PR. Describe **what** changed and **why**.

Maintainers will review as soon as they can. Be patient — mira is a
side-project.

## Reporting bugs

Open an issue with:

- The version (`mira --version`).
- Your OS and terminal.
- Steps to reproduce.
- What you expected vs. what happened.

Security issues should follow [SECURITY.md](SECURITY.md) instead.

## Adding dependencies

We try to keep the dependency tree small. Before adding a new module, ask
yourself whether the standard library can do it. If you do add one, please
explain the tradeoff in the PR description.

## Releasing (maintainers only)

Tag a release and push:

```sh
git tag -a v1.2.3 -m "v1.2.3"
git push origin v1.2.3
```

GitHub Actions runs `goreleaser` to publish binaries, checksums, and the
Homebrew tap update.
