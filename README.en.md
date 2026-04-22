# glance-cli

Modular Go CLI tool.

Purpose:
- Fetch any operating system ISO from a URL or local path
- Upload the ISO file to a target server over SSH/SFTP
- Display the MIT license text from the command line

## Features

- `--download`: Downloads from the provided ISO source or copies from a local path
- `--iso`: ISO source URL/path
- `--checksum`: Hash verification after download/copy (required)
- `--checksum-algo`: sha256 (default), sha512, md5
- Live progress during transfer: percentage, instant/average speed, network speed, and ETA
- `--upload`: Uploads an ISO or file to a remote server
- Authentication with password or SSH key
- SSH host key verification using the `known_hosts` file
- Prompts for missing `--host` and `--user` values
- `--license`: Prints MIT license and copyright text

## Build

```bash
go mod tidy
go build -o glance .
```

## Usage

Download ISO only:

```bash
./glance --download --iso https://releases.ubuntu.com/24.04/ubuntu-24.04.4-live-server-amd64.iso --checksum e907d92eeec9df64163a7e454cbc8d7755e8ddc7ed42f99dbc80c40f1a138433
```

Download and verify with a specific checksum:

```bash
./glance --download --iso https://releases.ubuntu.com/24.04/ubuntu-24.04.4-live-server-amd64.iso --checksum e907d92eeec9df64163a7e454cbc8d7755e8ddc7ed42f99dbc80c40f1a138433
```

Copy from a local ISO path:

```bash
./glance --download --iso /home/user/isos/archlinux-x86_64.iso
```

Verify a local file with sha512:

```bash
./glance --download --iso /home/user/isos/archlinux-x86_64.iso --checksum <sha512sum> --checksum-algo sha512
```

Upload downloaded ISO (password auth):

```bash
./glance --upload --file ./downloads/ubuntu-24.04.4-live-server-amd64.iso --host 192.168.1.50 --user root --password secret
```

Upload downloaded ISO (SSH key auth):

```bash
./glance --upload --file ./downloads/ubuntu-24.04.4-live-server-amd64.iso --host 192.168.1.50 --user root --ssh-key ~/.ssh/id_rsa --known-hosts ~/.ssh/known_hosts
```

Download and upload in one command:

```bash
./glance --download --iso https://releases.ubuntu.com/24.04/ubuntu-24.04.4-live-server-amd64.iso --checksum e907d92eeec9df64163a7e454cbc8d7755e8ddc7ed42f99dbc80c40f1a138433 --upload --host 192.168.1.50 --user root --ssh-key ~/.ssh/id_rsa --known-hosts ~/.ssh/known_hosts
```

Note: The `--download` operation does not complete without checksum verification. If you omit `--checksum`, the command prompts you interactively.

If the `known_hosts` entry is missing, add it first with:

```bash
ssh-keyscan -H 192.168.1.50 >> ~/.ssh/known_hosts
```

Show license text:

```bash
./glance --license
```

Help:

```bash
./glance --help
```

## Changelog Note

Commit timestamps in `ChangeLog.md` may occasionally appear one step behind.
To see the latest commit timing, run `git log` in your terminal or check the log/history view in your Git UI.

Linux/macOS (Terminal):

```bash
git log --date=iso --pretty=format:"%h %ad %an: %s"
```

Windows (PowerShell):

```powershell
git log --date=iso --pretty=format:"%h %ad %an: %s"
```

Windows (Git Bash):

```bash
git log --date=iso --pretty=format:"%h %ad %an: %s"
```

Note: You can also see commit timestamps directly in Git UIs such as VS Code Source Control, GitKraken, and GitHub Desktop.

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
