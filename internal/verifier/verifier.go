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

	h, algoName, err := hashForAlgorithm(algorithm)
	if err != nil {
		return err
	}

	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file for checksum: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("compute checksum: %w", err)
	}

	actualHash := hex.EncodeToString(h.Sum(nil))
	if actualHash != expectedHash {
		return fmt.Errorf("checksum mismatch (%s): expected %s, got %s", algoName, expectedHash, actualHash)
	}

	return nil
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
