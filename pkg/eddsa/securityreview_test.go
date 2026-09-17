package eddsa

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

const securityReviewDocPath = "../../docs/eddsa-protodoc-1-security-review.md"

// TestCON_015_ArgusSecurityReviewChecklist is T-0112's named test: the
// mechanical half of the argus security review, enforced continuously. It
// (1) confirms the review checklist doc exists and records zero open
// HIGH/CRITICAL findings, and (2) statically greps the package's own source
// for two prohibited patterns -- any math/rand usage anywhere in the package
// (non-crypto RNG), and any fmt/log formatting of private key material --
// failing if either is found.
func TestCON_015_ArgusSecurityReviewChecklist(t *testing.T) {
	// --- Checklist doc exists and is clean. ---
	raw, err := os.ReadFile(securityReviewDocPath)
	if err != nil {
		t.Fatalf("reading security review %s: %v", securityReviewDocPath, err)
	}
	doc := string(raw)
	if !strings.Contains(doc, "Zero open HIGH or CRITICAL findings") {
		t.Fatalf("security review does not record zero open HIGH/CRITICAL findings")
	}

	// --- Static grep gate over the package's own .go source (incl. tests,
	// since a test using math/rand for signing input would also be a smell,
	// but we scope the key-formatting check to all files too). ---
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("reading eddsa package dir: %v", err)
	}

	// math/rand import is prohibited anywhere in the package (non-crypto RNG).
	mathRandRe := regexp.MustCompile(`"math/rand"`)
	// Formatting a private key / seed via fmt or log is prohibited: match a
	// format/print call whose arguments name private key material.
	keyFormatRe := regexp.MustCompile(`(fmt\.(Errorf|Sprintf|Printf|Println|Sprint)|log\.[A-Za-z]+)\([^)]*\b(priv|privKey|privateKey|PrivateKey|seed|Seed)\b`)

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		// The gate's own file necessarily contains the pattern strings it
		// searches for; exclude it to avoid a self-match false positive.
		if e.Name() == "securityreview_test.go" {
			continue
		}
		src, err := os.ReadFile(e.Name())
		if err != nil {
			t.Fatalf("reading %s: %v", e.Name(), err)
		}
		text := string(src)

		if mathRandRe.MatchString(text) {
			t.Errorf("%s imports math/rand: the eddsa package must use no non-crypto RNG (T-0112)", e.Name())
		}
		if loc := keyFormatRe.FindString(text); loc != "" {
			t.Errorf("%s formats private key material in a fmt/log call (%q): key material must never appear in a log or error string (T-0112)", e.Name(), loc)
		}
	}
}
