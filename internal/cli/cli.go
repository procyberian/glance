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
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/procyberian/glance/v11/internal/downloader"
	"github.com/procyberian/glance/v11/internal/keygen"
	licensecontent "github.com/procyberian/glance/v11/internal/license"
	"github.com/procyberian/glance/v11/internal/uploader"
	"github.com/procyberian/glance/v11/internal/verifier"
)

// Version is the current release version of glance.
const Version = "v11.0.8"

type Config struct {
	Download         bool
	NoResume         bool
	Upload           bool
	ScanTimeout      int
	Keygen           bool
	KeyAlgo          string
	KeyOutputDir     string
	KeyName          string
	ShowLicense      bool
	ShowHelp         bool
	ShowVersion      bool
	JSON             bool
	ConnectTimeout   int
	ISOSource        string
	Checksum         string
	ChecksumAlgo     string
	AllowInsecureFTP bool
	OutputDir        string
	OutputPath       string
	OutputSet        bool
	OutputPathSet    bool
	FileToUpload     string
	Host             string
	Port             int
	User             string
	Password         string
	SSHKey           string
	KnownHosts       string
	RemotePath       string
}

// RunOutput holds structured results emitted when --json is used.
type RunOutput struct {
	DownloadedPath string `json:"downloaded_path,omitempty"`
	Checksum       string `json:"checksum,omitempty"`
	Algorithm      string `json:"checksum_algorithm,omitempty"`
	Uploaded       bool   `json:"uploaded,omitempty"`
	RemoteHost     string `json:"remote_host,omitempty"`
	RemoteFile     string `json:"remote_file,omitempty"`
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
		case "version":
			cfg.ShowVersion = true
			return cfg, nil
		}
	}

	fs := flag.NewFlagSet("glance", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	fs.BoolVar(&cfg.Download, "download", false, "Download/copy ISO from source")
	fs.BoolVar(&cfg.NoResume, "no-resume", false, "Do not resume from .download files; restart download from zero")
	fs.BoolVar(&cfg.Upload, "upload", false, "Upload ISO/file to remote server over SSH")
	fs.IntVar(&cfg.ScanTimeout, "scan-timeout", 60, "Directory scan timeout in seconds for HTTP/FTP listing (0 disables timeout)")
	fs.BoolVar(&cfg.Keygen, "keygen", false, "Generate SSH key pair (ed25519, rsa, ecdsa)")
	fs.StringVar(&cfg.KeyAlgo, "key-algo", "ed25519", "Key algorithm: ed25519, rsa, ecdsa")
	fs.StringVar(&cfg.KeyOutputDir, "key-output", "~/.ssh", "Directory to write generated key pair")
	fs.StringVar(&cfg.KeyName, "key-name", "", "Key filename without extension (default: id_<algo>)")
	fs.BoolVar(&cfg.ShowLicense, "license", false, "Show MIT license and copyright")
	fs.BoolVar(&cfg.ShowHelp, "help", false, "Show help")
	fs.BoolVar(&cfg.ShowVersion, "version", false, "Show version")
	fs.BoolVar(&cfg.JSON, "json", false, "Output results as JSON (machine-readable)")
	fs.IntVar(&cfg.ConnectTimeout, "connect-timeout", 30, "TCP connection timeout in seconds for HTTP (0 disables timeout)")
	fs.StringVar(&cfg.ISOSource, "iso", "", "ISO source URL/path (required for --download)")
	fs.StringVar(&cfg.ISOSource, "url", "", "ISO source URL/path (alias of --iso)")
	fs.BoolVar(&cfg.AllowInsecureFTP, "allow-insecure-ftp", false, "Allow insecure FTP sources (not recommended)")
	fs.StringVar(&cfg.Checksum, "checksum", "", "Expected checksum for downloaded ISO (optional; auto-discovered for HTTP/FTP or calculated locally)")
	fs.StringVar(&cfg.ChecksumAlgo, "checksum-algo", "sha256", "Checksum algorithm: sha256, sha512")
	fs.StringVar(&cfg.OutputDir, "output", "./downloads", "Download output directory")
	fs.StringVar(&cfg.OutputPath, "output-path", "", "Full destination file path for a single downloaded ISO")
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

	if extra := fs.Args(); len(extra) > 0 {
		return cfg, fmt.Errorf("unexpected positional arguments: %s", strings.Join(extra, " "))
	}

	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "output":
			cfg.OutputSet = true
		case "output-path":
			cfg.OutputPathSet = true
		}
	})

	if !cfg.Download && !cfg.Upload && !cfg.Keygen && !cfg.ShowLicense && !cfg.ShowHelp && !cfg.ShowVersion {
		cfg.ShowHelp = true
	}

	if cfg.ScanTimeout < 0 {
		return cfg, fmt.Errorf("--scan-timeout must be >= 0")
	}

	if cfg.Download && strings.HasPrefix(strings.TrimSpace(cfg.ISOSource), "-") {
		return cfg, fmt.Errorf("invalid --iso value %q; expected URL/path (example: --iso https://server/path/ --scan-timeout 60)", strings.TrimSpace(cfg.ISOSource))
	}

	if cfg.Download && isFTPSource(cfg.ISOSource) && !cfg.AllowInsecureFTP {
		return cfg, fmt.Errorf("ftp:// sources are disabled by default; re-run with --allow-insecure-ftp to accept insecure FTP transport")
	}

	algo := normalizedAlgo(cfg.ChecksumAlgo)
	switch algo {
	case "sha256", "sha512":
	default:
		return cfg, fmt.Errorf("unsupported checksum algorithm: %s (supported: sha256, sha512)", cfg.ChecksumAlgo)
	}

	cfg.ChecksumAlgo = algo

	return cfg, nil
}

