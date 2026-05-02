package downloader

import "testing"

func TestNormalizeChecksumAlgorithmSupportsStrongHashes(t *testing.T) {
	cases := []string{"", "sha256", "sha512", "SHA512"}
	for _, in := range cases {
		got, err := normalizeChecksumAlgorithm(in)
		if err != nil {
			t.Fatalf("expected %q to be accepted, got error: %v", in, err)
		}
		if got != "sha256" && got != "sha512" {
			t.Fatalf("unexpected normalized algorithm for %q: %q", in, got)
		}
	}
}

func TestNormalizeChecksumAlgorithmRejectsMD5(t *testing.T) {
	if _, err := normalizeChecksumAlgorithm("md5"); err == nil {
		t.Fatal("expected md5 to be rejected")
	}
}
