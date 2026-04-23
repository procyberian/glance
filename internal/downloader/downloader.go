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
// File purpose: Downloads ISO files from HTTP/HTTPS or copies local ISOs with live progress, throughput, and ETA reporting.

package downloader

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"
)

var checksumLinePattern = regexp.MustCompile(`(?i)\b([a-f0-9]{32}|[a-f0-9]{64}|[a-f0-9]{128})\b`)
var hrefISORegex = regexp.MustCompile(`(?i)href\s*=\s*["']([^"']+\.iso(?:\?[^"']*)?)["']`)
var hrefAnyRegex = regexp.MustCompile(`(?i)href\s*=\s*["']([^"']+)["']`)

type FTPISOOption struct {
	Name     string
	URL      string
	Size     int64
	Checksum string
}

type ftpISOEntry struct {
	Path string
	Size int64
}

func DownloadISO(source, outputDir, outputPath string, allowResume bool) (string, error) {
	source = strings.TrimSpace(source)
	if source == "" {
		return "", fmt.Errorf("iso source cannot be empty")
	}

	if outputDir == "" {
		outputDir = "."
	}

	filename := filepath.Base(source)
	if filename == "." || filename == "/" || filename == "" {
		filename = "image.iso"
	}

	outPath := strings.TrimSpace(outputPath)
	if outPath == "" {
		outPath = filepath.Join(outputDir, filename)
	} else if strings.HasSuffix(outPath, string(os.PathSeparator)) {
		outPath = filepath.Join(outPath, filename)
	} else if stat, err := os.Stat(outPath); err == nil && stat.IsDir() {
		outPath = filepath.Join(outPath, filename)
	}

	parentDir := filepath.Dir(outPath)
	if parentDir == "" {
		parentDir = "."
	}
	if err := os.MkdirAll(parentDir, 0o755); err != nil {
		return "", fmt.Errorf("create output dir: %w", err)
	}

	tempPath := outPath + ".download"

	if !isHTTPSource(source) {
		return copyLocalISO(source, outPath)
	}

	source = upgradeToHTTPS(source)

	client := &http.Client{
		Timeout: 0,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if req.URL.Scheme == "http" {
				req.URL.Scheme = "https"
			}
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	var existingSize int64
	if !allowResume {
		if _, statErr := os.Stat(tempPath); statErr == nil {
			if err := os.Remove(tempPath); err != nil {
				return "", fmt.Errorf("remove old resume file: %w", err)
			}
			fmt.Printf("Resume disabled, restarting from zero: %s\n", tempPath)
		}
	}
	if allowResume {
		if stat, statErr := os.Stat(tempPath); statErr == nil {
			existingSize = stat.Size()
		}
	}

	req, err := http.NewRequest(http.MethodGet, source, nil)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	if existingSize > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", existingSize))
		fmt.Printf("Resuming download from %.2f MB: %s\n", float64(existingSize)/(1024*1024), tempPath)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return "", fmt.Errorf("download failed with status: %s", resp.Status)
	}

	appendMode := existingSize > 0 && resp.StatusCode == http.StatusPartialContent
	if existingSize > 0 && resp.StatusCode == http.StatusOK {
		// Server ignored Range request. Restart from zero to avoid corruption.
		existingSize = 0
		appendMode = false
		fmt.Println("Server does not support resume, restarting full download")
	}

	if resp.StatusCode == http.StatusRequestedRangeNotSatisfiable {
		return "", fmt.Errorf("resume range rejected by server")
	}

	flags := os.O_CREATE | os.O_WRONLY
	if appendMode {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}

	outFile, err := os.OpenFile(tempPath, flags, 0o644)
	if err != nil {
		return "", fmt.Errorf("create output file: %w", err)
	}
	defer outFile.Close()

	total := resp.ContentLength
	if appendMode && resp.ContentLength > 0 {
		total = existingSize + resp.ContentLength
	}

	written, err := transferWithProgress(resp.Body, outFile, total, true, existingSize)
	if err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}

	if err := os.Rename(tempPath, outPath); err != nil {
		return "", fmt.Errorf("finalize download: %w", err)
	}

	fmt.Printf("Downloaded %s (%.2f MB)\n", outPath, float64(written)/(1024*1024))
	fmt.Printf("Completed at %s\n", time.Now().Format(time.RFC3339))
	return outPath, nil
}

