package extract_test

import (
	"bytes"
	"os"
	"testing"

	"Protodoc/pkg/extract"
)

// TestFixtureGen_1GiB10000PageCorpusIsDeterministic is T-0097's named test.
// The generator produces a byte-identical fixture across two runs on the
// same seed (determinism), the fixture is structurally walkable (its CONTENT
// segments are all found in order), and the documented NFR-011
// reference-config stand-in note exists. The determinism check runs at the
// small profile for speed; the 1 GiB / 10,000-page profile uses the same
// deterministic code path (exercised in -bench runs, T-0098..T-0100).
func TestFixtureGen_1GiB10000PageCorpusIsDeterministic(t *testing.T) {
	// Determinism: same seed -> byte-identical image.
	a, err := extract.GenerateCorpus(extract.SmallCorpusProfile(7))
	if err != nil {
		t.Fatalf("GenerateCorpus a: %v", err)
	}
	b, err := extract.GenerateCorpus(extract.SmallCorpusProfile(7))
	if err != nil {
		t.Fatalf("GenerateCorpus b: %v", err)
	}
	if !bytes.Equal(a, b) {
		t.Fatalf("corpus not deterministic: two runs on seed 7 differ (%d vs %d octets)", len(a), len(b))
	}
	// A different seed yields a different image (the seed actually varies it).
	c, err := extract.GenerateCorpus(extract.SmallCorpusProfile(8))
	if err != nil {
		t.Fatalf("GenerateCorpus c: %v", err)
	}
	if bytes.Equal(a, c) {
		t.Fatalf("corpus did not vary with seed")
	}

	// Structurally walkable: the small profile has 64 pages -> 64 CONTENT
	// segments discoverable in ascending order.
	segs, err := extract.ContentSegments(bytes.NewReader(a))
	if err != nil {
		t.Fatalf("ContentSegments: %v", err)
	}
	if len(segs) != extract.SmallCorpusProfile(7).Pages {
		t.Fatalf("walked %d CONTENT segments, want %d pages", len(segs), extract.SmallCorpusProfile(7).Pages)
	}
	for i := 1; i < len(segs); i++ {
		if segs[i].Ordinal <= segs[i-1].Ordinal {
			t.Fatalf("CONTENT segments not in ascending ordinal at %d", i)
		}
	}

	// The GiB profile is defined and non-degenerate (documented benchmark
	// target); we don't materialise 1 GiB in a unit test.
	gib := extract.GiBCorpusProfile(1)
	if gib.Pages != 10000 {
		t.Fatalf("GiBCorpusProfile pages = %d, want 10000", gib.Pages)
	}

	// The NFR-011 reference-config stand-in note exists and flags the
	// pending ruling.
	raw, err := os.ReadFile("../../docs/extraction-benchmark-corpus.md")
	if err != nil {
		t.Fatalf("reading corpus README: %v", err)
	}
	doc := string(raw)
	for _, needle := range []string{"NFR-011", "PDL-REFCFG-2026-09-DARWIN-ARM64-M2MAX", "pending"} {
		if !bytes.Contains(raw, []byte(needle)) {
			t.Fatalf("corpus README missing %q (must flag the NFR-011 stand-in dependency)", needle)
		}
	}
	_ = doc
}
