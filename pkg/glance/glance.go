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
// File purpose: Exposes a reusable public API for download, verification, upload, key generation, and CLI execution.

package glance

import (
	"fmt"
	"strings"
	"time"

	"github.com/procyberian/glance/v11/internal/cli"
	"github.com/procyberian/glance/v11/internal/downloader"
	"github.com/procyberian/glance/v11/internal/keygen"
	licensecontent "github.com/procyberian/glance/v11/internal/license"
	"github.com/procyberian/glance/v11/internal/uploader"
	"github.com/procyberian/glance/v11/internal/verifier"
)

type Config = cli.Config

type ISOOption = downloader.FTPISOOption

type UploadConfig = uploader.Config

type KeygenConfig = keygen.Config

type DownloadOptions struct {
	Source            string
	OutputDir         string
	OutputPath        string
	ExpectedChecksum  string
	ChecksumAlgorithm string
	AllowResume       bool
}

type DownloadResult struct {
	Path      string
	Checksum  string
	Algorithm string
}

const LicenseText = licensecontent.Text

func Parse(args []string) (Config, error) {
	return cli.Parse(args)
}

func Run(cfg Config) error {
	return cli.Run(cfg)
}

func HelpText() string {
	return cli.HelpText()
}

func DownloadISO(source, outputDir, outputPath string, allowResume bool) (string, error) {
	return downloader.DownloadISO(source, outputDir, outputPath, allowResume)
}

func DownloadAndVerify(options DownloadOptions) (DownloadResult, error) {
	algorithm := strings.TrimSpace(options.ChecksumAlgorithm)
	if algorithm == "" {
		algorithm = "sha256"
	}

	downloadedPath, err := downloader.DownloadISO(options.Source, options.OutputDir, options.OutputPath, options.AllowResume)
	if err != nil {
		return DownloadResult{}, err
	}

	checksum := strings.TrimSpace(options.ExpectedChecksum)
	if checksum == "" {
		if resolvedChecksum, resolveErr := downloader.ResolveChecksum(options.Source, algorithm); resolveErr == nil {
			checksum = resolvedChecksum
		}
	}

	if checksum == "" {
		calculatedChecksum, algoName, calcErr := verifier.CalculateFileHash(downloadedPath, algorithm)
		if calcErr != nil {
			return DownloadResult{}, calcErr
		}
		return DownloadResult{
			Path:      downloadedPath,
			Checksum:  calculatedChecksum,
			Algorithm: algoName,
		}, nil
	}

	if err := verifier.VerifyFileHash(downloadedPath, checksum, algorithm); err != nil {
		return DownloadResult{}, err
	}

	return DownloadResult{
		Path:      downloadedPath,
		Checksum:  strings.ToLower(strings.TrimSpace(checksum)),
		Algorithm: strings.ToLower(algorithm),
	}, nil
}

func ResolveChecksum(source, algorithm string) (string, error) {
	return downloader.ResolveChecksum(source, algorithm)
}

func ListFTPISOs(source, algorithm string) ([]ISOOption, error) {
	return downloader.ListFTPISOs(source, algorithm)
}

func ListFTPISOsWithTimeout(source, algorithm string, timeout time.Duration) ([]ISOOption, error) {
	return downloader.ListFTPISOsWithTimeout(source, algorithm, timeout)
}

func ListHTTPISOs(source, algorithm string) ([]ISOOption, error) {
	return downloader.ListHTTPISOs(source, algorithm)
}

func ListHTTPISOsWithTimeout(source, algorithm string, timeout time.Duration) ([]ISOOption, error) {
	return downloader.ListHTTPISOsWithTimeout(source, algorithm, timeout)
}

func UploadFile(cfg UploadConfig) error {
	return uploader.UploadFile(cfg)
}

func VerifyFileHash(filePath, expectedHash, algorithm string) error {
	return verifier.VerifyFileHash(filePath, expectedHash, algorithm)
}

func CalculateFileHash(filePath, algorithm string) (string, string, error) {
	return verifier.CalculateFileHash(filePath, algorithm)
}

func GenerateKeyPair(cfg KeygenConfig) (string, string, error) {
	return keygen.GenerateKeyPair(cfg)
}

func Execute(args []string) error {
	cfg, err := Parse(args)
	if err != nil {
		return err
	}

	if cfg.ShowLicense {
		fmt.Println(LicenseText)
		return nil
	}

	if cfg.ShowHelp {
		fmt.Print(HelpText())
		return nil
	}

	return Run(cfg)
}