func HelpText() string {
	return `glance - modular ISO download/upload CLI

Usage:
  ./glance --keygen
  ./glance --keygen --key-algo ed25519
  ./glance --keygen --key-algo rsa --key-output ~/.ssh --key-name id_rsa_glance
  ./glance --keygen --key-algo ecdsa
	./glance --download --iso https://releases.ubuntu.com/24.04/ubuntu-24.04.4-live-server-amd64.iso
	./glance --download --allow-insecure-ftp --iso ftp://ftp.example.com/iso/
	./glance --download --iso https://releases.ubuntu.com/24.04/ubuntu-24.04.4-live-server-amd64.iso --checksum e907d92eeec9df64163a7e454cbc8d7755e8ddc7ed42f99dbc80c40f1a138433
  ./glance --download --iso /home/user/isos/archlinux-x86_64.iso --checksum <sha256sum>
  ./glance --upload --file ./downloads/ubuntu-24.04.4-live-server-amd64.iso --host 192.168.1.50 --user root --ssh-key ~/.ssh/id_ed25519
	./glance --download --iso https://releases.ubuntu.com/24.04/ubuntu-24.04.4-live-server-amd64.iso --upload --host 192.168.1.50 --user root --ssh-key ~/.ssh/id_ed25519
  ./glance --license

Flags:
  --download        Download/copy ISO from source
	--no-resume       Ignore existing .download files and restart download from zero
  --upload          Upload ISO/file over SSH/SFTP
	--scan-timeout    Directory scan timeout in seconds for HTTP/FTP listing (default: 60, 0 disables timeout)
  --keygen          Generate SSH key pair
  --key-algo        Key algorithm: ed25519 (default), rsa, ecdsa
  --key-output      Directory for generated key (default: ~/.ssh)
  --key-name        Key filename without extension (default: id_<algo>)
	--iso             ISO source URL/path (HTTP, FTP, local) (required for --download)
  --url             Alias of --iso
	--allow-insecure-ftp Allow insecure FTP sources (not recommended)
	--checksum        Expected checksum for downloaded ISO (optional; auto-discovered for HTTP/FTP or calculated locally)
	--checksum-algo   Checksum algorithm: sha256, sha512 (default: sha256)
  --output          Download output directory (default: ./downloads)
	--output-path     Full destination file path for a single downloaded ISO
  --file            Local file path for upload
  --host            Remote SSH host/IP
  --port            Remote SSH port (default: 22)
  --user            Remote SSH username
  --password        Remote SSH password
  --ssh-key         SSH private key path (rsa, ed25519, ecdsa supported)
  --known-hosts     Path to SSH known_hosts file (default: ~/.ssh/known_hosts)
  --remote-path     Remote upload directory (default: /tmp)
  --connect-timeout TCP connection timeout in seconds for HTTP (default: 30, 0 disables)
  --json            Output results as JSON (machine-readable)
  --version         Show version
  --license         Print MIT license and copyright
  --help            Show this help

MIT License:
` + licensecontent.Text + `
`
}

