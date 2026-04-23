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
// File purpose: Verifies file checksums using sha256, sha512, or md5 against expected values.

package verifier

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"strings"
)

func VerifyFileHash(filePath, expectedHash, algorithm string) error {
	expectedHash = normalizeHash(expectedHash)
	if expectedHash == "" {
		return fmt.Errorf("expected checksum cannot be empty")
	}

	actualHash, _, err := CalculateFileHash(filePath, algorithm)
	if err != nil {
		return err
	}

	if actualHash != expectedHash {
		return fmt.Errorf("checksum verification failed: expected %s, got %s", expectedHash, actualHash)
	}

	return nil
}

func CalculateFileHash(filePath, algorithm string) (string, string, error) {

	h, algoName, err := hashForAlgorithm(algorithm)
	if err != nil {
		return "", "", err
	}

	f, err := os.Open(filePath)
	if err != nil {
		return "", "", fmt.Errorf("open file for checksum: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(h, f); err != nil {
		return "", "", fmt.Errorf("compute checksum: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), algoName, nil
}

func hashForAlgorithm(algorithm string) (hash.Hash, string, error) {
	switch strings.ToLower(strings.TrimSpace(algorithm)) {
	case "", "sha256":
		return sha256.New(), "sha256", nil
	case "sha512":
		return sha512.New(), "sha512", nil
	case "md5":
		return md5.New(), "md5", nil
	default:
		return nil, "", fmt.Errorf("unsupported checksum algorithm: %s (supported: sha256, sha512, md5)", algorithm)
	}
}

func normalizeHash(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(strings.ToLower(v), "sha256:")
	v = strings.TrimPrefix(v, "sha512:")
	v = strings.TrimPrefix(v, "md5:")
	return v
}
