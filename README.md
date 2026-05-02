# glance-cli

`glance` is a modular Go CLI for ISO discovery, download verification, resume-aware transfers, and SSH/SFTP delivery.

This project was initiated with the help of AI-powered Microsoft Copilot, but it is continuously updated under human guidance. The project is offered as Free Software under the MIT License and is open to contributions. PlusClouds warmly and amicably welcomes community contributions.

## Published Release

`glance` is now published as the `v11` Go module line and as the `v11.0.1` git release tag.

You can install the latest `v11` release directly with:

```bash
go install github.com/procyberian/glance/v11@latest
```

You can also download the source and build it locally:

```bash
git clone git@github.com:procyberian/glance.git
cd glance
git checkout v11.0.1
go build -o glance .
```

## Binary Downloads

Binary archives for `v11.0.1` are distributed from the project release pages:

- GitHub Releases: <https://github.com/procyberian/glance/releases>
- Codeberg Releases: <https://codeberg.org/procyberian/glance/releases>

Look for the `v11.0.1` release and download the asset that matches your platform.

Planned asset names for this release:

- `glance-linux-amd64.tar.gz`
- `glance-linux-arm64.tar.gz`
- `glance-darwin-amd64.tar.gz`
- `glance-darwin-arm64.tar.gz`
- `glance-windows-amd64.zip`

To publish the release entry and upload these assets automatically after setting API tokens:

```bash
GH_TOKEN=... CODEBERG_TOKEN=... ./scripts/publish-release.sh v11.0.1
```

Token guidance:

- GitHub classic personal access token: `repo`
- GitHub fine-grained token: repository `Contents` set to read/write
- Codeberg token: repository write access for releases and release assets

Useful script modes:

```bash
./scripts/publish-release.sh --dry-run v11.0.1
GH_TOKEN=... ./scripts/publish-release.sh --github-only v11.0.1
CODEBERG_TOKEN=... ./scripts/publish-release.sh --codeberg-only v11.0.1
GH_TOKEN=... CODEBERG_TOKEN=... ./scripts/publish-release.sh --dist-dir dist --notes-file release-notes/v11.0.1.md v11.0.1
```

This release adds:

- recursive FTP ISO discovery
- recursive HTTP/HTTPS directory discovery across parent and child paths on the same host
- interactive ISO selection with single, multiple, or `all` choices
- automatic checksum discovery for HTTP and FTP ISO sources
- `.download` resume files for interrupted downloads
- `--no-resume` to force a clean restart
- `--output-path` for an exact destination file path
- `--scan-timeout` to cap HTTP/FTP directory scan duration
- a post-selection prompt asking where the ISO should be downloaded when no destination flag is supplied
- interactive local file selection for `--upload` when `--file` is not provided

## Audience

### End Users

- discover available ISO files without manually copying deep mirror URLs
- verify downloads automatically
- resume interrupted downloads instead of starting over

### System Administrators

- standardize ISO acquisition from mirrors and package repositories
- upload verified files to remote systems over SSH/SFTP
- use host verification and remote checksum validation in controlled environments

### Developers

- work with a layered Go codebase split into CLI, downloader, verifier, uploader, and keygen modules
- extend either FTP or HTTP scanning behavior independently
- prepare releases with aligned README and changelog documentation

## Features

- `--download`: download an ISO or copy a local ISO file
- `--iso`, `--url`: accept HTTP, HTTPS, FTP, or local sources
- `--checksum`: verify against a provided checksum or auto-discovered checksum when available
- `--checksum-algo`: `sha256`, `sha512`, `md5`
- live transfer progress with percentage, instant speed, average speed, network speed, and ETA
- FTP scanning with total ISO detection first, then checksum resolution progress in a single progress bar
- HTTP/HTTPS scanning that traverses public directory listings recursively within the same host
- interactive selection with `1`, `1,3,5`, or `all`
- resume support via `.download` files
- `--no-resume` to ignore any partial file and restart from byte zero
- `--output` for a target directory
- `--output-path` for an explicit full file path when downloading a single ISO
- `--scan-timeout` for HTTP/FTP directory scan timeout (seconds)
- interactive destination prompt after ISO selection when no output flag is supplied
- interactive upload mode that lists local ISO/image files if `--file` is omitted
- `--upload` for SSH/SFTP delivery
- remote checksum verification after upload
- `--keygen` for `ed25519`, `rsa`, and `ecdsa`
- `--license` to print the MIT license text