func ResolveChecksum(source, algorithm string) (string, error) {
	source = strings.TrimSpace(source)
	if !isHTTPSource(source) {
		return "", fmt.Errorf("automatic checksum discovery is only supported for HTTP/HTTPS ISO sources")
	}

	source = upgradeToHTTPS(source)
	checksumURLCandidates, err := checksumCandidates(source, algorithm)
	if err != nil {
		return "", err
	}

	isoName := pathBaseFromURL(source)
	var lastErr error
	for _, candidate := range checksumURLCandidates {
		checksum, err := fetchChecksumCandidate(candidate, isoName)
		if err == nil {
			return checksum, nil
		}
		lastErr = err
	}

	if lastErr != nil {
		return "", fmt.Errorf("automatic checksum discovery failed: %w", lastErr)
	}

	return "", fmt.Errorf("automatic checksum discovery failed")
}

func copyLocalISO(sourcePath, outPath string) (string, error) {
	src, err := os.Open(sourcePath)
	if err != nil {
		return "", fmt.Errorf("open source iso: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(outPath)
	if err != nil {
		return "", fmt.Errorf("create output file: %w", err)
	}
	defer dst.Close()

	var total int64 = -1
	if stat, statErr := src.Stat(); statErr == nil {
		total = stat.Size()
	}

	written, err := transferWithProgress(src, dst, total, false, 0)
	if err != nil {
		return "", fmt.Errorf("copy iso: %w", err)
	}

	fmt.Printf("Copied %s to %s (%.2f MB)\n", sourcePath, outPath, float64(written)/(1024*1024))
	fmt.Printf("Completed at %s\n", time.Now().Format(time.RFC3339))
	return outPath, nil
}

func isHTTPSource(source string) bool {
	source = strings.ToLower(strings.TrimSpace(source))
	return strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://")
}

// upgradeToHTTPS replaces http:// with https:// to ensure encrypted transfers.
func upgradeToHTTPS(source string) string {
	if strings.HasPrefix(strings.ToLower(source), "http://") {
		upgraded := "https://" + source[len("http://"):]
		fmt.Printf("Warning: HTTP URL upgraded to HTTPS: %s\n", upgraded)
		return upgraded
	}
	return source
}

func transferWithProgress(src io.Reader, dst io.Writer, total int64, networkTransfer bool, initialWritten int64) (int64, error) {
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
			avgBps := float64(written) / elapsedSeconds
			printProgressLine(written, total, instantBps, avgBps, networkTransfer)

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

func checksumCandidates(source, algorithm string) ([]string, error) {
	parsedURL, err := url.Parse(source)
	if err != nil {
		return nil, fmt.Errorf("parse iso url: %w", err)
	}

	fileName := pathBaseFromURL(source)
	if fileName == "" || fileName == "." || fileName == "/" {
		fileName = "image.iso"
	}

	dirURL := *parsedURL
	dirURL.RawQuery = ""
	dirURL.Fragment = ""
	dirURL.Path = strings.TrimSuffix(parsedURL.Path, fileName)

	algoName := strings.ToLower(strings.TrimSpace(algorithm))
	if algoName == "" {
		algoName = "sha256"
	}

	suffix := map[string]string{
		"sha256": ".sha256sum",
		"sha512": ".sha512sum",
		"md5":    ".md5sum",
	}[algoName]
	if suffix == "" {
		return nil, fmt.Errorf("unsupported checksum algorithm: %s", algorithm)
	}

	indexName := map[string]string{
		"sha256": "SHA256SUMS",
		"sha512": "SHA512SUMS",
		"md5":    "MD5SUMS",
	}[algoName]

	return []string{
		source + suffix,
		dirURL.String() + indexName,
		dirURL.String() + "checksum",
	}, nil
}

func fetchChecksumCandidate(candidateURL, isoName string) (string, error) {
	resp, err := http.Get(candidateURL)
	if err != nil {
		return "", fmt.Errorf("fetch %s: %w", candidateURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch %s: unexpected status %s", candidateURL, resp.Status)
	}

	checksum, err := extractChecksum(resp.Body, isoName)
	if err != nil {
		return "", fmt.Errorf("parse %s: %w", candidateURL, err)
	}

	return checksum, nil
}

func extractChecksum(r io.Reader, isoName string) (string, error) {
	scanner := bufio.NewScanner(r)
	var fallback string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		checksum := checksumLinePattern.FindString(line)
		if checksum == "" {
			continue
		}

		if fallback == "" {
			fallback = strings.ToLower(checksum)
		}

		if isoName == "" {
			return strings.ToLower(checksum), nil
		}

		normalizedLine := strings.ToLower(strings.ReplaceAll(line, "\\", "/"))
		normalizedISOName := strings.ToLower(isoName)
		if strings.Contains(normalizedLine, normalizedISOName) {
			return strings.ToLower(checksum), nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	if fallback != "" {
		return fallback, nil
	}

	return "", fmt.Errorf("no checksum entry found")
}

func pathBaseFromURL(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return filepath.Base(rawURL)
	}
	return filepath.Base(parsedURL.Path)
}

func printProgressLine(written, total int64, instantBps, avgBps float64, networkTransfer bool) {
	kind := "Disk hizi"
	if networkTransfer {
		kind = "Ag hizi"
	}

	if total > 0 {
		remainingBytes := float64(total - written)
		eta := "bilinmiyor"
		if avgBps > 0 && remainingBytes > 0 {
			etaDuration := time.Duration((remainingBytes / avgBps) * float64(time.Second))
			eta = etaDuration.Truncate(time.Second).String()
		} else if remainingBytes <= 0 {
			eta = "0s"
		}

		percent := (float64(written) / float64(total)) * 100
		fmt.Printf(
			"\r%.2f%% | %s/%s | Anlik: %s/s | Ortalama: %s/s | %s: %s/s | ETA: %s",
			percent,
			formatBytes(float64(written)),
			formatBytes(float64(total)),
			formatBytes(instantBps),
			formatBytes(avgBps),
			kind,
			formatBytes(instantBps),
			eta,
		)
		return
	}

	fmt.Printf(
		"\r%s indirildi | Anlik: %s/s | Ortalama: %s/s | %s: %s/s | ETA: bilinmiyor",
		formatBytes(float64(written)),
		formatBytes(instantBps),
		formatBytes(avgBps),
		kind,
		formatBytes(instantBps),
	)
}

func formatBytes(v float64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	idx := 0
	for v >= 1024 && idx < len(units)-1 {
		v /= 1024
		idx++
	}
	return fmt.Sprintf("%.2f %s", v, units[idx])
}

func FormatSize(size int64) string {
	if size <= 0 {
		return "unknown"
	}
	return formatBytes(float64(size))
}

func ListFTPISOs(source, algorithm string) ([]FTPISOOption, error) {
	if !isFTPSource(source) {
		return nil, fmt.Errorf("source must start with ftp://")
	}

	algo, err := normalizeChecksumAlgorithm(algorithm)
	if err != nil {
		return nil, err
	}

	conn, ftpInfo, err := dialFTP(source)
	if err != nil {
		return nil, err
	}
	defer conn.Quit()

	fmt.Println("Detecting total ISO count...")
	entries, err := walkFTPISOEntries(conn, ftpInfo.BasePath)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Found %d total ISOs\n", len(entries))

	if len(entries) == 0 {
		return nil, fmt.Errorf("no ISO files found on FTP source")
	}

	options := make([]FTPISOOption, 0, len(entries))
	for i, entry := range entries {
		if i > 0 {
			// Keep checksum-resolution requests paced so FTP servers are not overloaded.
			time.Sleep(50 * time.Millisecond)
		}

		checksum, _ := resolveFTPChecksumForPath(conn, entry.Path, algo)
		printChecksumProgressWithTotal(i+1, len(entries))

		_, name := splitFTPDirAndFile(entry.Path)
		options = append(options, FTPISOOption{
			Name:     name,
			URL:      "ftp://" + ftpInfo.HostPort + entry.Path,
			Size:     entry.Size,
			Checksum: checksum,
		})
	}

	sort.Slice(options, func(i, j int) bool {
		return strings.ToLower(options[i].Name) < strings.ToLower(options[j].Name)
	})

	return options, nil
}

func ListHTTPISOs(source, algorithm string) ([]FTPISOOption, error) {
	if !isHTTPSource(source) {
		return nil, fmt.Errorf("source must start with http:// or https://")
	}

	algo, err := normalizeChecksumAlgorithm(algorithm)
	if err != nil {
		return nil, err
	}

	baseURL := upgradeToHTTPS(strings.TrimSpace(source))
	if !strings.HasSuffix(baseURL, "/") {
		return nil, fmt.Errorf("http/https ISO listing source must be a directory URL ending with '/'")
	}

	client := &http.Client{Timeout: 20 * time.Second}
	parsedBase, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse source URL: %w", err)
	}

	rootHost := strings.ToLower(parsedBase.Hostname())
	if rootHost == "" {
		return nil, fmt.Errorf("source host is empty")
	}

	fmt.Println("Detecting total ISO count...")

	queue := buildHTTPSeedDirs(parsedBase)
	visitedDirs := make(map[string]bool)
	isoSeen := make(map[string]bool)
	isoURLs := make([]string, 0, 64)

	const maxDirs = 6000
	scannedDirs := 0

	for len(queue) > 0 {
		dirURL := queue[0]
		queue = queue[1:]

		if visitedDirs[dirURL] {
			continue
		}
		visitedDirs[dirURL] = true

		resp, getErr := client.Get(dirURL)
		if getErr != nil {
			continue
		}

		bodyBytes, readErr := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
		resp.Body.Close()
		if readErr != nil {
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 400 {
			continue
		}

		scannedDirs++
		if scannedDirs >= maxDirs {
			break
		}

		baseDirParsed, parseErr := url.Parse(dirURL)
		if parseErr != nil {
			continue
		}

		matches := hrefAnyRegex.FindAllStringSubmatch(string(bodyBytes), -1)
		for _, m := range matches {
			if len(m) < 2 {
				continue
			}

			rawHref := strings.TrimSpace(m[1])
			if rawHref == "" {
				continue
			}

			ref, hrefErr := url.Parse(rawHref)
			if hrefErr != nil {
				continue
			}

			resolved := baseDirParsed.ResolveReference(ref)
			if resolved == nil {
				continue
			}

			if !strings.EqualFold(resolved.Scheme, "http") && !strings.EqualFold(resolved.Scheme, "https") {
				continue
			}
			if strings.ToLower(resolved.Hostname()) != rootHost {
				continue
			}

			resolved.Fragment = ""
			resolvedQueryless := *resolved
			resolvedQueryless.RawQuery = ""
			resolvedQueryless.Fragment = ""

			if isISOFilePath(resolvedQueryless.Path) {
				resolvedQueryless.Scheme = "https"
				isoURL := resolvedQueryless.String()
				if !isoSeen[isoURL] {
					isoSeen[isoURL] = true
					isoURLs = append(isoURLs, isoURL)
				}
				continue
			}

			if isLikelyDirectoryHref(rawHref, resolvedQueryless.Path) {
				nextDir := canonicalHTTPDirURL(&resolvedQueryless)
				if nextDir != "" && !visitedDirs[nextDir] {
					queue = append(queue, nextDir)
				}
			}
		}
	}

	fmt.Printf("Found %d total ISOs\n", len(isoURLs))

	if len(isoURLs) == 0 {
		return nil, fmt.Errorf("no ISO links found on HTTP/HTTPS source")
	}

	sort.Slice(isoURLs, func(i, j int) bool {
		return strings.ToLower(pathBaseFromURL(isoURLs[i])) < strings.ToLower(pathBaseFromURL(isoURLs[j]))
	})

	options := make([]FTPISOOption, 0, len(isoURLs))
	for i, isoURL := range isoURLs {
		if i > 0 {
			time.Sleep(50 * time.Millisecond)
		}

		size := int64(0)
		if req, reqErr := http.NewRequest(http.MethodHead, isoURL, nil); reqErr == nil {
			if headResp, headErr := client.Do(req); headErr == nil {
				if headResp.StatusCode >= 200 && headResp.StatusCode < 400 {
					size = headResp.ContentLength
				}
				headResp.Body.Close()
			}
		}

		checksum, _ := ResolveChecksum(isoURL, algo)
		printChecksumProgressWithTotal(i+1, len(isoURLs))

		options = append(options, FTPISOOption{
			Name:     pathBaseFromURL(isoURL),
			URL:      isoURL,
			Size:     size,
			Checksum: checksum,
		})
	}

	return options, nil
}

func buildHTTPSeedDirs(start *url.URL) []string {
	if start == nil {
		return nil
	}

	seedMap := make(map[string]bool)
	seeds := make([]string, 0, 8)

	current := *start
	for {
		dir := canonicalHTTPDirURL(&current)
		if dir == "" {
			break
		}
		if !seedMap[dir] {
			seedMap[dir] = true
			seeds = append(seeds, dir)
		}

		if current.Path == "/" || current.Path == "" {
			break
		}

		trimmed := strings.TrimRight(current.Path, "/")
		idx := strings.LastIndex(trimmed, "/")
		if idx <= 0 {
			current.Path = "/"
		} else {
			current.Path = trimmed[:idx+1]
		}
		current.RawQuery = ""
		current.Fragment = ""
	}

	return seeds
}

func canonicalHTTPDirURL(u *url.URL) string {
	if u == nil {
		return ""
	}

	clone := *u
	clone.RawQuery = ""
	clone.Fragment = ""
	if strings.TrimSpace(clone.Path) == "" {
		clone.Path = "/"
	}
	if !strings.HasSuffix(clone.Path, "/") {
		clone.Path += "/"
	}
	return clone.String()
}

func isLikelyDirectoryHref(rawHref, resolvedPath string) bool {
	raw := strings.TrimSpace(rawHref)
	if raw == "" {
		return false
	}
	if strings.HasPrefix(strings.ToLower(raw), "mailto:") || strings.HasPrefix(strings.ToLower(raw), "javascript:") {
		return false
	}
	if strings.HasSuffix(raw, "/") {
		return true
	}

	pathValue := strings.TrimSpace(resolvedPath)
	if pathValue == "" {
		return false
	}
	if strings.HasSuffix(pathValue, "/") {
		return true
	}

	base := filepath.Base(pathValue)
	if base == "." || base == "/" || base == "" {
		return true
	}
	if strings.Contains(base, ".") {
		return false
	}
	return true
}

func isISOFilePath(pathValue string) bool {
	base := strings.ToLower(strings.TrimSpace(filepath.Base(pathValue)))
	return strings.HasSuffix(base, ".iso")
}

func walkFTPISOEntries(conn *ftp.ServerConn, startPath string) ([]ftpISOEntry, error) {
	start := strings.TrimSpace(startPath)
	if start == "" {
		start = "/"
	}

	entries := make([]ftpISOEntry, 0, 32)
	queue := []string{start}
	visited := map[string]bool{}
	firstListCall := true

	for len(queue) > 0 {
		dir := queue[0]
		queue = queue[1:]

		if visited[dir] {
			continue
		}
		visited[dir] = true

		if !firstListCall {
			// Rate-limit directory traversal calls to avoid hammering FTP servers.
			time.Sleep(100 * time.Millisecond)
		}
		firstListCall = false

		list, err := conn.List(dir)
		if err != nil {
			return nil, fmt.Errorf("list ftp directory %s: %w", dir, err)
		}

		for _, item := range list {
			if item == nil {
				continue
			}

			name := strings.TrimSpace(item.Name)
			if name == "" || name == "." || name == ".." {
				continue
			}

			fullPath := joinFTPPath(dir, name)
			switch item.Type {
			case ftp.EntryTypeFile:
				if strings.HasSuffix(strings.ToLower(name), ".iso") {
					entries = append(entries, ftpISOEntry{Path: fullPath, Size: int64(item.Size)})
				}
			case ftp.EntryTypeFolder:
				queue = append(queue, fullPath)
			}
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return strings.ToLower(entries[i].Path) < strings.ToLower(entries[j].Path)
	})

	return entries, nil
}

func printChecksumProgressWithTotal(current, total int) {
	if total <= 0 {
		return
	}
	if current < 0 {
		current = 0
	}
	if current > total {
		current = total
	}

	percent := (float64(current) / float64(total)) * 100
	bar := progressBar(current, total)
	fmt.Printf("\r%s %.2f%% | %d/%d ISOs", bar, percent, current, total)
	if current == total {
		fmt.Println()
	}
}

type ftpConnectionInfo struct {
	Address  string
	HostPort string
	Username string
	Password string
	BasePath string
}

func isFTPSource(source string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(source)), "ftp://")
}

func parseFTPSource(source string) (ftpConnectionInfo, error) {
	parsed, err := url.Parse(strings.TrimSpace(source))
	if err != nil {
		return ftpConnectionInfo{}, fmt.Errorf("parse ftp source: %w", err)
	}
	if !strings.EqualFold(parsed.Scheme, "ftp") {
		return ftpConnectionInfo{}, fmt.Errorf("source must start with ftp://")
	}
	if parsed.Hostname() == "" {
		return ftpConnectionInfo{}, fmt.Errorf("ftp source host is empty")
	}

	port := parsed.Port()
	if port == "" {
		port = "21"
	}

	username := "anonymous"
	password := "anonymous@"
	if parsed.User != nil {
		if user := parsed.User.Username(); strings.TrimSpace(user) != "" {
			username = user
		}
		if pass, ok := parsed.User.Password(); ok {
			password = pass
		}
	}

	basePath := parsed.Path
	if strings.TrimSpace(basePath) == "" {
		basePath = "/"
	}
	if !strings.HasPrefix(basePath, "/") {
		basePath = "/" + basePath
	}

	hostPort := net.JoinHostPort(parsed.Hostname(), port)
	return ftpConnectionInfo{
		Address:  hostPort,
		HostPort: hostPort,
		Username: username,
		Password: password,
		BasePath: basePath,
	}, nil
}

func dialFTP(source string) (*ftp.ServerConn, ftpConnectionInfo, error) {
	info, err := parseFTPSource(source)
	if err != nil {
		return nil, ftpConnectionInfo{}, err
	}

	conn, err := ftp.Dial(info.Address, ftp.DialWithTimeout(12*time.Second))
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil, ftpConnectionInfo{}, fmt.Errorf("ftp connection timed out: %w", err)
		}
		return nil, ftpConnectionInfo{}, fmt.Errorf("connect ftp server: %w", err)
	}

	if err := conn.Login(info.Username, info.Password); err != nil {
		conn.Quit()
		return nil, ftpConnectionInfo{}, fmt.Errorf("ftp login failed: %w", err)
	}

	return conn, info, nil
}

