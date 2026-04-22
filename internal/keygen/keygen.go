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
// File purpose: Generates SSH key pairs (ed25519, RSA, ECDSA) and writes private/public key files with proper permissions.

package keygen

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
)

type Config struct {
	Algorithm string
	OutputDir string
	KeyName   string
	BitSize   int
}

func GenerateKeyPair(cfg Config) (privateKeyPath, publicKeyPath string, err error) {
	algo := strings.ToLower(strings.TrimSpace(cfg.Algorithm))
	if algo == "" {
		algo = "ed25519"
	}

	if cfg.OutputDir == "" {
		cfg.OutputDir = "."
	}
	if err := os.MkdirAll(cfg.OutputDir, 0o700); err != nil {
		return "", "", fmt.Errorf("create output dir: %w", err)
	}

	keyName := cfg.KeyName
	if keyName == "" {
		keyName = "id_" + algo
	}

	privPath := filepath.Join(cfg.OutputDir, keyName)
	pubPath := privPath + ".pub"

	var privPEM []byte
	var pubKey ssh.PublicKey

	switch algo {
	case "ed25519":
		privPEM, pubKey, err = generateEd25519()
	case "rsa":
		bits := cfg.BitSize
		if bits == 0 {
			bits = 4096
		}
		privPEM, pubKey, err = generateRSA(bits)
	case "ecdsa":
		privPEM, pubKey, err = generateECDSA()
	default:
		return "", "", fmt.Errorf("unsupported algorithm: %s (supported: ed25519, rsa, ecdsa)", algo)
	}
	if err != nil {
		return "", "", err
	}

	if err := os.WriteFile(privPath, privPEM, 0o600); err != nil {
		return "", "", fmt.Errorf("write private key: %w", err)
	}

	pubBytes := ssh.MarshalAuthorizedKey(pubKey)
	if err := os.WriteFile(pubPath, pubBytes, 0o644); err != nil {
		return "", "", fmt.Errorf("write public key: %w", err)
	}

	return privPath, pubPath, nil
}

func generateEd25519() ([]byte, ssh.PublicKey, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generate ed25519 key: %w", err)
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal ed25519 key: %w", err)
	}

	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privBytes,
	})

	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		return nil, nil, fmt.Errorf("create ssh public key: %w", err)
	}

	return privPEM, sshPub, nil
}

func generateRSA(bits int) ([]byte, ssh.PublicKey, error) {
	priv, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, fmt.Errorf("generate rsa key: %w", err)
	}

	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})

	sshPub, err := ssh.NewPublicKey(&priv.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("create ssh public key: %w", err)
	}

	return privPEM, sshPub, nil
}

func generateECDSA() ([]byte, ssh.PublicKey, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generate ecdsa key: %w", err)
	}

	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal ecdsa key: %w", err)
	}

	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privBytes,
	})

	sshPub, err := ssh.NewPublicKey(&priv.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("create ssh public key: %w", err)
	}

	return privPEM, sshPub, nil
}
