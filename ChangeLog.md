# ChangeLog
Generated on: 2026-04-23T00:25:48+03:00

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
