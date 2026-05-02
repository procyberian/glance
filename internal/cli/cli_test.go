package cli

import "testing"

func TestParseRejectsMD5ChecksumAlgo(t *testing.T) {
	_, err := Parse([]string{"--download", "--iso", "https://example.com/test.iso", "--checksum-algo", "md5"})
	if err == nil {
		t.Fatal("expected Parse to reject md5 checksum algorithm")
	}
}

func TestParseRejectsFTPWithoutOptIn(t *testing.T) {
	_, err := Parse([]string{"--download", "--iso", "ftp://ftp.example.com/test.iso"})
	if err == nil {
		t.Fatal("expected Parse to reject ftp source without --allow-insecure-ftp")
	}
}

func TestParseAcceptsFTPWithOptIn(t *testing.T) {
	cfg, err := Parse([]string{"--download", "--allow-insecure-ftp", "--iso", "ftp://ftp.example.com/test.iso"})
	if err != nil {
		t.Fatalf("expected Parse to accept ftp source with --allow-insecure-ftp, got error: %v", err)
	}
	if !cfg.AllowInsecureFTP {
		t.Fatal("expected AllowInsecureFTP to be true")
	}
}
