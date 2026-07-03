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

package glance

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVersionNotEmpty(t *testing.T) {
	if Version == "" {
		t.Fatal("expected Version constant to be non-empty")
	}
}

func TestVersionHasVPrefix(t *testing.T) {
	if !strings.HasPrefix(Version, "v") {
		t.Fatalf("expected Version to start with 'v', got %q", Version)
	}
}

func TestLicenseTextNotEmpty(t *testing.T) {
	if LicenseText == "" {
		t.Fatal("expected LicenseText to be non-empty")
	}
}

func TestLicenseTextContainsMIT(t *testing.T) {
	if !strings.Contains(LicenseText, "MIT License") {
		t.Fatal("expected LicenseText to contain 'MIT License'")
	}
}

func TestHelpTextNotEmpty(t *testing.T) {
	if HelpText() == "" {
		t.Fatal("expected HelpText() to return non-empty string")
	}
}

func TestParseVersionFlag(t *testing.T) {
	cfg, err := Parse([]string{"--version"})
	if err != nil {
		t.Fatalf("unexpected error parsing --version: %v", err)
	}
	if !cfg.ShowVersion {
		t.Fatal("expected ShowVersion to be true")
	}
}

func TestParseLicenseFlag(t *testing.T) {
	cfg, err := Parse([]string{"--license"})
	if err != nil {
		t.Fatalf("unexpected error parsing --license: %v", err)
	}
	if !cfg.ShowLicense {
		t.Fatal("expected ShowLicense to be true")
	}
}

func TestParseHelpFlag(t *testing.T) {
	cfg, err := Parse([]string{"--help"})
	if err != nil {
		t.Fatalf("unexpected error parsing --help: %v", err)
	}
	if !cfg.ShowHelp {
		t.Fatal("expected ShowHelp to be true")
	}
}

func TestExecuteVersionReturnsNil(t *testing.T) {
	if err := Execute([]string{"--version"}); err != nil {
		t.Fatalf("Execute --version returned error: %v", err)
	}
}

func TestExecuteLicenseReturnsNil(t *testing.T) {
	if err := Execute([]string{"--license"}); err != nil {
		t.Fatalf("Execute --license returned error: %v", err)
	}
}

func TestExecuteHelpReturnsNil(t *testing.T) {
	if err := Execute([]string{"--help"}); err != nil {
		t.Fatalf("Execute --help returned error: %v", err)
	}
}

func TestCalculateFileHashSHA256(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(f, []byte("hello glance\n"), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	hash, algo, err := CalculateFileHash(f, "sha256")
	if err != nil {
		t.Fatalf("CalculateFileHash returned error: %v", err)
	}
	if algo != "sha256" {
		t.Fatalf("expected algo 'sha256', got %q", algo)
	}
	if len(hash) != 64 {
		t.Fatalf("expected 64-char hex sha256, got %d chars: %q", len(hash), hash)
	}
}

func TestCalculateFileHashSHA512(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(f, []byte("hello glance\n"), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	hash, algo, err := CalculateFileHash(f, "sha512")
	if err != nil {
		t.Fatalf("CalculateFileHash returned error: %v", err)
	}
	if algo != "sha512" {
		t.Fatalf("expected algo 'sha512', got %q", algo)
	}
	if len(hash) != 128 {
		t.Fatalf("expected 128-char hex sha512, got %d chars: %q", len(hash), hash)
	}
}

func TestVerifyFileHashCorrect(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "test.txt")
	content := []byte("verify me\n")
	if err := os.WriteFile(f, content, 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	hash, _, err := CalculateFileHash(f, "sha256")
	if err != nil {
		t.Fatalf("CalculateFileHash: %v", err)
	}

	if err := VerifyFileHash(f, hash, "sha256"); err != nil {
		t.Fatalf("VerifyFileHash with correct hash returned error: %v", err)
	}
}

func TestVerifyFileHashWrong(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(f, []byte("verify me\n"), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	wrongHash := strings.Repeat("0", 64)
	if err := VerifyFileHash(f, wrongHash, "sha256"); err == nil {
		t.Fatal("expected error for wrong hash, got nil")
	}
}

func TestGenerateKeyPairEd25519(t *testing.T) {
	dir := t.TempDir()
	priv, pub, err := GenerateKeyPair(KeygenConfig{Algorithm: "ed25519", OutputDir: dir})
	if err != nil {
		t.Fatalf("GenerateKeyPair returned error: %v", err)
	}
	for _, p := range []string{priv, pub} {
		info, err := os.Stat(p)
		if err != nil {
			t.Fatalf("key file missing %q: %v", p, err)
		}
		if info.Size() == 0 {
			t.Fatalf("key file empty: %q", p)
		}
	}
}
