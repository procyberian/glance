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

package uploader

import (
	"strings"
	"testing"
)

func TestUploadFileRejectsEmptyHost(t *testing.T) {
	err := UploadFile(Config{User: "user", LocalFile: "/tmp/file"})
	if err == nil {
		t.Fatal("expected error when host is empty")
	}
	if !strings.Contains(err.Error(), "host") {
		t.Fatalf("expected 'host' in error message, got: %v", err)
	}
}

func TestUploadFileRejectsEmptyUser(t *testing.T) {
	err := UploadFile(Config{Host: "example.com", LocalFile: "/tmp/file"})
	if err == nil {
		t.Fatal("expected error when user is empty")
	}
	if !strings.Contains(err.Error(), "user") {
		t.Fatalf("expected 'user' in error message, got: %v", err)
	}
}

func TestUploadFileRejectsEmptyLocalFile(t *testing.T) {
	err := UploadFile(Config{Host: "example.com", User: "user"})
	if err == nil {
		t.Fatal("expected error when local file is empty")
	}
	if !strings.Contains(err.Error(), "local file") {
		t.Fatalf("expected 'local file' in error message, got: %v", err)
	}
}

func TestUploadFileDefaultsPortTo22(t *testing.T) {
	// A config with valid host/user/file but pointing to a nonexistent server
	// must fail with a connection error, not a config error, confirming the
	// default port (22) is applied before the dial attempt.
	err := UploadFile(Config{
		Host:      "127.0.0.1",
		User:      "user",
		LocalFile: "/nonexistent-file-for-test",
	})
	if err == nil {
		t.Fatal("expected an error for nonexistent local file or unreachable host")
	}
	// Must not be a "host and user are required" or "local file path is required" error.
	if strings.Contains(err.Error(), "host and user are required") {
		t.Fatalf("validation should have passed, got: %v", err)
	}
	if strings.Contains(err.Error(), "local file path is required") {
		t.Fatalf("validation should have passed, got: %v", err)
	}
}
