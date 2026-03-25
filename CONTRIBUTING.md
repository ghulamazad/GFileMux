# Contributing to GFileMux

First off, **thank you for taking the time to contribute!** 🎉  
Every contribution — bug reports, feature ideas, documentation improvements, or code — is appreciated and helps make GFileMux better for everyone.

---

## Table of Contents
- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [How to Contribute](#how-to-contribute)
  - [Reporting Bugs](#reporting-bugs)
  - [Suggesting Features](#suggesting-features)
  - [Submitting a Pull Request](#submitting-a-pull-request)
- [Development Setup](#development-setup)
- [Code Style](#code-style)
- [Testing](#testing)
- [Commit Messages](#commit-messages)
- [Versioning](#versioning)

---

## Code of Conduct

This project follows a simple rule: **be kind and respectful**. Harassment, discrimination, or hostile behaviour of any kind will not be tolerated. Please treat every contributor the way you would want to be treated.

---

## Getting Started

1. **Fork** the repository on GitHub.
2. **Clone** your fork locally:
   ```sh
   git clone https://github.com/<your-username>/GFileMux.git
   cd GFileMux
   ```
3. **Add the upstream remote** so you can pull future changes:
   ```sh
   git remote add upstream https://github.com/ghulamazad/GFileMux.git
   ```
4. **Install dependencies:**
   ```sh
   go mod download
   ```

---

## How to Contribute

### Reporting Bugs

Before opening an issue, please:
- Search [existing issues](https://github.com/ghulamazad/GFileMux/issues) to avoid duplicates.
- Check that you are on the latest version.

When filing a bug report, include:
- **Go version** (`go version`)
- **GFileMux version** (the `go.mod` entry or git tag)
- **Minimal reproducing example** — a short code snippet or test is ideal
- **Expected vs actual behaviour**
- **Stack trace / error message** (if applicable)

### Suggesting Features

Open a [GitHub Discussion](https://github.com/ghulamazad/GFileMux/discussions) or an issue tagged `enhancement`. Describe:
- The problem you are trying to solve
- Your proposed solution (even a rough sketch helps)
- Any alternatives you considered

### Submitting a Pull Request

1. **Create a branch** from `main`:
   ```sh
   git checkout -b feat/my-new-feature
   ```
2. **Make your changes** (see [Code Style](#code-style) and [Testing](#testing) below).
3. **Run the full test suite** including the race detector:
   ```sh
   go test ./... -race
   ```
4. **Run `go vet`:**
   ```sh
   go vet ./...
   ```
5. **Commit** your changes (see [Commit Messages](#commit-messages)).
6. **Push** your branch:
   ```sh
   git push origin feat/my-new-feature
   ```
7. **Open a Pull Request** against `main`. Fill in the PR template — describe *what* changed and *why*.

> **Tip:** For large changes, open an issue first to discuss the design before writing code. This saves everyone time.

---

## Development Setup

| Tool | Purpose |
|------|---------|
| Go ≥ 1.23 | Required |
| `go vet` | Static analysis (built-in) |
| `golangci-lint` (optional) | Extended linting — `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest` |

Run all tests with the race detector:
```sh
go test ./... -race -v
```

Run only storage tests:
```sh
go test ./storage/... -race -v
```

Build everything:
```sh
go build ./...
```

---

## Code Style

- Follow [Effective Go](https://go.dev/doc/effective_go) and the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).
- All exported types, functions, and methods **must have doc comments**.
- Use `errors.As` / `errors.Is` — do not compare error strings.
- Return structured errors from the `errors.go` types (`ValidationError`, `StorageError`, etc.) instead of plain `fmt.Errorf` strings where applicable.
- Keep functions small and focused. If a function needs a long comment explaining what it does, consider splitting it.
- Format your code with `gofmt` (or `goimports`) before committing.

---

## Testing

- Every new feature or bug fix **must** include a test.
- Tests live alongside the code they test (`foo_test.go` next to `foo.go`).
- Use `t.TempDir()` for temporary files — it is automatically cleaned up.
- For concurrency-sensitive code, add a test that runs under `-race`.
- Avoid network calls in tests; use interfaces and mocks (see `MockStorage` in `handler_test.go`).

**Test naming convention:**
```
Test<TypeOrFunction>_<Scenario>
```
Examples: `TestUpload_MaxFiles`, `TestDiskStorage_AutoCreateDir`

---

## Commit Messages

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>(<scope>): <short summary>

[optional body]

[optional footer]
```

**Types:**

| Type | When to use |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation only |
| `test` | Adding or fixing tests |
| `refactor` | Code change (no feature, no fix) |
| `chore` | Build process, dependency updates |
| `perf` | Performance improvement |

**Examples:**
```
feat(storage): add Delete method to all backends
fix(handler): replace shared map with sync.Map to eliminate race condition
docs(readme): document WithMaxFiles and WithLogger options
test(storage): add concurrent upload stress test for MemoryStorage
```

---

## Versioning

GFileMux follows [Semantic Versioning](https://semver.org/):

| Change type | Version bump |
|-------------|-------------|
| Breaking API change (e.g. new required interface method) | **MAJOR** (`v1.0.0 → v2.0.0`) |
| Backward-compatible new feature | **MINOR** (`v0.1.0 → v0.2.0`) |
| Backward-compatible bug fix | **PATCH** (`v0.1.0 → v0.1.1`) |

All changes are recorded in [CHANGELOG.md](CHANGELOG.md) under the `[Unreleased]` section. When a release is cut, that section is renamed to the version + date.

---

## Questions?

Open a [GitHub Discussion](https://github.com/ghulamazad/GFileMux/discussions) or reach out via an issue. We're happy to help!