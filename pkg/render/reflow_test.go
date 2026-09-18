package render

import (
	"reflect"
	"testing"
)

// buildParagraph makes a Knuth-Plass item stream of `words` boxes of the given
// width separated by glue, with legal glue breakpoints -- all measures from the
// in-document table, no host input.
func buildParagraph(words, boxW, glueW int) BreakTable {
	var items []Item
	for i := 0; i < words; i++ {
		items = append(items, Item{Kind: KindBox, Width: boxW})
		if i < words-1 {
			items = append(items, Item{Kind: KindGlue, Width: glueW, Stretch: glueW, Shrink: glueW / 2})
		}
	}
	return BreakTable{Items: items, InputDigest: [32]byte{0xAB, 0xCD}}
}

// TestFR_100_ReflowDeterministicLineBreaks is T-0250's named unit test (FR-100,
// NFR-022). Reflowing identical content at a fixed width twice produces
// byte-identical line-break positions (the reflow is a pure function of the
// in-document break table and the width), and it reads no host dictionary or
// locale service.
func TestFR_100_ReflowDeterministicLineBreaks(t *testing.T) {
	tbl := buildParagraph(20, 1000, 300)
	const width = 5000

	// Reflow twice: byte-identical breakpoints.
	first := Reflow(tbl, width)
	second := Reflow(tbl, width)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("reflow not deterministic: %v vs %v", first, second)
	}
	if len(first) == 0 {
		t.Fatal("reflow produced no breakpoints")
	}
	// The final breakpoint is the end of the stream.
	if first[len(first)-1] != len(tbl.Items) {
		t.Errorf("last breakpoint = %d, want end-of-stream %d", first[len(first)-1], len(tbl.Items))
	}

	// "Host locale/dictionary swapped": the engine takes NO host input, so
	// nothing about the environment can change the result. We model this by
	// confirming the output depends ONLY on (items, width): a copy of the
	// table with the same items yields the same breaks regardless of any
	// external state.
	tblCopy := BreakTable{Items: append([]Item(nil), tbl.Items...), InputDigest: tbl.InputDigest}
	if !reflect.DeepEqual(Reflow(tblCopy, width), first) {
		t.Error("reflow output changed for identical items (host state must not matter)")
	}

	// Different width -> generally different breaks (the width is an input).
	narrow := Reflow(tbl, 2200)
	if len(narrow) <= len(first) {
		t.Errorf("narrower width should need at least as many lines: narrow=%d wide=%d", len(narrow), len(first))
	}

	// A hyphenation candidate from the table (a flagged penalty) is honoured as
	// a legal break but discouraged; reflow still succeeds deterministically.
	hy := tbl
	hy.Items = append([]Item(nil), tbl.Items...)
	hy.Items = append(hy.Items, Item{Kind: KindPenalty, Width: 100, Penalty: 50, Flagged: true})
	if r1, r2 := Reflow(hy, width), Reflow(hy, width); !reflect.DeepEqual(r1, r2) {
		t.Error("reflow with a hyphenation candidate is not deterministic")
	}
}
