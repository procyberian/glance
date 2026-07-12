# ChangeLog
Generated on: 2026-07-13T00:00:00+03:00

## 2026-07-13 - v11.0.9 Dependency Maintenance Release

- Upgraded direct dependency `github.com/pkg/sftp` from `v1.13.10` to `v1.13.11`.
- Upgraded direct dependency `golang.org/x/crypto` from `v0.53.0` to `v0.54.0`.
- Upgraded indirect dependency `golang.org/x/sys` from `v0.46.0` to `v0.47.0`.
- Ran module refresh to update transitive dependency set and checksums (including `golang.org/x/term` and `golang.org/x/text` updates).
- Bumped `Version` constant to `v11.0.9` in `internal/cli/cli.go`.
- Updated `scripts/publish-release.sh` default tag to `v11.0.9`.
- Updated `README.md` and `README.tr.md`: advanced all release references to `v11.0.9` and refreshed the release summary section.

## 2026-07-03 - v11.0.8 Script Enhancement

- Enhanced `scripts/publish-release.sh`: token prompting is now interactive — when `GH_TOKEN` or `CODEBERG_TOKEN` are not set as environment variables and the script is running in a terminal (`-t 0`), the script prompts the user to type each token with `read -rsp` (input is hidden and not echoed). Non-interactive runs (CI, pipes) still require the tokens as environment variables and exit with an error if absent.
- Bumped `Version` constant to `v11.0.8` in `internal/cli/cli.go`.
- Updated `scripts/publish-release.sh` default tag to `v11.0.8`.
- Updated `README.md` and `README.tr.md`: advanced all version references to `v11.0.8`.

## 2026-07-03 - v11.0.7 Documentation Update (Minor)

- Updated `README.md` and `README.tr.md`: advanced all version references from `v11.0.5` to `v11.0.6`; rewrote "This release adds" section in both English and Turkish to describe v11.0.6 dependency and test coverage changes instead of the old v11.0.5 feature list.
- Corrected `ChangeLog.md` v11.0.6 entry: removed an inaccurate `golang.org/x/net` explicit-pin note (the module is transitive-only and `go mod tidy` correctly omits it from `go.mod`).
- Bumped `Version` constant to `v11.0.7` in `internal/cli/cli.go`.
- Updated `scripts/publish-release.sh` default tag to `v11.0.7`.
- This is a documentation-only minor update with no changes to binaries or source logic.

## 2026-07-03 - v11.0.6 Maintenance Release

- Updated `go` directive from `go 1.25.0` to `go 1.26` (matched to installed toolchain).
- Upgraded `github.com/jlaffaye/ftp` from `v0.2.0` to `v0.2.1`.
- Upgraded `golang.org/x/crypto` from `v0.50.0` to `v0.53.0` (security fixes).
- Upgraded `golang.org/x/sys` from `v0.43.0` to `v0.46.0`.
- Upgraded `github.com/hashicorp/errwrap` from `v1.0.0` to `v1.1.0`.
- Ran `go mod tidy` to remove unused indirect dependencies (`hashicorp/go-multierror`, `stretchr/testify`, `golang.org/x/term`, `golang.org/x/text`, `gopkg.in/yaml.v3`, `davecgh/go-spew`, `pmezard/go-difflib`).
- Added `internal/keygen/keygen_test.go`: tests for `GenerateKeyPair` covering ed25519, RSA (2048-bit), ECDSA, empty algorithm defaulting, custom key names, unsupported algorithm rejection, and private key file permission enforcement (0600).
- Added `internal/license/license_test.go`: tests asserting `Text` is non-empty and contains the MIT header, copyright holder, permission grant, and warranty disclaimer clauses.
- Added `internal/uploader/uploader_test.go`: tests for config validation — rejects empty host, empty user, and empty local file path; confirms default port (22) is applied before any connection attempt.
- Added `pkg/glance/glance_test.go`: tests for the public API surface — `Version` format, `LicenseText` content, `HelpText` output, `Parse` flag handling (`--version`, `--license`, `--help`), `Execute` short-circuit paths, `CalculateFileHash` (sha256 and sha512 output lengths), `VerifyFileHash` (correct and wrong hash), and `GenerateKeyPair` key file creation.
- Bumped `Version` constant to `v11.0.6` in `internal/cli/cli.go`.
- Updated `scripts/publish-release.sh` default tag to `v11.0.6`.
- Updated `README.md` and `README.tr.md`: all version references advanced to `v11.0.6` and release description updated.

## 2026-05-03 - v11.0.5 Feature Release