func Run(cfg Config) error {
	if cfg.ShowVersion {
		fmt.Printf("glance %s\n", Version)
		return nil
	}

	var output RunOutput
	var downloadedPath string

	algo := normalizedAlgo(cfg.ChecksumAlgo)
	if algo != "sha256" && algo != "sha512" {
		return fmt.Errorf("unsupported checksum algorithm: %s (supported: sha256, sha512)", cfg.ChecksumAlgo)
	}
	cfg.ChecksumAlgo = algo

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

		if isFTPSource(cfg.ISOSource) && !cfg.AllowInsecureFTP {
			return fmt.Errorf("ftp:// sources are disabled by default; re-run with --allow-insecure-ftp to accept insecure FTP transport")
		}

		type selectedDownload struct {
			Source   string
			Checksum string
		}

		selectedDownloads := make([]selectedDownload, 0)

		scanTimeout := scanTimeoutDuration(cfg.ScanTimeout)

		if isHTTPDirectorySource(cfg.ISOSource) {
			selectedOptions, selectErr := promptHTTPISOSelections(cfg.ISOSource, cfg.ChecksumAlgo, scanTimeout, cfg.AllowInsecureFTP)
			if selectErr != nil {
				return selectErr
			}
			for _, option := range selectedOptions {
				selectedDownloads = append(selectedDownloads, selectedDownload{Source: option.URL, Checksum: option.Checksum})
			}
		} else if isFTPDirectorySource(cfg.ISOSource) {
			selectedOptions, selectErr := promptFTPISOSelections(cfg.ISOSource, cfg.ChecksumAlgo, scanTimeout)
			if selectErr != nil {
				return selectErr
			}
			for _, option := range selectedOptions {
				selectedDownloads = append(selectedDownloads, selectedDownload{Source: option.URL, Checksum: option.Checksum})
			}
		} else {
			selectedDownloads = append(selectedDownloads, selectedDownload{Source: cfg.ISOSource, Checksum: cfg.Checksum})
		}

		if cfg.Upload && len(selectedDownloads) > 1 {
			return fmt.Errorf("multiple ISO selection is not supported together with --upload; choose one ISO or run upload separately")
		}

		if strings.TrimSpace(cfg.OutputPath) != "" && len(selectedDownloads) > 1 {
			return fmt.Errorf("--output-path can only be used when downloading a single ISO; use --output for multiple selections")
		}

		if len(selectedDownloads) == 1 && isHTTPSource(selectedDownloads[0].Source) && !isHTTPSourceLikelyISO(selectedDownloads[0].Source) {
			return fmt.Errorf("http/https source must be a direct ISO URL; for listing use an FTP server address like ftp://server or https://server/")
		}

		if !cfg.OutputSet && !cfg.OutputPathSet {
			outputDir, promptErr := promptDownloadDirectory(cfg.OutputDir)
			if promptErr != nil {
				return promptErr
			}
			cfg.OutputDir = outputDir
		}

		for i, item := range selectedDownloads {
			targetPath := ""
			if len(selectedDownloads) == 1 {
				targetPath = strings.TrimSpace(cfg.OutputPath)
			}
			if !cfg.NoResume && isHTTPSource(item.Source) {
				if resumePath, resumeSize, ok := findResumeFile(item.Source, cfg.OutputDir, targetPath); ok {
					fmt.Printf("Resume file found: %s (%.2f MB). Download will continue from where it left off.\n", resumePath, float64(resumeSize)/(1024*1024))
				}
			}
			fmt.Printf("Starting ISO download (%d/%d)...\n", i+1, len(selectedDownloads))
			connectTimeout := time.Duration(cfg.ConnectTimeout) * time.Second
			path, checksum, dlErr := downloadAndVerifySource(item.Source, item.Checksum, cfg.OutputDir, targetPath, cfg.ChecksumAlgo, !cfg.NoResume, connectTimeout)
			if dlErr != nil {
				return dlErr
			}
			fmt.Println("Checksum verified")
			downloadedPath = path
			cfg.ISOSource = item.Source
			cfg.Checksum = checksum
			output.DownloadedPath = path
			output.Checksum = checksum
			output.Algorithm = normalizedAlgo(cfg.ChecksumAlgo)
		}
	}

	if cfg.Upload {
		fileToUpload := cfg.FileToUpload
		if fileToUpload == "" {
			if downloadedPath != "" {
				fileToUpload = downloadedPath
			} else if strings.TrimSpace(cfg.ISOSource) != "" {
				if isHTTPSource(cfg.ISOSource) || isFTPSource(cfg.ISOSource) {
					selectedPath, selectErr := promptUploadFileSelection(cfg.OutputDir, cfg.ISOSource)
					if selectErr != nil {
						return selectErr
					}
					fileToUpload = selectedPath
				} else {
					fileToUpload = cfg.ISOSource
				}
			} else {
				selectedPath, selectErr := promptUploadFileSelection(cfg.OutputDir, "")
				if selectErr != nil {
					return selectErr
				}
				fileToUpload = selectedPath
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
			Checksum:   cfg.Checksum,
			Algorithm:  cfg.ChecksumAlgo,
			NoResume:   cfg.NoResume,
		}
		if err := uploader.UploadFile(u); err != nil {
			return err
		}
		output.Uploaded = true
		output.RemoteHost = cfg.Host
		output.RemoteFile = cfg.RemotePath + "/" + filepath.Base(fileToUpload)
	}

	if cfg.JSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(output); err != nil {
			return fmt.Errorf("encode json output: %w", err)
		}
	}

	return nil
}