## Build

```bash
go mod tidy
go build -o glance .
```

## Library Usage

The tagged project now exposes a reusable public package at `github.com/procyberian/glance/v11/pkg/glance`.

Example:

```go
package main

import (
    "log"

    glance "github.com/procyberian/glance/v11/pkg/glance"
)

func main() {
    result, err := glance.DownloadAndVerify(glance.DownloadOptions{
        Source:            "https://releases.ubuntu.com/24.04/ubuntu-24.04.4-live-server-amd64.iso",
        OutputDir:         "./downloads",
        ChecksumAlgorithm: "sha256",
        AllowResume:       true,
    })
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("downloaded=%s checksum=%s algo=%s", result.Path, result.Checksum, result.Algorithm)
}
```

The same package also exposes direct wrappers for `DownloadISO`, `ListFTPISOs`, `ListHTTPISOs`, `ResolveChecksum`, `VerifyFileHash`, `CalculateFileHash`, `UploadFile`, `GenerateKeyPair`, `Parse`, `Run`, and `Execute`.

## End User Guide

### 1. Download a direct ISO URL

```bash
./glance --download --iso https://releases.ubuntu.com/24.04/ubuntu-24.04.4-live-server-amd64.iso
```

### 2. Scan an FTP directory, list ISOs, choose one, and download it

```bash
./glance --download --iso ftp://ftp.example.com/iso/
```

Expected flow:

```text
Detecting total ISO count...
Found 100 total ISOs
[==========----------] 50.00% | 50/100 ISOs
FTP ISO list:
    1) example-1.iso | size: 2.80 GB | checksum: ...
    2) example-2.iso | size: 1.90 GB | checksum: ...
Select ISO number(s) (example: 1 or 1,3,5 or all) [1-100]:
Which directory should the ISO be downloaded to? [default: ./downloads]:
```

### 3. Scan an HTTP/HTTPS directory recursively

```bash
./glance --download --iso https://ftp.example.com/
```

In this mode the tool:

- reads public directory listings
- traverses relevant child directories on the same host
- includes parent directories for broader ISO discovery
- keeps only real `.iso` files in the final selection list

### 3.1 Download a direct FTP ISO file

```bash
./glance --download --iso ftp://ftp.example.com/iso/example.iso
```

In this mode the tool downloads the file directly from FTP. If a partial `.download` file exists it attempts resume first; if resume is not supported by the server, it safely restarts from zero.

### 4. Selection formats

```text
1
1,3,5
all
```

### 4.1 Set directory scan timeout

```bash
./glance --download --iso https://ftp.uni-stuttgart.de/debian-cd/current/amd64/iso-cd/ --scan-timeout 180
```

Notes:

- Default is `60` seconds
- Use `--scan-timeout 0` to disable timeout

Troubleshooting (argument order):

- Wrong: `./glance --download --iso --scan-timeout https://ftp.uni-stuttgart.de/debian-cd/`
- Correct: `./glance --download --iso https://ftp.uni-stuttgart.de/debian-cd/ --scan-timeout 180`
- On malformed input, the CLI now returns an error such as: `unexpected positional arguments: ...`

### 5. Resume an interrupted download

```bash
./glance --download --iso https://www.ututo.org/downloads/Candidato-Ututo-2017-UL.iso
```

If `downloads/Candidato-Ututo-2017-UL.iso.download` already exists, the tool behaves like this:

```text
Resume file found: downloads/Candidato-Ututo-2017-UL.iso.download (44.69 MB). Download will continue from where it left off.
Starting ISO download (1/1)...
Resuming download from 44.69 MB: downloads/Candidato-Ututo-2017-UL.iso.download
```

### 6. Force a clean restart instead of resuming

```bash
./glance --download --no-resume --iso https://www.ututo.org/downloads/Candidato-Ututo-2017-UL.iso
```

### 7. Download into a specific directory

```bash
./glance --download --iso https://www.ututo.org/downloads/Candidato-Ututo-2017-UL.iso --output /srv/iso-cache
```

### 8. Download to an exact destination file path

```bash
./glance --download --iso https://www.ututo.org/downloads/Candidato-Ututo-2017-UL.iso --output-path /srv/releases/custom-ututo.iso
```

Notes:

- `--output-path` only applies to a single ISO download
- for multiple selected ISOs, use `--output`

### 9. Copy a local ISO and calculate its checksum locally

```bash
./glance --download --iso /home/user/isos/archlinux-x86_64.iso
```

