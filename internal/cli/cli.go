// MIT License
//
// Copyright (c) 2026 PlusClouds
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//
// File purpose: Defines CLI configuration, flags, help text, prompts, and top-level command orchestration.

package cli

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"glance/internal/downloader"
	"glance/internal/keygen"
	licensecontent "glance/internal/license"
	"glance/internal/uploader"
	"glance/internal/verifier"
)

type Config struct {
	Download     bool
	Upload       bool
	Keygen       bool
	KeyAlgo      string
	KeyOutputDir string
	KeyName      string
	ShowLicense  bool
	ShowHelp     bool
	ISOSource    string
	Checksum     string
	ChecksumAlgo string
	OutputDir    string
	FileToUpload string
	Host         string
	Port         int
	User         string
	Password     string
	SSHKey       string
	KnownHosts   string
	RemotePath   string
}

func Parse(args []string) (Config, error) {
	var cfg Config

	if len(args) > 0 {
		switch strings.ToLower(strings.TrimSpace(args[0])) {
		case "help":
			cfg.ShowHelp = true
			return cfg, nil
		case "license":
			cfg.ShowLicense = true
			return cfg, nil
		}
	}

	fs := flag.NewFlagSet("glance", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	fs.BoolVar(&cfg.Download, "download", false, "Download/copy ISO from source")
	fs.BoolVar(&cfg.Upload, "upload", false, "Upload ISO/file to remote server over SSH")
	fs.BoolVar(&cfg.Keygen, "keygen", false, "Generate SSH key pair (ed25519, rsa, ecdsa)")
	fs.StringVar(&cfg.KeyAlgo, "key-algo", "ed25519", "Key algorithm: ed25519, rsa, ecdsa")
	fs.StringVar(&cfg.KeyOutputDir, "key-output", "~/.ssh", "Directory to write generated key pair")
	fs.StringVar(&cfg.KeyName, "key-name", "", "Key filename without extension (default: id_<algo>)")
	fs.BoolVar(&cfg.ShowLicense, "license", false, "Show MIT license and copyright")
	fs.BoolVar(&cfg.ShowHelp, "help", false, "Show help")
	fs.StringVar(&cfg.ISOSource, "iso", "", "ISO source URL/path (required for --download)")
	fs.StringVar(&cfg.ISOSource, "url", "", "ISO source URL/path (alias of --iso)")
	fs.StringVar(&cfg.Checksum, "checksum", "", "Expected checksum for downloaded ISO (required for --download)")
	fs.StringVar(&cfg.ChecksumAlgo, "checksum-algo", "sha256", "Checksum algorithm: sha256, sha512, md5")
	fs.StringVar(&cfg.OutputDir, "output", "./downloads", "Download output directory")
	fs.StringVar(&cfg.FileToUpload, "file", "", "Local file path to upload")
	fs.StringVar(&cfg.Host, "host", "", "Remote SSH host/IP")
	fs.IntVar(&cfg.Port, "port", 22, "Remote SSH port")
	fs.StringVar(&cfg.User, "user", "", "Remote SSH user")
	fs.StringVar(&cfg.Password, "password", "", "Remote SSH password")
	fs.StringVar(&cfg.SSHKey, "ssh-key", "", "SSH private key path")
	fs.StringVar(&cfg.KnownHosts, "known-hosts", "~/.ssh/known_hosts", "Path to SSH known_hosts file")
	fs.StringVar(&cfg.RemotePath, "remote-path", "/tmp", "Remote directory path")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			cfg.ShowHelp = true
			return cfg, nil
		}
		return cfg, err
	}

	if !cfg.Download && !cfg.Upload && !cfg.Keygen && !cfg.ShowLicense && !cfg.ShowHelp {
		cfg.ShowHelp = true
	}

	return cfg, nil
}

func HelpText() string {
	return `glance - modular ISO download/upload CLI

Usage:
  ./glance --keygen
  ./glance --keygen --key-algo ed25519
  ./glance --keygen --key-algo rsa --key-output ~/.ssh --key-name id_rsa_glance
  ./glance --keygen --key-algo ecdsa
  ./glance --download --iso https://releases.ubuntu.com/24.04/ubuntu-24.04.4-live-server-amd64.iso --checksum e907d92eeec9df64163a7e454cbc8d7755e8ddc7ed42f99dbc80c40f1a138433
  ./glance --download --iso /home/user/isos/archlinux-x86_64.iso --checksum <sha256sum>
  ./glance --upload --file ./downloads/ubuntu-24.04.4-live-server-amd64.iso --host 192.168.1.50 --user root --ssh-key ~/.ssh/id_ed25519
  ./glance --download --iso https://releases.ubuntu.com/24.04/ubuntu-24.04.4-live-server-amd64.iso --checksum e907d92eeec9df64163a7e454cbc8d7755e8ddc7ed42f99dbc80c40f1a138433 --upload --host 192.168.1.50 --user root --ssh-key ~/.ssh/id_ed25519
  ./glance --license

Flags:
  --download        Download/copy ISO from source
  --upload          Upload ISO/file over SSH/SFTP
  --keygen          Generate SSH key pair
  --key-algo        Key algorithm: ed25519 (default), rsa, ecdsa
  --key-output      Directory for generated key (default: ~/.ssh)
  --key-name        Key filename without extension (default: id_<algo>)
  --iso             ISO source URL/path (required for --download)
  --url             Alias of --iso
  --checksum        Expected checksum for downloaded ISO (required for --download)
  --checksum-algo   Checksum algorithm: sha256, sha512, md5 (default: sha256)
  --output          Download output directory (default: ./downloads)
  --file            Local file path for upload
  --host            Remote SSH host/IP
  --port            Remote SSH port (default: 22)
  --user            Remote SSH username
  --password        Remote SSH password
  --ssh-key         SSH private key path (rsa, ed25519, ecdsa supported)
  --known-hosts     Path to SSH known_hosts file (default: ~/.ssh/known_hosts)
  --remote-path     Remote upload directory (default: /tmp)
  --license         Print MIT license and copyright
  --help            Show this help

MIT License:
` + licensecontent.Text + `
`
}