- Added live upload progress reporting to `internal/uploader`: percentage, instant speed, average speed, and ETA are now printed during every SSH/SFTP transfer, replacing the silent `io.Copy` call.
- Added upload resume support to `internal/uploader`: if a partial remote file exists and `--no-resume` is not set, the uploader seeks the local file to the remote offset and appends instead of overwriting. `uploader.Config` gains a `NoResume bool` field.
- Added `--version` flag and `version` subcommand: `./glance --version` and `./glance version` both print the current release string. A `Version` constant is now exported from both `internal/cli` and `pkg/glance`.
- Added `--json` flag: when present, `Run` writes a structured JSON object to stdout after all operations complete, containing `downloaded_path`, `checksum`, `checksum_algorithm`, `uploaded`, `remote_host`, and `remote_file` fields. This enables scripted and CI usage without screen-scraping.
- Added `--connect-timeout` flag (default 30 s): sets a hard TCP dial timeout for HTTP/HTTPS downloads, preventing indefinite hangs on unresponsive servers. `downloader.DownloadISOWithConnectTimeout` is exposed as a new library function; the original `DownloadISO` delegates to it with a zero timeout for backward compatibility.
- Exported `DownloadISOWithConnectTimeout` in `pkg/glance` so library consumers can pass their own connect timeout without forking internal packages.

## 2026-05-03 - v11.0.4 Documentation Update (Minor)

- Updated README.md to reflect v11.0.3 as the current published release version.
- Updated all code examples and release references in installation and publish instructions.
- This is a documentation-only minor update with no changes to binaries or source code.

## 2026-05-03 - v11.0.3 Release Preparation

- Added release notes at `release-notes/v11.0.3.md` for the next patch release.
- Built fresh multi-architecture prebuilt binaries and generated `dist/sha256sums.txt` for release asset integrity.
- Prepared signed release flow for `v11.0.3` publication to both PSD and Codeberg remotes.

## 2026-05-03 - Security Enforcement and CI Gate

- Added `.github/workflows/security.yml` to enforce `go mod verify`, `go build`, `go test`, `go vet`, and `govulncheck` in CI.
- Updated module dependencies and upgraded `golang.org/x/net` to `v0.53.0`.
- Added strict FTP policy: `ftp://` sources are now blocked by default and require explicit `--allow-insecure-ftp` opt-in.
- Added regression tests for checksum algorithm restrictions and FTP opt-in parsing under `internal/cli`, `internal/downloader`, and `internal/verifier`.

## 2026-05-03 - Strict Hash Policy

- Removed `md5` support from CLI checksum selection and now allow only `sha256` and `sha512`.
- Removed `--allow-weak-hash` and all weak-hash opt-in paths.
- Updated downloader, verifier, and uploader checksum logic to reject unsupported algorithms consistently.
- Updated README and help text to reflect strict hash policy and checksum discovery candidates.

## 2026-05-02 - Checksum Security Hardening

- Hardened automatic checksum parsing so multi-entry checksum files must match the selected ISO instead of accepting the first hash blindly.
- Added explicit request timeouts for HTTP checksum candidate retrieval to reduce hanging network operations.
- Introduced `--allow-weak-hash` and disabled `--checksum-algo md5` by default unless this opt-in flag is provided.
- Updated README and CLI help references to document weak-hash opt-in behavior.

## 2026-05-02 - v11.0.2 Release Preparation

- Added `release-notes/v11.0.2.md` for the new patch tag that carries the public `pkg/glance` library surface.
- Advanced published install, checkout, and release publishing examples from `v11.0.1` to `v11.0.2` to avoid reusing the existing tag.
- Updated the default `scripts/publish-release.sh` tag target to `v11.0.2`.

## 2026-05-02 - Public Library Packaging

- Added a reusable public package at `pkg/glance` so tagged releases can be consumed as a Go library instead of only as a CLI binary.
- Exposed library-safe wrappers for download, listing, checksum resolution, verification, upload, key generation, and CLI execution.
- Updated `main.go` to use the exported package surface, keeping the CLI path aligned with the new public API.
- Documented library import and usage examples in both `README.md` and `README.tr.md`.

## 2026-05-02 - v11.0.1 Release Preparation

- Added `release-notes/v11.0.1.md` to capture the patch release scope for post-`v11.0.0` release automation and documentation updates.
- Prepared a patch release path so Go module consumers can receive the latest repository state via `github.com/procyberian/glance/v11@v11.0.1` instead of reusing the immutable `v11.0.0` tag.
- Confirmed that multi-architecture binaries are intended to be attached to the release entry associated with the `v11.0.1` git tag.
- Promoted the new `pkg/glance` public API into the `v11.0.1` patch release scope and aligned README install/publish examples with that tag.

