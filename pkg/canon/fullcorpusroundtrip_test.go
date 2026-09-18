package canon

import (
	"bytes"
	"math/rand"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestCONF_PROJECT_001_FullCorpusRoundTrip is T-0322's named conformance test
// (vector CONF-PROJECT-001; TR-004). It runs the projection round-trip check
// across an accumulated corpus of diverse fixtures (varied subtree counts,
// frame sizes, and byte values, plus the degenerate cases) and tracks the pass
// rate; the round-trip must hold for 100% of the corpus.
func TestCONF_PROJECT_001_FullCorpusRoundTrip(t *testing.T) {
	r := rand.New(rand.NewSource(0xC0FFEE))
	var corpus []*Document

	// Degenerate cases.
	corpus = append(corpus, &Document{}, EmptyState())

	// Diverse generated states.
	for f := 0; f < 40; f++ {
		n := 1 + r.Intn(8)
		subs := make([]ContentSubtree, n)
		for i := range subs {
			var id pdlfmt.UnitID
			id[0] = byte(r.Intn(256))
			id[1] = byte(f)
			id[15] = byte(i)
			frame := make([]byte, r.Intn(64))
			r.Read(frame)
			subs[i] = ContentSubtree{UnitID: id, Frame: frame}
		}
		corpus = append(corpus, &Document{Subtrees: subs})
	}

	passed := 0
	for idx, state := range corpus {
		var canonBuf bytes.Buffer
		if err := Canonicalize(state, &canonBuf); err != nil {
			t.Errorf("fixture %d: Canonicalize: %v", idx, err)
			continue
		}
		want := canonBuf.Bytes()
		textOK := bytes.Equal(recoverCanonicalFromText(t, ProjectText(state)), want)
		htmlOK := bytes.Equal(recoverCanonicalFromHTML(t, ProjectHTML(state)), want)
		if textOK && htmlOK {
			passed++
		} else {
			t.Errorf("fixture %d: round-trip failed (text=%v html=%v)", idx, textOK, htmlOK)
		}
	}

	if passed != len(corpus) {
		t.Errorf("round-trip pass rate %d/%d, want 100%%", passed, len(corpus))
	}
	t.Logf("CONF-PROJECT-001 round-trip pass rate: %d/%d", passed, len(corpus))
}
