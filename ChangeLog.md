# ChangeLog
Generated on: 2026-04-23T00:25:48+03:00

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

### 2026-05-02 - README Project Statement Update

- Added a new project statement to `README.md` in Turkey Turkish explaining that:
	- the project was initiated with Microsoft Copilot assistance,
	- it is continuously updated under human guidance,
	- it is offered as Free Software under the MIT License,
	- contributions are welcome and PlusClouds accepts contributions in a friendly manner.
- Added the equivalent U.S. English project statement to `README.en.md` with the same intent and contribution policy.

## Commit History
- bec52f4 2026-04-22 Mert Gör: ilk MIT/Expat ile
- 4cf8635 2026-04-22 Mert Gör: initial commit
- f59921e 2026-04-22 Mert Gör: feat: modular ISO download/upload CLI with hash verification and SSH known_hosts support
- 6273be1 2026-04-22 Mert Gör: feat(keygen): add SSH key pair generation module
- dc42c77 2026-04-22 Mert Gör: feat(downloader): enforce HTTPS for all HTTP download URLs
- 0bcbece 2026-04-23 Mert Gör: docs(readme): add U.S. English copy as README.en.md
- 543ed11 2026-04-23 Mert Gör: docs(readme): embed full MIT license text in both readme files
- 69eb5ef 2026-04-23 Mert Gör: docs(source): add MIT license header and file purpose comments to all Go source files
- bf7e17b 2026-04-23 Mert Gör: docs(changelog): add full commit history changelog
- 362c8a8 2026-04-23 Mert Gör: docs(changelog): refresh full history
- 358da80 2026-04-23 Mert Gör: docs(changelog): include latest changelog commit in history
- 48afdea 2026-04-23 Mert Gör: docs(readme): add changelog timestamp guidance in Turkish and U.S. English
- 2026-04-23 Mert Gör: feat(checksum): auto-discover checksum from ISO URL when available, calculate locally otherwise, verify on remote upload, and standardize verification output to ok/try again