## 2026-05-02 - v11.0.0 Release Publication

- Updated the Go module path from `github.com/procyberian/glance/v10` to `github.com/procyberian/glance/v11` so the new major release remains valid for Go module consumers.
- Updated internal imports in `main.go` and `internal/cli/cli.go` to the `v11` module path.
- Updated `README.md` and `README.tr.md` to explicitly state that the project has been published and documented how to install it with `go install github.com/procyberian/glance/v11@latest` or build it from the `v11.0.0` tag.
- Added a dedicated binary download section to both README files and prepared reusable release notes in `release-notes/v11.0.0.md`.
- Added `scripts/publish-release.sh` to create or update GitHub and Codeberg releases and upload packaged assets from `dist/`.
- Extended `scripts/publish-release.sh` with `--github-only`, `--codeberg-only`, and `--dry-run`, and documented token scope guidance for both hosting providers.
- Extended `scripts/publish-release.sh` with `--dist-dir` and `--notes-file` so release inputs can be overridden from the command line.
- Prepared the repository for signed commit, signed `v11.0.0` tag creation, and push to both `PSD` and `codeberg` remotes.

## 2026-05-02 - Module Path and Publishing Fix (v10)

- Updated `go.mod` module path from `glance` to `github.com/procyberian/glance/v10` to comply with Go semantic versioning rules (v2+ requires major version suffix in module path).
- Updated all internal import paths across `main.go` and `internal/cli/cli.go` to use `github.com/procyberian/glance/v10/internal/...`.
- Project can now be installed directly via `go install github.com/procyberian/glance/v10@latest`.
- To publish a release: tag with `git tag v10.x.x` and push the tag; Go module proxy will index it automatically.
- Note: packages under `internal/` remain intentionally unexportable; move to `pkg/` if a reusable library surface is desired in the future.

## 2026-05-02 - README Language Default Update

- `README.en.md` has been replaced by `README.md` as the default project README in U.S. English.
- `README.tr.md` is now provided as the Turkish language option.

## 2026-05-02 - README Project Statement Update

- Added a new project statement to `README.md` in Turkey Turkish explaining that:
	- the project was initiated with Microsoft Copilot assistance,
	- it is continuously updated under human guidance,
	- it is offered as Free Software under the MIT License,
	- contributions are welcome and PlusClouds accepts contributions in a friendly manner.
- Added the equivalent U.S. English project statement to `README.en.md` with the same intent and contribution policy.

## 2026-05-02 - Session Activity Log

### User Requests and Technical Actions (Chronological)

- User requested a logical consistency review across CLI arguments and general code flow.
- Main entry and CLI orchestration were inspected to validate argument parsing and command routing.
- Downloader, uploader, verifier, and keygen modules were reviewed for cross-module consistency.
- A core contradiction was identified: FTP directory listing existed, but direct FTP file download path fell back to local file copy behavior.
- This contradiction was verified by running `go run . --download --iso ftp://example.com/test.iso --output /tmp/glance-test` and observing local-path style open failure.
- User requested fixing item 1 and item 2 with interactive behavior and selectable lists.

### Implemented Code Changes During Session

- Added direct FTP file download support into downloader flow.
- Added resume-aware FTP download attempt (`RetrFrom`) with fallback to full restart when server resume is unavailable.
- Preserved `.download` temporary file finalization pattern for FTP direct file downloads.
- Updated upload-only behavior when `--file` is missing and source is HTTP/FTP URL:
	- local candidate files are listed interactively,
	- user can select by number or provide full path manually.
- Added helper functions for local upload candidate enumeration and selection prompt.

### Runtime Validation and Mirror Tests

- Build checks were repeatedly run successfully (`go build ./...`, `go build -o glance .`).
- Multiple mirror endpoints were tested to isolate code-vs-mirror behavior.
- `ftp.lip6.fr` path used earlier was validated as missing (550/404 path mismatch).
- `https://ftp.uni-stuttgart.de/debian-cd/` and child paths were validated reachable over HTTPS.
- Directory scan on some mirrors was observed to be long-running; bounded runs timed out as expected.
- Direct ISO download on the Stuttgart mirror was validated end-to-end (download + checksum auto-discovery + checksum verification).

### CLI Robustness Improvements

- Added `--scan-timeout` flag for HTTP/FTP directory listing timeout control.
- Default behavior documented as 60 seconds; `0` disables timeout.
- Added parser validation for malformed invocations:
	- unexpected positional args now return explicit error,
	- invalid `--iso` values that look like flags are rejected.
