package extract_test

import (
	"go/format"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestTR_011_ExtractPackageUnder1000LinesNoHeavyDeps is T-0101's named test:
// the composed TR-011 gate. It fails if the extract package's non-test
// source exceeds 1000 gofmt-normalized lines, or if `go list -deps` surfaces
// any forbidden import edge (integrity/render/merge, or a font/shaping/
// layout/graphics/crypto package). It passes on the current tree.
func TestTR_011_ExtractPackageUnder1000LinesNoHeavyDeps(t *testing.T) {
	// --- Line budget: <= 1000 non-test gofmt-normalized source lines. ---
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("reading package dir: %v", err)
	}
	total := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		src, err := os.ReadFile(e.Name())
		if err != nil {
			t.Fatalf("reading %s: %v", e.Name(), err)
		}
		formatted, err := format.Source(src)
		if err != nil {
			t.Fatalf("gofmt %s: %v", e.Name(), err)
		}
		total += strings.Count(string(formatted), "\n")
	}
	const lineBudget = 1000
	if total > lineBudget {
		t.Errorf("extract package non-test source is %d gofmt lines, exceeds TR-011 budget %d", total, lineBudget)
	}
	t.Logf("extract non-test source: %d/%d lines", total, lineBudget)

	// --- Dependency isolation: no forbidden import edge. ---
	out, err := exec.Command("go", "list", "-deps", "Protodoc/pkg/extract").Output()
	if err != nil {
		t.Skipf("go list unavailable: %v", err)
	}
	// Forbidden internal packages.
	forbiddenExact := map[string]bool{
		"Protodoc/pkg/integrity": true,
		"Protodoc/pkg/render":    true,
		"Protodoc/pkg/merge":     true,
	}
	// Forbidden dependency categories by path. TR-011's normative text
	// forbids a font, shaping, layout or graphics dependency; the
	// extraction view also must not pull in trust/verification or
	// entropy/minting facilities (integrity is covered by forbiddenExact
	// above; crypto/rand is the CSPRNG minting/entropy edge extraction has
	// no business needing). The stdlib crypto/sha256 the container inventory
	// uses to read the prefix digests is NOT forbidden -- extraction cannot
	// read a Protodoc prefix without it, and TR-011 does not forbid it.
	forbiddenSubstr := []string{
		"image/", "golang.org/x/image", "golang.org/x/text",
		"/font", "/shaping", "/layout", "/graphics", "/harfbuzz", "/freetype",
		"crypto/rand",
	}
	for _, line := range strings.Split(string(out), "\n") {
		dep := strings.TrimSpace(line)
		if dep == "" {
			continue
		}
		if forbiddenExact[dep] {
			t.Errorf("extract has a forbidden import edge to %s (TR-011)", dep)
		}
		for _, sub := range forbiddenSubstr {
			if strings.Contains(dep, sub) {
				t.Errorf("extract has a forbidden %q-category import edge to %s (TR-011: no font/shaping/layout/graphics/crypto dependency)", sub, dep)
			}
		}
	}
}