func downloadAndVerifySource(source, providedChecksum, outputDir, outputPath, checksumAlgo string, allowResume bool, connectTimeout time.Duration) (string, string, error) {
	downloadedPath, err := downloader.DownloadISOWithConnectTimeout(source, outputDir, outputPath, allowResume, connectTimeout)
	if err != nil {
		return "", "", err
	}

	checksum := strings.TrimSpace(providedChecksum)
	if checksum == "" {
		if isHTTPSource(source) || isFTPSource(source) {
			expectedChecksum, resolveErr := downloader.ResolveChecksum(source, checksumAlgo)
			if resolveErr == nil {
				checksum = expectedChecksum
				fmt.Printf("Checksum auto-discovered (%s): %s\n", normalizedAlgo(checksumAlgo), checksum)
			}
		}
	}

	if checksum == "" {
		calculatedChecksum, algoName, calcErr := verifier.CalculateFileHash(downloadedPath, checksumAlgo)
		if calcErr != nil {
			return "", "", calcErr
		}
		checksum = calculatedChecksum
		fmt.Printf("Checksum calculated locally (%s): %s\n", algoName, checksum)
	}

	if strings.TrimSpace(checksum) == "" {
		return "", "", fmt.Errorf("--checksum is required for --download")
	}

	if err := verifier.VerifyFileHash(downloadedPath, checksum, checksumAlgo); err != nil {
		return "", "", err
	}

	return downloadedPath, checksum, nil
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

func promptDownloadDirectory(defaultDir string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	label := fmt.Sprintf("Which directory should the ISO be downloaded to? [default: %s]", defaultDir)
	value, err := prompt(reader, label)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(value) == "" {
		return defaultDir, nil
	}
	return strings.TrimSpace(value), nil
}

func promptUploadFileSelection(defaultDir, sourceHint string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	candidates := listLocalISOCandidates(defaultDir)

	if len(candidates) > 0 {
		fmt.Println("Local ISO/file candidates for upload:")
		for i, candidate := range candidates {
			fmt.Printf("  %d) %s\n", i+1, candidate)
		}

		for {
			selection, err := prompt(reader, fmt.Sprintf("Select file number [1-%d] or type a full path", len(candidates)))
			if err != nil {
				return "", err
			}

			trimmed := strings.TrimSpace(selection)
			if trimmed == "" {
				continue
			}

			if value, convErr := strconv.Atoi(trimmed); convErr == nil {
				if value >= 1 && value <= len(candidates) {
					return candidates[value-1], nil
				}
				fmt.Printf("Invalid selection: '%d' must be between 1 and %d\n", value, len(candidates))
				continue
			}

			return trimmed, nil
		}
	}

	label := "Local file path to upload"
	if strings.TrimSpace(sourceHint) != "" {
		label = fmt.Sprintf("No local file found for source %s. Enter local file path to upload", strings.TrimSpace(sourceHint))
	}
	value, err := prompt(reader, label)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("a local file path is required for --upload")
	}
	return strings.TrimSpace(value), nil
}

