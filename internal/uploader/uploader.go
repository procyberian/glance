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
// File purpose: Uploads files to remote servers over SSH/SFTP with host key verification via known_hosts.

package uploader

import (
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type Config struct {
	Host       string
	Port       int
	User       string
	Password   string
	KeyPath    string
	KnownHosts string
	LocalFile  string
	RemotePath string
	Checksum   string
	Algorithm  string
	NoResume   bool
}

var remoteChecksumPattern = regexp.MustCompile(`(?i)\b([a-f0-9]{32}|[a-f0-9]{64}|[a-f0-9]{128})\b`)

func UploadFile(cfg Config) error {
	if cfg.Host == "" || cfg.User == "" {
		return fmt.Errorf("host and user are required")
	}
	if cfg.LocalFile == "" {
		return fmt.Errorf("local file path is required")
	}
	if cfg.Port == 0 {
		cfg.Port = 22
	}
	if cfg.RemotePath == "" {
		cfg.RemotePath = "/tmp"
	}

	localStat, err := os.Stat(cfg.LocalFile)
	if err != nil {
		return fmt.Errorf("stat local file: %w", err)
	}
	localSize := localStat.Size()

	auth, err := buildAuth(cfg.Password, cfg.KeyPath)
	if err != nil {
		return err
	}

	hostKeyCallback, err := buildHostKeyCallback(cfg.KnownHosts)
	if err != nil {
		return err
	}

	sshConfig := &ssh.ClientConfig{
		User:            cfg.User,
		Auth:            auth,
		HostKeyCallback: hostKeyCallback,
		Timeout:         30 * time.Second,
	}

	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	conn, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return fmt.Errorf("ssh dial: %w", err)
	}
	defer conn.Close()

	client, err := sftp.NewClient(conn)
	if err != nil {
		return fmt.Errorf("create sftp client: %w", err)
	}
	defer client.Close()

	if err := client.MkdirAll(cfg.RemotePath); err != nil {
		return fmt.Errorf("create remote directory: %w", err)
	}

	remoteFile := path.Join(cfg.RemotePath, filepath.Base(cfg.LocalFile))

	// Determine resume offset.
	var remoteSize int64
	if !cfg.NoResume {
		if remoteStat, statErr := client.Stat(remoteFile); statErr == nil {
			remoteSize = remoteStat.Size()
		}
	}

	src, err := os.Open(cfg.LocalFile)
	if err != nil {
		return fmt.Errorf("open local file: %w", err)
	}
	defer src.Close()

	var dst *sftp.File
	if remoteSize > 0 && remoteSize < localSize {
		if _, seekErr := src.Seek(remoteSize, io.SeekStart); seekErr != nil {
			return fmt.Errorf("seek local file for resume: %w", seekErr)
		}
		dst, err = client.OpenFile(remoteFile, os.O_WRONLY|os.O_APPEND)
		if err != nil {
			// Server does not support append; fall back to a fresh upload.
			remoteSize = 0
			if _, seekErr := src.Seek(0, io.SeekStart); seekErr != nil {
				return fmt.Errorf("seek local file: %w", seekErr)
			}
			dst, err = client.Create(remoteFile)
		} else {
			fmt.Printf("Resuming upload from %.2f MB: %s\n", float64(remoteSize)/(1024*1024), remoteFile)
		}
	} else {
		remoteSize = 0
		dst, err = client.Create(remoteFile)
	}
	if err != nil {
		return fmt.Errorf("open remote file: %w", err)
	}
	defer dst.Close()

	fmt.Printf("Uploading %s to %s:%s...\n", cfg.LocalFile, cfg.Host, remoteFile)
	written, err := uploadWithProgress(src, dst, localSize, remoteSize)
	if err != nil {
		return fmt.Errorf("upload file: %w", err)
	}

	fmt.Printf("Uploaded %s to %s:%s (%.2f MB)\n", cfg.LocalFile, cfg.Host, remoteFile, float64(written)/(1024*1024))

	if strings.TrimSpace(cfg.Checksum) != "" {
		if err := verifyRemoteChecksum(conn, remoteFile, cfg.Checksum, cfg.Algorithm); err != nil {
			return err
		}
		fmt.Println("Remote checksum verified")
	}

	return nil
}

func uploadWithProgress(src io.Reader, dst io.Writer, total int64, initialWritten int64) (int64, error) {
	buf := make([]byte, 64*1024)
	start := time.Now()
	lastPrint := start
	written := initialWritten
	lastBytes := initialWritten

	for {
		n, readErr := src.Read(buf)
		if n > 0 {
			wn, writeErr := dst.Write(buf[:n])
			written += int64(wn)
			if writeErr != nil {
				return written, writeErr
			}
			if wn != n {
				return written, io.ErrShortWrite
			}
		}

		now := time.Now()
		if now.Sub(lastPrint) >= time.Second || readErr == io.EOF {
			intervalSeconds := now.Sub(lastPrint).Seconds()
			if intervalSeconds <= 0 {
				intervalSeconds = 1
			}
			elapsedSeconds := now.Sub(start).Seconds()
			if elapsedSeconds <= 0 {
				elapsedSeconds = 1
			}
			instantBps := float64(written-lastBytes) / intervalSeconds
			avgBps := float64(written-initialWritten) / elapsedSeconds
			printUploadProgress(written, total, instantBps, avgBps)
			lastPrint = now
			lastBytes = written
		}

		if readErr == io.EOF {
			fmt.Println()
			break
		}
		if readErr != nil {
			return written, readErr
		}
	}

	return written, nil
}