- This prevented malformed command order from silently entering the wrong flow.

### Timeout Architecture Improvement

- Identified risk: goroutine-based timeout wrappers could return while background scans continued.
- Refactored timeout handling into downloader scan functions:
	- added timeout-aware scan variants,
	- inserted timeout checks within FTP traversal/checksum loops and HTTP traversal loops,
	- moved connect/request timeout selection to elapsed-time-aware helper logic.
- CLI wrappers now call downloader timeout-aware APIs directly (no local goroutine timeout wrappers).

### Documentation Updates

- Updated `README.md` and `README.en.md` with:
	- direct FTP ISO download behavior,
	- interactive upload file selection without `--file`,
	- `--scan-timeout` usage and examples,
	- troubleshooting notes for incorrect argument order.

### Session Outcome Summary

- FTP direct file download contradiction resolved.
- Upload-only URL case now interactive and user-selectable.
- Scan timeout made configurable and parser hardened against malformed ordering.
- Timeout logic moved closer to scan execution paths to reduce background-work risk after timeout.

## 2026-04-23 - v7.0.0 Release Preparation

### Documentation Changes

- Refined the Turkish README with full Turkish characters for headings, feature descriptions, and usage guidance
- Standardized Turkey Turkish terminology across user, administrator, and developer sections for better release-readiness
- Aligned README section wording around resume handling, destination prompts, progress reporting, and changelog guidance

## 2026-04-23 - v6.0.0 Release Preparation

### User-Facing Changes

- Added FTP directory scanning with recursive ISO discovery and checksum-aware selection
- Added HTTP/HTTPS public directory scanning across parent and child directories on the same host
- Added interactive single, multiple, and `all` ISO selection
- Added progress display for total ISO detection and checksum resolution
- Added automatic prompt asking where the selected ISO files should be downloaded when no output destination flag is provided
- Added `--output-path` for downloading a single ISO to an exact target file path
- Added resumable HTTP downloads using `.download` partial files
- Added `--no-resume` to ignore old partial files and restart from zero

### Operational and Admin Changes

- Added rate limiting for FTP traversal and checksum probing to reduce aggressive mirror traffic patterns
- Improved checksum verification output for both local verification and remote upload validation
- Preserved upload flow with SSH/SFTP and remote checksum comparison

### Developer and Maintenance Changes

- Extended CLI orchestration for directory scanning, selection parsing, prompt handling, resume hints, and destination prompts
- Expanded downloader module with FTP listing, recursive HTTP discovery, progress helpers, resume logic, and explicit output path support
- Refreshed Turkish and U.S. English documentation for end users, system administrators, and developers
- Prepared changelog content for the `v6.0.0` release process

## Commit History
- b48bfe0 2026-05-02 Mert Gör: docs(readme): add AI-assisted origin and contribution policy statement
- 3dd94dd 2026-05-02 Mert Gör: feat(cli): add FTP direct download, scan timeout, and robust argument validation
- 0b173d4 2026-04-23 Mert Gör: docs: refine Turkish README and changelog for v7.0.0
- 6b6ac00 2026-04-23 Mert Gör: release(v6.0.0): ship recursive ISO discovery and resumable downloads
- 3804f46 2026-04-23 Mert Gör: feat(checksum): automate checksum flow and remote verification
- 08874f7 2026-04-23 Mert Gör: docs(changelog): include latest README guidance commit
- 48afdea 2026-04-23 Mert Gör: docs(readme): add changelog timestamp guidance in Turkish and U.S. English
- 358da80 2026-04-23 Mert Gör: docs(changelog): include latest changelog commit in history
- 362c8a8 2026-04-23 Mert Gör: docs(changelog): refresh full history
- bf7e17b 2026-04-23 Mert Gör: docs(changelog): add full commit history changelog
- 69eb5ef 2026-04-23 Mert Gör: docs(source): add MIT license header and file purpose comments to all Go source files
- 543ed11 2026-04-23 Mert Gör: docs(readme): embed full MIT license text in both readme files
- 0bcbece 2026-04-23 Mert Gör: docs(readme): add U.S. English copy as README.en.md
- dc42c77 2026-04-22 Mert Gör: feat(downloader): enforce HTTPS for all HTTP download URLs
- 6273be1 2026-04-22 Mert Gör: feat(keygen): add SSH key pair generation module
- f59921e 2026-04-22 Mert Gör: feat: modular ISO download/upload CLI with hash verification and SSH known_hosts support
- 4cf8635 2026-04-22 Mert Gör: initial commit