func resolveFTPChecksumForPath(conn *ftp.ServerConn, isoPath, algorithm string) (string, error) {
	dir, file := splitFTPDirAndFile(isoPath)
	if file == "" {
		return "", fmt.Errorf("invalid ftp iso path: %s", isoPath)
	}

	suffix := map[string]string{
		"sha256": ".sha256sum",
		"sha512": ".sha512sum",
		"md5":    ".md5sum",
	}[algorithm]
	indexName := map[string]string{
		"sha256": "SHA256SUMS",
		"sha512": "SHA512SUMS",
		"md5":    "MD5SUMS",
	}[algorithm]

	candidates := []string{
		joinFTPPath(dir, file+suffix),
		joinFTPPath(dir, indexName),
		joinFTPPath(dir, strings.ToLower(indexName)),
		joinFTPPath(dir, "checksums"),
		joinFTPPath(dir, "checksum"),
		joinFTPPath(dir, "CHECKSUMS"),
	}

	var lastErr error
	seen := map[string]bool{}
	for _, candidate := range candidates {
		if candidate == "" || seen[candidate] {
			continue
		}
		seen[candidate] = true

		checksum, err := fetchFTPChecksumCandidate(conn, candidate, file)
		if err == nil {
			return checksum, nil
		}
		lastErr = err
	}

	if lastErr != nil {
		return "", lastErr
	}
	return "", fmt.Errorf("checksum not found")
}

