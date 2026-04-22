package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func DownloadISO(source, outputDir string) (string, error) {
	source = strings.TrimSpace(source)
	if source == "" {
		return "", fmt.Errorf("iso source cannot be empty")
	}

	if outputDir == "" {
		outputDir = "."
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", fmt.Errorf("create output dir: %w", err)
	}

	filename := filepath.Base(source)
	if filename == "." || filename == "/" || filename == "" {
		filename = "image.iso"
	}
	outPath := filepath.Join(outputDir, filename)

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
	resp, err := client.Get(source)
	if err != nil {
		return "", fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status: %s", resp.Status)
	}

	outFile, err := os.Create(outPath)
	if err != nil {
		return "", fmt.Errorf("create output file: %w", err)
	}
	defer outFile.Close()

	written, err := transferWithProgress(resp.Body, outFile, resp.ContentLength, true)
	if err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}

	fmt.Printf("Downloaded %s (%.2f MB)\n", outPath, float64(written)/(1024*1024))
	fmt.Printf("Completed at %s\n", time.Now().Format(time.RFC3339))
	return outPath, nil
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

	written, err := transferWithProgress(src, dst, total, false)
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

func transferWithProgress(src io.Reader, dst io.Writer, total int64, networkTransfer bool) (int64, error) {
	buf := make([]byte, 64*1024)
	start := time.Now()
	lastPrint := start
	var written int64
	var lastBytes int64

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