### 10. Verify against a specific checksum

```bash
./glance --download --iso https://releases.ubuntu.com/24.04/ubuntu-24.04.4-live-server-amd64.iso --checksum e907d92eeec9df64163a7e454cbc8d7755e8ddc7ed42f99dbc80c40f1a138433
```

## System Administrator Guide

### Mirror and repository guidance

- for hosts that do not allow anonymous FTP, prefer HTTP/HTTPS directory scanning
- large mirrors are scanned with pacing controls to reduce the risk of looking like abusive traffic
- FTP traversal includes delays between directory listing requests to remain polite toward public servers

### SSH/SFTP delivery

Password-based authentication:

```bash
./glance --upload --file ./downloads/ubuntu-24.04.4-live-server-amd64.iso --host 192.168.1.50 --user root --password secret
```

SSH key authentication:

```bash
./glance --upload --file ./downloads/ubuntu-24.04.4-live-server-amd64.iso --host 192.168.1.50 --user root --ssh-key ~/.ssh/id_rsa --known-hosts ~/.ssh/known_hosts
```

Interactive selection without `--file`:

```bash
./glance --upload --iso https://releases.ubuntu.com/24.04/
```

Expected flow:

```text
Local ISO/file candidates for upload:
    1) downloads/ubuntu-24.04.4-live-server-amd64.iso
    2) downloads/debian-12.11.0-amd64-netinst.iso
Select file number [1-2] or type a full path:
SSH host/IP:
SSH username:
Auth method (password/key):
```

Download and upload in one command:

```bash
./glance --download --iso https://releases.ubuntu.com/24.04/ubuntu-24.04.4-live-server-amd64.iso --upload --host 192.168.1.50 --user root --ssh-key ~/.ssh/id_rsa --known-hosts ~/.ssh/known_hosts
```

If the `known_hosts` entry is missing, add it first:

```bash
ssh-keyscan -H 192.168.1.50 >> ~/.ssh/known_hosts
```

### Operational notes

- resume files are stored as the final filename plus `.download`
- when the transfer completes, the temporary file is renamed to the final target path
- if the server ignores HTTP `Range`, the tool safely restarts the full download to avoid corruption
- if an FTP server does not support resume (`REST`), the tool safely restarts the full download

## Developer Guide

### Package layout

- `internal/cli`: flag parsing, prompts, selection flow, and orchestration
- `internal/downloader`: HTTP/FTP scanning, transfers, resume handling, and checksum source resolution
- `internal/verifier`: local checksum calculation and validation
- `internal/uploader`: SSH/SFTP upload and remote checksum validation
- `internal/keygen`: SSH key generation

### Recommended development commands

```bash
go build -o glance .
./glance --help
./glance --download --iso https://www.ututo.org/downloads/
```

### Checksum resolution order

For HTTP/HTTPS:

- `file.iso.sha256sum` / `file.iso.sha512sum` / `file.iso.md5sum`
- `SHA256SUMS` / `SHA512SUMS` / `MD5SUMS`
- `checksum`

For FTP:

- sibling checksum files in the same directory as the ISO
- common checksum index files and uppercase/lowercase variants

If no remote checksum can be used, the tool calculates the file hash locally.

### Interactive control flow

1. If the source is a directory, the tool builds an ISO list.
2. The user selects one or more ISO files.
3. If neither `--output` nor `--output-path` was provided, the tool asks where to download them.
4. If a `.download` file exists, the tool reports that resume will be used.
5. The download finishes and checksum verification runs.
6. If `--upload` is enabled, the verified file is sent to the remote target.
7. During upload, if `--file` is missing, local candidates are listed and the user picks one.

## Complete Flag Reference

- `--download`
- `--no-resume`
- `--upload`
- `--keygen`
- `--key-algo`
- `--key-output`
- `--key-name`
- `--iso`
- `--url`
- `--checksum`
- `--checksum-algo`
- `--output`
- `--output-path`
- `--scan-timeout`
- `--file`
- `--host`
- `--port`
- `--user`
- `--password`
- `--ssh-key`
- `--known-hosts`
- `--remote-path`
- `--license`
- `--help`

## Changelog Note

`ChangeLog.md` is used as a running release and change journal. To verify the newest commit timestamps directly:

```bash
git log --date=iso --pretty=format:"%h %ad %an: %s"
```

## License

MIT License

Copyright (c) 2026 PlusClouds

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
