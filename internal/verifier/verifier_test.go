package verifier

import "testing"

func TestHashForAlgorithmRejectsMD5(t *testing.T) {
	if _, _, err := hashForAlgorithm("md5"); err == nil {
		t.Fatal("expected md5 to be rejected")
	}
}

func TestHashForAlgorithmSupportsSHA(t *testing.T) {
	for _, algo := range []string{"", "sha256", "sha512"} {
		_, name, err := hashForAlgorithm(algo)
		if err != nil {
			t.Fatalf("expected %q to be accepted, got error: %v", algo, err)
		}
		if name != "sha256" && name != "sha512" {
			t.Fatalf("unexpected algorithm name %q for input %q", name, algo)
		}
	}
}
