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

package keygen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateKeyPairEd25519(t *testing.T) {
	dir := t.TempDir()
	priv, pub, err := GenerateKeyPair(Config{Algorithm: "ed25519", OutputDir: dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	checkKeyFiles(t, dir, priv, pub, "id_ed25519")
	assertPubKeyType(t, pub, "ssh-ed25519")
}

func TestGenerateKeyPairRSA(t *testing.T) {
	dir := t.TempDir()
	priv, pub, err := GenerateKeyPair(Config{Algorithm: "rsa", OutputDir: dir, BitSize: 2048})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	checkKeyFiles(t, dir, priv, pub, "id_rsa")
	assertPubKeyType(t, pub, "ssh-rsa")
}

func TestGenerateKeyPairECDSA(t *testing.T) {
	dir := t.TempDir()
	priv, pub, err := GenerateKeyPair(Config{Algorithm: "ecdsa", OutputDir: dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	checkKeyFiles(t, dir, priv, pub, "id_ecdsa")
	assertPubKeyType(t, pub, "ecdsa-sha2-nistp256")
}

func TestGenerateKeyPairDefaultsToEd25519(t *testing.T) {
	dir := t.TempDir()
	priv, pub, err := GenerateKeyPair(Config{OutputDir: dir})
	if err != nil {
		t.Fatalf("unexpected error with empty algorithm: %v", err)
	}
	checkKeyFiles(t, dir, priv, pub, "id_ed25519")
}

func TestGenerateKeyPairCustomKeyName(t *testing.T) {
	dir := t.TempDir()
	priv, pub, err := GenerateKeyPair(Config{Algorithm: "ed25519", OutputDir: dir, KeyName: "deploy_key"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	checkKeyFiles(t, dir, priv, pub, "deploy_key")
}

func TestGenerateKeyPairUnsupportedAlgorithm(t *testing.T) {
	dir := t.TempDir()
	_, _, err := GenerateKeyPair(Config{Algorithm: "dsa", OutputDir: dir})
	if err == nil {
		t.Fatal("expected error for unsupported algorithm, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported algorithm") {
		t.Fatalf("expected 'unsupported algorithm' in error, got: %v", err)
	}
}

func TestGenerateKeyPairPrivateKeyPermissions(t *testing.T) {
	dir := t.TempDir()
	priv, _, err := GenerateKeyPair(Config{Algorithm: "ed25519", OutputDir: dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	info, err := os.Stat(priv)
	if err != nil {
		t.Fatalf("stat private key: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("expected private key permissions 0600, got %04o", perm)
	}
}

// checkKeyFiles verifies both key files exist, have non-zero size, and use the expected base name.
func checkKeyFiles(t *testing.T, dir, privPath, pubPath, wantBase string) {
	t.Helper()

	if filepath.Base(privPath) != wantBase {
		t.Errorf("expected private key name %q, got %q", wantBase, filepath.Base(privPath))
	}
	if filepath.Base(pubPath) != wantBase+".pub" {
		t.Errorf("expected public key name %q, got %q", wantBase+".pub", filepath.Base(pubPath))
	}

	for _, p := range []string{privPath, pubPath} {
		info, err := os.Stat(p)
		if err != nil {
			t.Fatalf("key file not found %q: %v", p, err)
		}
		if info.Size() == 0 {
			t.Fatalf("key file is empty: %q", p)
		}
	}

	_ = dir
}

// assertPubKeyType checks that the public key file starts with the given algorithm token.
func assertPubKeyType(t *testing.T, pubPath, wantType string) {
	t.Helper()
	data, err := os.ReadFile(pubPath)
	if err != nil {
		t.Fatalf("read public key: %v", err)
	}
	if !strings.HasPrefix(strings.TrimSpace(string(data)), wantType) {
		t.Errorf("expected public key to start with %q, got: %q", wantType, string(data))
	}
}