func printUploadProgress(written, total int64, instantBps, avgBps float64) {
	if total > 0 {
		percent := (float64(written) / float64(total)) * 100
		remainingBytes := float64(total - written)
		eta := "bilinmiyor"
		if avgBps > 0 && remainingBytes > 0 {
			etaDuration := time.Duration((remainingBytes / avgBps) * float64(time.Second))
			eta = etaDuration.Truncate(time.Second).String()
		} else if remainingBytes <= 0 {
			eta = "0s"
		}
		fmt.Printf(
			"\r%.2f%% | %s/%s | Anlik: %s/s | Ortalama: %s/s | ETA: %s",
			percent,
			formatUploadBytes(float64(written)),
			formatUploadBytes(float64(total)),
			formatUploadBytes(instantBps),
			formatUploadBytes(avgBps),
			eta,
		)
		return
	}
	fmt.Printf(
		"\r%s yuklendi | Anlik: %s/s | Ortalama: %s/s",
		formatUploadBytes(float64(written)),
		formatUploadBytes(instantBps),
		formatUploadBytes(avgBps),
	)
}

func formatUploadBytes(v float64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	idx := 0
	for v >= 1024 && idx < len(units)-1 {
		v /= 1024
		idx++
	}
	return fmt.Sprintf("%.2f %s", v, units[idx])
}

func verifyRemoteChecksum(conn *ssh.Client, remoteFile, expectedChecksum, algorithm string) error {
	commands, err := checksumCommands(remoteFile, algorithm)
	if err != nil {
		return err
	}
	var lastErr error
	for _, command := range commands {
		actualChecksum, err := runRemoteChecksum(conn, command)
		if err != nil {
			lastErr = err
			continue
		}

		if actualChecksum != strings.ToLower(strings.TrimSpace(expectedChecksum)) {
			return fmt.Errorf("remote checksum verification failed: expected %s, got %s", strings.ToLower(strings.TrimSpace(expectedChecksum)), actualChecksum)
		}

		return nil
	}

	if lastErr != nil {
		return fmt.Errorf("remote checksum command failed: %w", lastErr)
	}

	return fmt.Errorf("remote checksum command failed")
}

func runRemoteChecksum(conn *ssh.Client, command string) (string, error) {
	session, err := conn.NewSession()
	if err != nil {
		return "", fmt.Errorf("create ssh session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return "", fmt.Errorf("run %q: %w (%s)", command, err, strings.TrimSpace(string(output)))
	}

	checksum := remoteChecksumPattern.FindString(strings.ToLower(string(output)))
	if checksum == "" {
		return "", fmt.Errorf("no checksum found in command output: %s", strings.TrimSpace(string(output)))
	}

	return checksum, nil
}

func checksumCommands(remoteFile, algorithm string) ([]string, error) {
	quotedPath := shellQuote(remoteFile)
	algo, err := normalizeAlgorithm(algorithm)
	if err != nil {
		return nil, err
	}

	switch algo {
	case "sha512":
		return []string{
			"sha512sum " + quotedPath,
			"shasum -a 512 " + quotedPath,
		}, nil
	case "sha256":
		return []string{
			"sha256sum " + quotedPath,
			"shasum -a 256 " + quotedPath,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported checksum algorithm: %s (supported: sha256, sha512)", algorithm)
	}
}

func normalizeAlgorithm(algorithm string) (string, error) {
	algorithm = strings.ToLower(strings.TrimSpace(algorithm))
	if algorithm == "" {
		return "sha256", nil
	}
	if algorithm != "sha256" && algorithm != "sha512" {
		return "", fmt.Errorf("unsupported checksum algorithm: %s (supported: sha256, sha512)", algorithm)
	}
	return algorithm, nil
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func buildAuth(password, keyPath string) ([]ssh.AuthMethod, error) {
	if keyPath != "" {
		key, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("read key file: %w", err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("parse key file: %w", err)
		}
		return []ssh.AuthMethod{ssh.PublicKeys(signer)}, nil
	}

	if password == "" {
		return nil, fmt.Errorf("either password or key path must be provided")
	}
	return []ssh.AuthMethod{ssh.Password(password)}, nil
}

func buildHostKeyCallback(knownHostsPath string) (ssh.HostKeyCallback, error) {
	resolvedPath, err := resolvePath(knownHostsPath)
	if err != nil {
		return nil, fmt.Errorf("resolve known_hosts path: %w", err)
	}

	if _, err := os.Stat(resolvedPath); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("known_hosts file not found at %s; create it with: ssh-keyscan -H <host> >> %s", resolvedPath, resolvedPath)
		}
		return nil, fmt.Errorf("stat known_hosts: %w", err)
	}

	cb, err := knownhosts.New(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("load known_hosts file: %w", err)
	}

	return cb, nil
}

func resolvePath(p string) (string, error) {
	p = strings.TrimSpace(p)
	if p == "" {
		return "", fmt.Errorf("known_hosts path is required")
	}

	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, strings.TrimPrefix(p, "~/")), nil
	}

	return p, nil
}
