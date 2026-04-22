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
}

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

	src, err := os.Open(cfg.LocalFile)
	if err != nil {
		return fmt.Errorf("open local file: %w", err)
	}
	defer src.Close()

	remoteFile := path.Join(cfg.RemotePath, filepath.Base(cfg.LocalFile))
	dst, err := client.Create(remoteFile)
	if err != nil {
		return fmt.Errorf("create remote file: %w", err)
	}
	defer dst.Close()

	written, err := io.Copy(dst, src)
	if err != nil {
		return fmt.Errorf("upload file: %w", err)
	}

	fmt.Printf("Uploaded %s to %s:%s (%.2f MB)\n", cfg.LocalFile, cfg.Host, remoteFile, float64(written)/(1024*1024))
	return nil
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