func listLocalISOCandidates(rootDir string) []string {
	trimmedRoot := strings.TrimSpace(rootDir)
	if trimmedRoot == "" {
		trimmedRoot = "."
	}

	entries, err := os.ReadDir(trimmedRoot)
	if err != nil {
		return nil
	}

	candidates := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		lowerName := strings.ToLower(name)
		if strings.HasSuffix(lowerName, ".download") {
			continue
		}
		if strings.HasSuffix(lowerName, ".iso") || strings.HasSuffix(lowerName, ".img") || strings.HasSuffix(lowerName, ".qcow2") {
			candidates = append(candidates, filepath.Join(trimmedRoot, name))
		}
	}

	sort.Strings(candidates)
	return candidates
}

func isHTTPSource(source string) bool {
	source = strings.ToLower(strings.TrimSpace(source))
	return strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://")
}

func isFTPSource(source string) bool {
	source = strings.ToLower(strings.TrimSpace(source))
	return strings.HasPrefix(source, "ftp://")
}

func isFTPDirectorySource(source string) bool {
	trimmed := strings.TrimSpace(source)
	if !isFTPSource(trimmed) {
		return false
	}
	lower := strings.ToLower(trimmed)
	if strings.HasSuffix(lower, "/") {
		return true
	}
	return !strings.HasSuffix(lower, ".iso")
}

func isHTTPDirectorySource(source string) bool {
	trimmed := strings.TrimSpace(source)
	if !isHTTPSource(trimmed) {
		return false
	}
	lower := strings.ToLower(trimmed)
	if strings.HasSuffix(lower, ".iso") {
		return false
	}
	if strings.Contains(lower, "?") {
		return false
	}
	return strings.HasSuffix(lower, "/")
}

func isHTTPSourceLikelyISO(source string) bool {
	trimmed := strings.ToLower(strings.TrimSpace(source))
	return strings.HasSuffix(trimmed, ".iso") || strings.Contains(trimmed, ".iso?")
}

func shouldTryFTPFirst(source string) bool {
	if !isHTTPDirectorySource(source) {
		return false
	}

	u, err := url.Parse(strings.TrimSpace(source))
	if err != nil {
		return false
	}

	host := strings.ToLower(strings.TrimSpace(u.Hostname()))
	if host == "" {
		return false
	}

	// Try FTP-first only for clearly FTP-like hosts.
	return strings.HasPrefix(host, "ftp.") || strings.Contains(host, ".ftp.")
}

func promptHTTPISOSelections(source, algorithm string, scanTimeout time.Duration, allowInsecureFTP bool) ([]downloader.FTPISOOption, error) {
	if allowInsecureFTP && shouldTryFTPFirst(source) {
		ftpSource, ok := toFTPDirectoryURL(source)
		if !ok {
			goto httpFallback
		}

		fmt.Printf("Attempting FTP scan first: %s\n", ftpSource)
		if options, err := listFTPWithTimeout(ftpSource, algorithm, scanTimeout); err == nil && len(options) > 0 {
			fmt.Println("Using FTP scan results")
			return promptISOSelections("FTP ISO list:", options, algorithm)
		} else if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "login") {
				fmt.Println("FTP scan unavailable")
			} else {
				fmt.Printf("FTP scan failed: %v\n", err)
			}
		}
		fmt.Println("FTP scan unavailable, falling back to HTTP scan")
	}

httpFallback:
	options, err := listHTTPWithTimeout(source, algorithm, scanTimeout)
	if err != nil {
		return nil, err
	}
	return promptISOSelections("HTTP/HTTPS ISO list:", options, algorithm)
}

func promptFTPISOSelections(source, algorithm string, scanTimeout time.Duration) ([]downloader.FTPISOOption, error) {
	options, err := listFTPWithTimeout(source, algorithm, scanTimeout)
	if err != nil {
		return nil, err
	}
	return promptISOSelections("FTP ISO list:", options, algorithm)
}

func scanTimeoutDuration(seconds int) time.Duration {
	if seconds <= 0 {
		return 0
	}
	return time.Duration(seconds) * time.Second
}

