package canon

import (
	"bytes"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// degenerateVerdict is the single defined outcome for a degenerate fixture:
// whether canonicalization succeeded and how many canonical content octets it
// produced. Exactly one such verdict exists per fixture (no crash, no ambiguity).
type degenerateVerdict struct {
	ok           bool
	contentBytes int
}

func canonDegenerate(state *Document) degenerateVerdict {
	var buf bytes.Buffer
	err := Canonicalize(state, &buf)
	return degenerateVerdict{ok: err == nil, contentBytes: buf.Len()}
}

// TestCONF_DEGENERATE_001_EmptyContentCaseCorpus is T-0316's named conformance
// test (vector CONF-DEGENERATE-001; FR-124). At least 6 degenerate/empty-content
// fixtures each yield exactly one defined verdict, identical across independent
// runs, with no crash or undefined case.
func TestCONF_DEGENERATE_001_EmptyContentCaseCorpus(t *testing.T) {
	u := func(b byte) pdlfmt.UnitID {
		var id pdlfmt.UnitID
		id[0] = b
		return id
	}

	fixtures := []struct {
		name  string
		state *Document
		want  degenerateVerdict
	}{
		{"empty-document", EmptyState(), degenerateVerdict{ok: true, contentBytes: 0}},
		{"nil-document", nil, degenerateVerdict{ok: true, contentBytes: 0}},
		{"single-empty-frame", &Document{Subtrees: []ContentSubtree{{UnitID: u(0x1), Frame: []byte{}}}}, degenerateVerdict{ok: true, contentBytes: 0}},
		{"zero-run-empty-text-unit", &Document{Subtrees: []ContentSubtree{{UnitID: u(0x2), Frame: []byte{0x01}}}}, degenerateVerdict{ok: true, contentBytes: 1}},
		{"empty-table-zero-rows-cols", &Document{Subtrees: []ContentSubtree{{UnitID: u(0x3), Frame: []byte{0x03}}}}, degenerateVerdict{ok: true, contentBytes: 1}},
		{"zero-page-pagination-decl", &Document{Subtrees: []ContentSubtree{{UnitID: u(0x4), Frame: []byte{0x0C}}}}, degenerateVerdict{ok: true, contentBytes: 1}},
		{"all-empty-frames", &Document{Subtrees: []ContentSubtree{{UnitID: u(0x5), Frame: []byte{}}, {UnitID: u(0x6), Frame: []byte{}}}}, degenerateVerdict{ok: true, contentBytes: 0}},
	}
	if len(fixtures) < 6 {
		t.Fatalf("need >=6 degenerate fixtures, have %d", len(fixtures))
	}

	for _, f := range fixtures {
		t.Run(f.name, func(t *testing.T) {
			// Exactly one defined verdict, and it matches the golden reference.
			v1 := canonDegenerate(f.state)
			if v1 != f.want {
				t.Errorf("%s: verdict = %+v, want golden %+v", f.name, v1, f.want)
			}
			// Identical across independent runs (determinism).
			v2 := canonDegenerate(f.state)
			if v1 != v2 {
				t.Errorf("%s: verdict not deterministic across runs (%+v vs %+v)", f.name, v1, v2)
			}
		})
	}
}
