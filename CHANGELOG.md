# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

---

## [0.2.0] — 2026-03-25

### Added
- **`errors.go`** — Structured, type-safe error types for precise caller-side handling:
  - `ValidationError` — file failed a validator (wrong MIME type, extension, or size)
  - `SizeError` — file or body exceeds the configured limit
  - `MaxFilesError` — too many files supplied for a single field
  - `StorageError` — wraps backend errors with `errors.Unwrap()` support
- **`ValidateFileExtension(exts ...string)`** — built-in validator; rejects files whose original extension is not in the allowed list (case-insensitive).
- **`ValidateMinFileSize(minBytes int64)`** — built-in validator; rejects files below a minimum byte threshold.
- **`WithMaxFiles(n int)`** — new option to cap the number of files per form field. Triggers `MaxFilesError` when exceeded.
- **`WithLogger(*slog.Logger)`** — attach a structured `log/slog` logger for lifecycle events (upload started / completed / failed).
- **`WithAllowedBuckets(buckets ...string)`** — whitelist permitted bucket names. Requests using an unlisted bucket are rejected before any I/O.
- **`WithChecksumValidation(bool)`** — compute a SHA-256 digest per file and expose it as `File.ChecksumSHA256`.
- **`UploadSingle(bucket, key string)`** — convenience middleware that enforces exactly one file per field.
- **`File.ChecksumSHA256 string`** — new field, populated when `WithChecksumValidation` is enabled.
- **`Files.All() []File`** — flat slice of all uploaded files across every form field.
- **`Files.Count() int`** — total uploaded file count across all fields.
- **`Storage.Delete(ctx, bucket, key)`** — all three backends now implement file deletion.
- **`MemoryStorage.Get(bucket, key)`** — retrieve raw bytes of a previously stored in-memory file.
- **`utils.ComputeSHA256`** — internal helper computing a hex-encoded SHA-256 of an `io.ReadSeeker`.
- **`CONTRIBUTING.md`** — comprehensive contributor guide covering setup, code style, testing conventions, commit message format, and versioning policy.

### Fixed
- **Race condition in `handler.go`** — concurrent goroutines previously wrote to a plain `map` without synchronisation. Replaced with `sync.Map`; each goroutine exclusively owns one key, so there is zero lock contention while being race-detector clean.
- **`MemoryStorage` data loss** — files were read into a buffer that was immediately discarded, making retrieval impossible. Files are now persisted in a `sync.RWMutex`-protected `map[string][]byte`.
- **`DiskStorage` bucket ignored** — the `bucket` parameter passed to `Upload()` was previously unused; all files landed in the same root directory. Bucket is now created as a subdirectory.
- **`DiskStorage.NewDiskStorage` fails on missing directory** — no longer returns an error when the target directory does not exist; it is created automatically via `os.MkdirAll`.
- **`ValidateMimeType` case comparison** — allowed types are now normalised once at construction time; `file.MimeType` is lowercased consistently at comparison time. Eliminated redundant double-lowercasing.
- **`DefaultUploadErrorHandlerFunc` unsafe JSON** — error string now formatted with `%q` to produce valid, escaped JSON.

### Changed
- **`Storage.Upload` signature** — unified to `*UploadFileOptions` (pointer) across all backends. Previously S3 accepted a value type, which was inconsistent.
- **All storage backend errors** are now wrapped in `*StorageError` for uniform `errors.As` usage.
- **All built-in validators** now return `*ValidationError` instead of plain `fmt.Errorf` strings.

### Breaking Changes

> [!WARNING]
> The following changes require updates to existing code.

- **`Storage` interface** now requires a `Delete(ctx context.Context, bucket, key string) error` method. Any custom `Storage` implementation must add this method.
- **`Storage.Upload`** parameter changed from `UploadFileOptions` (value) to `*UploadFileOptions` (pointer) in the S3 backend. If you were passing a struct literal directly, add `&`.

---

## [0.1.0] — Initial Release

- Core multipart file upload middleware compatible with any Go HTTP framework.
- Disk, Memory, and Amazon S3 storage backends.
- Pluggable file validation with `ValidateMimeType` and `ChainValidators`.
- Custom filename generation via `FileNameGeneratorFunc`.
- Concurrent multi-field upload processing using `errgroup`.
- Context helpers: `GetUploadedFilesFromContext`, `GetFilesByFieldFromContext`.
- Options: `WithStorage`, `WithMaxFileSize`, `WithFileValidatorFunc`, `WithFileNameGeneratorFunc`, `WithIgnoreNonExistentKey`, `WithUploadErrorHandlerFunc`.

[Unreleased]: https://github.com/ghulamazad/GFileMux/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/ghulamazad/GFileMux/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/ghulamazad/GFileMux/releases/tag/v0.1.0