func Run(cfg Config) error {
	var downloadedPath string
	var err error

	if cfg.Keygen {
		if err := runKeygen(cfg); err != nil {
			return err
		}
	}

	if cfg.Download {
		if strings.TrimSpace(cfg.ISOSource) == "" {
			source, promptErr := prompt(bufio.NewReader(os.Stdin), "ISO source URL/path")
			if promptErr != nil {
				return promptErr
			}
			cfg.ISOSource = source
		}

		if strings.TrimSpace(cfg.ISOSource) == "" {
			return fmt.Errorf("--iso (or --url) is required for --download")
		}

		fmt.Println("Starting ISO download...")
		downloadedPath, err = downloader.DownloadISO(cfg.ISOSource, cfg.OutputDir)
		if err != nil {
			return err
		}

		if strings.TrimSpace(cfg.Checksum) == "" {
			expectedChecksum, promptErr := prompt(bufio.NewReader(os.Stdin), "Expected checksum")
			if promptErr != nil {
				return promptErr
			}
			cfg.Checksum = expectedChecksum
		}

		if strings.TrimSpace(cfg.Checksum) == "" {
			return fmt.Errorf("--checksum is required for --download")
		}

		if err := verifier.VerifyFileHash(downloadedPath, cfg.Checksum, cfg.ChecksumAlgo); err != nil {
			return err
		}

		fmt.Printf("Checksum verified (%s): %s\n", strings.ToLower(strings.TrimSpace(cfg.ChecksumAlgo)), downloadedPath)
	}

	if cfg.Upload {
		fileToUpload := cfg.FileToUpload
		if fileToUpload == "" {
			if downloadedPath != "" {
				fileToUpload = downloadedPath
			} else if strings.TrimSpace(cfg.ISOSource) != "" {
				if isHTTPSource(cfg.ISOSource) {
					base := filepath.Base(cfg.ISOSource)
					fileToUpload = filepath.Join(cfg.OutputDir, base)
				} else {
					fileToUpload = cfg.ISOSource
				}
			} else {
				return fmt.Errorf("--file is required for --upload when no downloaded ISO is available")
			}
		}

		if err := promptMissingUploadFields(&cfg); err != nil {
			return err
		}

		fmt.Println("Starting upload over SSH/SFTP...")
		u := uploader.Config{
			Host:       cfg.Host,
			Port:       cfg.Port,
			User:       cfg.User,
			Password:   cfg.Password,
			KeyPath:    cfg.SSHKey,
			KnownHosts: cfg.KnownHosts,
			LocalFile:  fileToUpload,
			RemotePath: cfg.RemotePath,
		}
		if err := uploader.UploadFile(u); err != nil {
			return err
		}
	}

	return nil
}

func promptMissingUploadFields(cfg *Config) error {
	reader := bufio.NewReader(os.Stdin)

	if strings.TrimSpace(cfg.Host) == "" {
		host, err := prompt(reader, "SSH host/IP")
		if err != nil {
			return err
		}
		cfg.Host = host
	}

	if strings.TrimSpace(cfg.User) == "" {
		user, err := prompt(reader, "SSH username")
		if err != nil {
			return err
		}
		cfg.User = user
	}

	if strings.TrimSpace(cfg.SSHKey) == "" && strings.TrimSpace(cfg.Password) == "" {
		method, err := prompt(reader, "Auth method (password/key)")
		if err != nil {
			return err
		}
		method = strings.ToLower(strings.TrimSpace(method))
		if method == "key" {
			keyPath, err := prompt(reader, "SSH private key path")
			if err != nil {
				return err
			}
			cfg.SSHKey = keyPath
		} else {
			password, err := prompt(reader, "SSH password")
			if err != nil {
				return err
			}
			cfg.Password = password
		}
	}

	if cfg.Host == "" || cfg.User == "" {
		return fmt.Errorf("host and user are required for upload")
	}

	if cfg.Password == "" && cfg.SSHKey == "" {
		return fmt.Errorf("either password or ssh key is required for upload")
	}

	return nil
}

func prompt(reader *bufio.Reader, label string) (string, error) {
	fmt.Printf("%s: ", label)
	value, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(value), nil
}

func isHTTPSource(source string) bool {
	source = strings.ToLower(strings.TrimSpace(source))
	return strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://")
}

func runKeygen(cfg Config) error {
	outputDir := cfg.KeyOutputDir
	if strings.HasPrefix(outputDir, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("resolve home dir: %w", err)
		}
		outputDir = filepath.Join(home, strings.TrimPrefix(outputDir, "~/"))
	}

	algo := cfg.KeyAlgo
	if algo == "" {
		algo = "ed25519"
	}

	kc := keygen.Config{
		Algorithm: algo,
		OutputDir: outputDir,
		KeyName:   cfg.KeyName,
	}

	privPath, pubPath, err := keygen.GenerateKeyPair(kc)
	if err != nil {
		return err
	}

	fmt.Printf("Generated %s key pair:\n  Private: %s\n  Public:  %s\n", algo, privPath, pubPath)
	fmt.Printf("Add public key to remote server:\n  ssh-copy-id -i %s user@host\n", pubPath)
	return nil
}
