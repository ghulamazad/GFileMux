# Contributing to GFileMux

This document is for anyone thinking about opening an issue, discussion, or pull request.

> [!NOTE]
>
> I appreciate you being here. GFileMux is a personal project that I maintain in my free time — it's not backed by a company or a team. If you're expecting me to dedicate my personal time to reviewing contributions and fixing bugs, I just ask that you spend a few minutes reading this first. Thank you. ❤️

---

## Table of Contents
- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Ways to Contribute](#ways-to-contribute)
  - [Found a Bug?](#found-a-bug)
  - [Have an Idea?](#have-an-idea)
  - [Sending a Pull Request](#sending-a-pull-request)
- [Development Setup](#development-setup)
- [Code Style](#code-style)
- [Writing Tests](#writing-tests)
- [Commit Messages](#commit-messages)
- [Versioning](#versioning)

---

## Code of Conduct

Keep it simple: be kind. Treat people the way you'd want to be treated. Hostile or dismissive behaviour isn't welcome here.

---

## Getting Started

```sh
# Fork on GitHub, then:
git clone https://github.com/<your-username>/GFileMux.git
cd GFileMux

# Keep in sync with the original
git remote add upstream https://github.com/ghulamazad/GFileMux.git

# Pull in dependencies
go mod download
```

---

## Ways to Contribute

### Found a Bug?

Before opening an issue, do a quick search to make sure it hasn't been reported already. When you do open one, include:

- Your **Go version** (`go version`)
- Your **GFileMux version** (from `go.mod` or the git tag)
- A **short code snippet** that reproduces the problem
- What you **expected** to happen vs what **actually** happened
- Any relevant error output or stack traces

The more context you give us, the faster we can help.

### Have an Idea?

Open a [GitHub Discussion](https://github.com/ghulamazad/GFileMux/discussions) or create an issue labeled `enhancement`. Tell us:

- What problem you're trying to solve
- How you'd like it to work
- Any alternatives you considered

No judgment on rough ideas — early conversation is better than a surprise PR.

### Sending a Pull Request

1. Create a branch off `main`:
   ```sh
   git checkout -b fix/some-bug
   # or
   git checkout -b feat/cool-feature
   ```
2. Make your changes.
3. Run the full test suite (with the race detector):
   ```sh
   go test ./... -race
   ```
4. Run `go vet`:
   ```sh
   go vet ./...
   ```
5. Commit and push:
   ```sh
   git push origin fix/some-bug
   ```
6. Open a pull request against `main` and describe what you did and why.

> **Heads-up:** For bigger changes, open an issue first to talk through the approach. It saves a lot of back-and-forth later.

---

## Development Setup

You'll need **Go ≥ 1.23**. That's really it.

```sh
# Run all tests with race detection
go test ./... -race -v

# Just the storage tests
go test ./storage/... -race -v

# Make sure everything compiles
go build ./...
```

If you want extended linting, [golangci-lint](https://golangci-lint.run/) is nice to have but not required.

---

## Code Style

Nothing exotic here — just standard Go conventions:

- Follow [Effective Go](https://go.dev/doc/effective_go). When in doubt, that's the guide.
- All exported symbols need doc comments. Even a single sentence is better than nothing.
- Use `errors.As` / `errors.Is` — don't compare error strings.
- Return the structured error types from `errors.go` (`ValidationError`, `StorageError`, etc.) instead of bare `fmt.Errorf` where it makes sense.
- Format with `gofmt` or `goimports` before committing. Most editors will do this automatically.

---

## Writing Tests

- New features and bug fixes need tests — no exceptions.
- Tests live next to the code (`foo_test.go` beside `foo.go`).
- Use `t.TempDir()` for temp files. It cleans itself up so you don't have to.
- Anything involving goroutines? Add a test that exercises it under `-race`.
- Avoid real network calls. Use mocks and interfaces — check out `MockStorage` in `handler_test.go` for an example.

**Naming:** `Test<Type>_<Scenario>` works well.  
Examples: `TestUpload_MaxFiles`, `TestDiskStorage_AutoCreateDir`


## Questions?

Open a [GitHub Discussion](https://github.com/ghulamazad/GFileMux/discussions) or just drop an issue. Happy to help!