func fetchFTPChecksumCandidate(conn *ftp.ServerConn, candidatePath, isoFile string) (string, error) {
	r, err := conn.Retr(candidatePath)
	if err != nil {
		return "", fmt.Errorf("open ftp checksum file %s: %w", candidatePath, err)
	}
	defer r.Close()

	checksum, err := extractChecksum(r, isoFile)
	if err != nil {
		return "", fmt.Errorf("parse ftp checksum file %s: %w", candidatePath, err)
	}
	return checksum, nil
}

func splitFTPDirAndFile(path string) (string, string) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return "/", ""
	}
	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}
	idx := strings.LastIndex(trimmed, "/")
	if idx <= 0 {
		return "/", strings.TrimPrefix(trimmed, "/")
	}
	return trimmed[:idx], trimmed[idx+1:]
}

func joinFTPPath(parts ...string) string {
	joined := ""
	for _, part := range parts {
		p := strings.TrimSpace(part)
		if p == "" {
			continue
		}
		if joined == "" {
			joined = p
			continue
		}
		joined = strings.TrimRight(joined, "/") + "/" + strings.TrimLeft(p, "/")
	}
	if joined == "" {
		return "/"
	}
	if !strings.HasPrefix(joined, "/") {
		joined = "/" + joined
	}
	return joined
}

func progressBar(current, total int) string {
	const width = 20
	if total <= 0 {
		return "[--------------------]"
	}
	if current < 0 {
		current = 0
	}
	if current > total {
		current = total
	}

	filled := int((float64(current) / float64(total)) * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	return "[" + strings.Repeat("=", filled) + strings.Repeat("-", width-filled) + "]"
}

func normalizeChecksumAlgorithm(algorithm string) (string, error) {
	algo := strings.ToLower(strings.TrimSpace(algorithm))
	if algo == "" {
		algo = "sha256"
	}
	switch algo {
	case "sha256", "sha512", "md5":
		return algo, nil
	default:
		return "", fmt.Errorf("unsupported checksum algorithm: %s", algorithm)
	}
}