func promptISOSelections(title string, options []downloader.FTPISOOption, algorithm string) ([]downloader.FTPISOOption, error) {
	fmt.Println(title)
	for i, option := range options {
		checksumText := option.Checksum
		if strings.TrimSpace(checksumText) == "" {
			checksumText = "not found"
		}
		fmt.Printf("  %d) %s | size: %s | checksum: %s\n", i+1, option.Name, downloader.FormatSize(option.Size), checksumText)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		selection, promptErr := prompt(reader, fmt.Sprintf("Select ISO number(s) (example: 1 or 1,3,5 or all) [1-%d]", len(options)))
		if promptErr != nil {
			return nil, promptErr
		}

		indices, parseErr := parseSelectionInput(selection, len(options))
		if parseErr != nil {
			fmt.Printf("Invalid selection: %v\n", parseErr)
			continue
		}

		selected := make([]downloader.FTPISOOption, 0, len(indices))
		for _, idx := range indices {
			item := options[idx]
			selected = append(selected, item)
			fmt.Printf("Selected ISO: %s\n", item.URL)
			if strings.TrimSpace(item.Checksum) != "" {
				fmt.Printf("Checksum auto-discovered (%s): %s\n", normalizedAlgo(algorithm), item.Checksum)
			}
		}
		return selected, nil
	}
}

func parseSelectionInput(input string, max int) ([]int, error) {
	trimmed := strings.ToLower(strings.TrimSpace(input))
	if trimmed == "" {
		return nil, fmt.Errorf("selection cannot be empty")
	}
	if trimmed == "all" || trimmed == "*" {
		all := make([]int, 0, max)
		for i := 0; i < max; i++ {
			all = append(all, i)
		}
		return all, nil
	}

	parts := strings.FieldsFunc(trimmed, func(r rune) bool {
		return r == ',' || r == ' ' || r == ';'
	})
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid selection")
	}

	seen := map[int]bool{}
	indices := make([]int, 0, len(parts))
	for _, part := range parts {
		value, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("'%s' is not a number", part)
		}
		if value < 1 || value > max {
			return nil, fmt.Errorf("'%d' must be between 1 and %d", value, max)
		}
		idx := value - 1
		if !seen[idx] {
			seen[idx] = true
			indices = append(indices, idx)
		}
	}

	sort.Ints(indices)
	return indices, nil
}

func normalizedAlgo(algorithm string) string {
	algo := strings.ToLower(strings.TrimSpace(algorithm))
	if algo == "" {
		return "sha256"
	}
	return algo
}

func findResumeFile(source, outputDir, outputPath string) (string, int64, bool) {
	targetPath := strings.TrimSpace(outputPath)
	if targetPath == "" {
		filename := filepath.Base(strings.TrimSpace(source))
		if filename == "." || filename == "/" || filename == "" {
			filename = "image.iso"
		}
		targetPath = filepath.Join(outputDir, filename)
	}
	tempPath := targetPath + ".download"
	stat, err := os.Stat(tempPath)
	if err != nil || stat.Size() <= 0 {
		return "", 0, false
	}
	return tempPath, stat.Size(), true
}

func toFTPDirectoryURL(source string) (string, bool) {
	u, err := url.Parse(strings.TrimSpace(source))
	if err != nil {
		return "", false
	}
	if !strings.EqualFold(u.Scheme, "http") && !strings.EqualFold(u.Scheme, "https") {
		return "", false
	}
	if u.Hostname() == "" {
		return "", false
	}
	if u.Path == "" {
		u.Path = "/"
	}
	if !strings.HasSuffix(u.Path, "/") {
		u.Path += "/"
	}
	u.Scheme = "ftp"
	u.RawQuery = ""
	u.Fragment = ""
	if port := u.Port(); port == "80" || port == "443" {
		u.Host = u.Hostname()
	}
	return u.String(), true
}

func listFTPWithTimeout(source, algorithm string, timeout time.Duration) ([]downloader.FTPISOOption, error) {
	return downloader.ListFTPISOsWithTimeout(source, algorithm, timeout)
}

func listHTTPWithTimeout(source, algorithm string, timeout time.Duration) ([]downloader.FTPISOOption, error) {
	return downloader.ListHTTPISOsWithTimeout(source, algorithm, timeout)
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
